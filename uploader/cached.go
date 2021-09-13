package uploader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/lomik/carbon-clickhouse/helper/binary_store"

	"go.uber.org/zap"
)

type DebugCacheDumper interface {
	CacheDump(io.Writer)
}

type cached struct {
	*Base
	existsCache *CMap // store known keys and don't load it to clickhouse tree
	parser      func(filename string, out io.Writer) (*uploaderStat, map[string]bool, error)
	expired     uint32 // atomic counter

	hash     string
	storeDir string
	// db       *leveldb.DB
	peers []string
	// peersChan []chan string
}

func newCached(base *Base, dataDir string) *cached {
	u := &cached{Base: base}
	u.Base.handler = u.upload

	// distributed and persistent cache
	if len(u.config.Hash) > 0 {
		u.hash = u.config.Hash
	}

	u.existsCache = NewCMap(u.config.CachePersistQueue)
	u.query = fmt.Sprintf("%s (Date, Level, Path, Version)", u.config.TableName)

	if u.config.CachePersistQueue > 0 {
		var hashMode string
		if u.config.Hash == "" {
			hashMode = "origin"
		} else {
			hashMode = u.config.Hash
		}

		u.storeDir = path.Join(dataDir, u.config.TableName+".cache_"+hashMode)
		u.peers = u.config.CachePeers
	}

	return u
}

func (u *cached) Stat(send func(metric string, value float64)) {
	u.Base.Stat(send)

	send("cacheSize", float64(u.existsCache.Count()))

	expired := atomic.LoadUint32(&u.expired)
	atomic.AddUint32(&u.expired, -expired)
	send("expired", float64(expired))
}

func makeDirectory(path string) error {
	if stat, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModeDir|0755)
	} else if err != nil {
		return nil
	} else if !stat.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", path)
	}
	return nil
}

func (u *cached) Start() error {
	var err error

	if u.config.CachePersistQueue > 0 {
		if err = makeDirectory(u.storeDir); err != nil {
			return err
		}
		if err = u.loadCache(); err != nil {
			return err
		}
	}

	err = u.Base.Start()
	if err != nil {
		return err
	}

	if u.config.CachePersistQueue > 0 {
		u.Go(func(ctx context.Context) {
			u.cacheStoreWorker(ctx)
		})
	}

	if u.config.CacheTTL.Value() != 0 {
		u.Go(func(ctx context.Context) {
			u.existsCache.ExpireWorker(ctx, u.config.CacheTTL.Value(), &u.expired)
		})
	}

	return nil
}

func (u *cached) Reset() {
	u.existsCache.Clear()
	u.resetCache()
	debug.FreeOSMemory()
}

func (u *cached) upload(ctx context.Context, logger *zap.Logger, filename string) (*uploaderStat, error) {
	var stat *uploaderStat
	var err error
	var newSeries map[string]bool

	pipeReader, pipeWriter := io.Pipe()
	writer := bufio.NewWriter(pipeWriter)
	startTime := time.Now()

	uploadResult := make(chan error, 1)

	u.Go(func(ctx context.Context) {
		err = u.insertRowBinary(
			u.query,
			pipeReader,
		)
		uploadResult <- err
		if err != nil {
			pipeReader.CloseWithError(err)
		}
	})

	stat, newSeries, err = u.parser(filename, writer)
	if err == nil {
		err = writer.Flush()
	}
	pipeWriter.CloseWithError(err)

	var uploadErr error

	select {
	case uploadErr = <-uploadResult:
		// pass
	case <-ctx.Done():
		return stat, fmt.Errorf("upload aborted")
	}

	if err != nil {
		return stat, err
	}

	if uploadErr != nil {
		return stat, uploadErr
	}

	// commit new series
	u.existsCache.Merge(newSeries, startTime.Unix())

	return stat, nil
}

