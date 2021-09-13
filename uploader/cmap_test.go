package uploader

import (
	"math/rand"
	"strconv"
	"testing"
	// "github.com/spaolacci/murmur3"
	// cityhash32 "github.com/tenfyzhong/cityhash"
)

const maxMapValsCount = 1024000

func BenchmarkCMap(b *testing.B) {
	m := NewCMap(0)

	values := make([]string, maxMapValsCount)
	for i := 0; i < maxMapValsCount; i++ {
		values[i] = strconv.FormatInt(rand.Int63(), 10)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := rand.Intn(maxMapValsCount)
			_ = m.Exists(values[i+1%maxMapValsCount-1])
			m.Add(values[i], int64(i), false)
		}
	})
}
