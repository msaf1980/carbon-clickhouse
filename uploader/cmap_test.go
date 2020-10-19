package uploader

import (
	"testing"
)

func Benchmark_fnv32(b *testing.B) {
	s := "asdfghjklqwe.rtyuiopzxc.vbnm1234567890"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fnv32(s)
	}
}