func (u *cached) resetCache() error {
	files, err := ioutil.ReadDir(u.storeDir)
	if err != nil {
		return err
	}
	count := 0
	for _, f := range files {
		_, err := strconv.ParseUint(f.Name(), 10, 64)
		if err != nil {
			continue
		}
		fileName := path.Join(u.storeDir, f.Name())
		os.Remove(fileName)
		count++
	}
	u.logger.Info("reset cache files", zap.Int("count", count))

	return nil
}

func (u *cached) loadCache() error {
	files, err := ioutil.ReadDir(u.storeDir)
	if err != nil {
		return err
	}

	expire := uint64(time.Now().Add(-u.config.CacheTTL.Value() - time.Minute).Unix())
	br := binary_store.NewReader()

	for _, f := range files {
		fileTimestamp, err := strconv.ParseUint(f.Name(), 10, 64)
		if err != nil {
			continue
		}
		fileName := path.Join(u.storeDir, f.Name())
		fileTimestamp /= 1000000000

		if fileTimestamp < expire {
			os.Remove(fileName)
			continue
		}
		if err := br.Open(fileName); err != nil {
			return err
		}

	READ_LOOP:
		for {
			if key, v, err := br.GetKeyValue(); err == nil {
				if v > expire {
					u.existsCache.Add(key, int64(v), false)
				}
			} else {
				break READ_LOOP
			}
		}

		br.Close()
	}

	return nil
}

func (u *cached) expireCache() error {
	files, err := ioutil.ReadDir(u.storeDir)
	if err != nil {
		return err
	}

	expire := uint64(time.Now().Add(-u.config.CacheTTL.Value() - time.Minute).Unix())
	count := 0

	for _, f := range files {
		fileTimestamp, err := strconv.ParseUint(f.Name(), 10, 64)
		if err != nil {
			continue
		}
		fileName := path.Join(u.storeDir, f.Name())
		fileTimestamp /= 1000000000
		if fileTimestamp < expire {
			os.Remove(fileName)
			count++
			continue
		}
	}
	u.logger.Info("expire cache files", zap.Int("count", count))

	return nil
}

func (u *cached) cacheStoreWorker(ctx context.Context) {
	var (
		err      error
		storeErr bool
		fileName string
	)

	binWritter := binary_store.NewWriter()
	prevTime := time.Now().Unix()
	interval := time.Second * 10

	for {
		select {
		case <-ctx.Done():
			return
		case key := <-u.existsCache.storeChan:
			ts := time.Now()
			now := ts.Unix()

			if !binWritter.IsOpen() {
				fileName = path.Join(u.storeDir, strconv.FormatInt(ts.UnixNano(), 10))
				if err := binWritter.Open(fileName); err != nil {
					if !storeErr {
						u.logger.Error("unable to open file for persistent cache", zap.Error(err))
						storeErr = true
					}
				}
			}
			if binWritter.PutKeyValue(key, uint64(now)) != nil {
				if !storeErr {
					u.logger.Error("unable put to persistent cache", zap.Error(err))
					storeErr = true
				}
				binWritter.BufReset() // buffer might be lost, if write to fail failed
			}
		case <-time.After(interval):
			ts := time.Now()
			now := ts.Unix()

			if binWritter.Size() == 0 {
				continue // don't need to rotate, file is empty
			}

			if !binWritter.IsOpen() {
				fileName = path.Join(u.storeDir, strconv.FormatInt(ts.UnixNano(), 10))
				if err := binWritter.Open(fileName); err != nil {
					if !storeErr {
						u.logger.Error("unable to open file for persistent cache", zap.Error(err))
						storeErr = true
					}
				}
			}

			if binWritter.IsOpen() {
				if binWritter.Len() > 0 {
					if err := binWritter.Flush(); err == nil {
						if storeErr {
							u.logger.Info("resume flush to persistent cache")
							storeErr = false
						}
					} else if !storeErr {
						u.logger.Error("unable to flush persistent cache", zap.Error(err))
						storeErr = true
					}
				}
				binWritter.Close()

				if now > prevTime+int64(u.config.CacheTTL.Value().Seconds()) {
					prevTime = now
					go u.expireCache()
				}
			}
		}
	}
}
