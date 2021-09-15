package uploader

import (
	"testing"

	"github.com/lomik/carbon-clickhouse/helper/escape"
)

func Benchmark_fnv32(b *testing.B) {
	// make metric name as receiver
	metric := escape.Path("instance:cpu_utilization:ratio_avg") +
		"?" + escape.Query("dc") + "=" + escape.Query("qwe") +
		"&" + escape.Query("fqdn") + "=" + escape.Query("asd") +
		"&" + escape.Query("instance") + "=" + escape.Query("10.33.10.10_9100") +
		"&" + escape.Query("job") + "=" + escape.Query("node")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = fnv32(metric)
	}
}

func Benchmark_cityHash64(b *testing.B) {
	// make metric name as receiver
	metric := escape.Path("instance:cpu_utilization:ratio_avg") +
		"?" + escape.Query("dc") + "=" + escape.Query("qwe") +
		"&" + escape.Query("fqdn") + "=" + escape.Query("asd") +
		"&" + escape.Query("instance") + "=" + escape.Query("10.33.10.10_9100") +
		"&" + escape.Query("job") + "=" + escape.Query("node")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cityHash64(metric)
	}
}

func Benchmark_xxH3_64(b *testing.B) {
	// make metric name as receiver
	metric := escape.Path("instance:cpu_utilization:ratio_avg") +
		"?" + escape.Query("dc") + "=" + escape.Query("qwe") +
		"&" + escape.Query("fqdn") + "=" + escape.Query("asd") +
		"&" + escape.Query("instance") + "=" + escape.Query("10.33.10.10_9100") +
		"&" + escape.Query("job") + "=" + escape.Query("node")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = xxh3Hash64(metric)
	}
}

// func TestCollisionCityHash64(t *testing.T) {

// 	hashes := make(map[string]struct{})
// 	max := 50000000

// 	var b [40]byte
// 	for i := 0; i < max; i++ {

// 		_, err := rand.Read(b[:])
// 		if err != nil {
// 			t.Error(err)
// 			t.Fail()
// 		}

// 		result := cityHash64(unsafeString(b[:]))

// 		if _, ok := hashes[result]; ok {
// 			t.Logf("%s == %v", result, b)
// 			t.Errorf("Found collision after %d searches", i)
// 		}

// 		hashes[result] = struct{}{}
// 	}

// 	fmt.Println("no collisions found", max)
// }

// func TestCollisionXXH3_64(t *testing.T) {

// 	hashes := make(map[string]struct{})
// 	max := 50000000

// 	var b [40]byte
// 	for i := 0; i < max; i++ {

// 		_, err := rand.Read(b[:])
// 		if err != nil {
// 			t.Error(err)
// 			t.Fail()
// 		}

// 		result := xxh3Hash64(unsafeString(b[:]))

// 		if _, ok := hashes[result]; ok {
// 			t.Logf("%s == %v", result, b)
// 			t.Errorf("Found collision after %d searches", i)
// 		}

// 		hashes[result] = struct{}{}
// 	}

// 	fmt.Println("no collisions found", max)
// }
