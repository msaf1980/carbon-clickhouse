package uploader

import (
	"strconv"
	"testing"
)

func BenchmarkCMapMerge(b *testing.B) {
	// make metric name as receiver
	count := 1000000
	values := make([]string, count)
	for i := range values {
		values[i] = strconv.Itoa(i)
	}
	cmap := NewCMap()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m := make(map[string]bool)
			for i := 0; i < count; i++ {
				m[values[i]] = true
			}
			cmap.Merge(m, int64(1))
		}
	})
}

func BenchmarkCMapMergeSlice(b *testing.B) {
	// make metric name as receiver
	count := 1000000
	values := make([]string, count)
	for i := range values {
		values[i] = strconv.Itoa(i)
	}
	cmap := NewCMap()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m := make([]string, count)
			for i := 0; i < count; i++ {
				m[i] = values[i]
			}
			cmap.MergeSlice(m, int64(1))
		}
	})
}
