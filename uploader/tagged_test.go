package uploader

import (
	"net/url"
	"strings"
	"testing"

	"github.com/lomik/carbon-clickhouse/helper/escape"
	"github.com/msaf1980/stringutils"
	"github.com/stretchr/testify/assert"
)

func TestUrlParse(t *testing.T) {
	assert := assert.New(t)

	// make metric name as receiver
	metric := escape.Path("instance:cpu_utilization:ratio_avg") +
		"?" + escape.Query("dc") + "=" + escape.Query("qwe") +
		"&" + escape.Query("fqdn") + "=" + escape.Query("asd") +
		"&" + escape.Query("instance") + "=" + escape.Query("10.33.10.10:9100") +
		"&" + escape.Query("job") + "=" + escape.Query("node")

	assert.Equal("instance:cpu_utilization:ratio_avg?dc=qwe&fqdn=asd&instance=10.33.10.10%3A9100&job=node", metric)

	// original url.Parse
	m, err := url.Parse(metric)
	assert.NotNil(m)
	assert.NoError(err)
	assert.Equal("", m.Path)

	// from tagged uploader
	m, err = urlParse(metric)
	assert.NotNil(m)
	assert.NoError(err)
	assert.Equal("instance:cpu_utilization:ratio_avg", m.Path)
}

func BenchmarkStringsBuffer(b *testing.B) {
	var sb strings.Builder
	sb.Grow(1000000)
	sb.Reset()
	s := "asdfghjklqwertyuiopzxcvbnm1234567890"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if sb.Len()+len(s) > sb.Cap() {
			sb.Reset()
		}
		sb.WriteString(s)
	}
}

func BenchmarkStringBuffer(b *testing.B) {
	var sb stringutils.Builder
	sb.Grow(1000000)
	sb.Reset()
	s := "asdfghjklqwertyuiopzxcvbnm1234567890"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if sb.Len()+len(s) > sb.Cap() {
			sb.Reset()
		}
		sb.WriteString(s)
	}
}
