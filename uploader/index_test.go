package uploader

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func BenchmarkSprintf(b *testing.B) {
	m := []byte("carbon.agents.carbon-clickhouse.graphite1.tcp.metricsReceived")
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%d:%s", i, unsafeString(m))
	}
}

func BenchmarkStringAppend(b *testing.B) {
	m := []byte("carbon.agents.carbon-clickhouse.graphite1.tcp.metricsReceived")
	b.ReportAllocs()
	b.ResetTimer()

	var out string
	for i := 0; i < b.N; i++ {
		out = strconv.FormatInt(int64(i), 10) + ":" + unsafeString(m)
	}
	b.StopTimer()
	b.ReportAllocs()
	z := fmt.Sprintf("%d:%s", b.N-1, unsafeString(m))
	if strings.Compare(string(z), out) != 0 {
		b.Fatalf("Unexpected string: %s, want: %s", out, z)
	}
}
