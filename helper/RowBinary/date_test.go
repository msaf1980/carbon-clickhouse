package RowBinary

import (
	"strconv"
	"testing"
	"time"
)

func TestSlowTimestampToDays(t *testing.T) {
	tests := []struct {
		ts      uint32
		want    uint16
		wantStr string
	}{
		{
			ts:      1668106870, // 2022-11-11 00:01:10 +05:00 ; 2022-11-10 19:01:10
			want:    19307,
			wantStr: "2022-11-11",
		},
		{
			ts:      1668124800, // 2022-11-11 00:00:00 UTC
			want:    19307,
			wantStr: "2022-11-11",
		},
		{
			ts:      1668142799, // 2022-11-10 23:59:59 -05:00; 2022-11-11 04:59:59 UTC
			want:    19307,
			wantStr: "2022-11-11",
		},
		{
			ts: 1650776160, // graphite-clickhouse issue #184, graphite-clickhouse in UTC, clickhouse in PDT(UTC-7)
			// 2022-04-24 4:56:00
			// select toDate(1650776160,'UTC')
			//                        2022-04-24
			// select toDate(1650776160,'Etc/GMT+7')
			//                        2022-04-23
			want:    19106,
			wantStr: "2022-04-24",
		},
	}
	dayStart, err := time.Parse("2006-01-02 MST", "1970-01-01 UTC") // see Clickhouse Date type, it's number of days from 1970-01-01
	if err != nil {
		t.Fatal(err)
	}
	// https://clickhouse.com/docs/en/sql-reference/data-types/date/
	for i, tt := range tests {
		t.Run(strconv.FormatInt(int64(tt.ts), 10)+" "+time.Unix(int64(tt.ts), 0).UTC().Format(time.RFC3339)+" ["+strconv.Itoa(i)+"]", func(t *testing.T) {
			got := SlowTimestampToDays(tt.ts)
			gotStr := dayStart.Add(time.Duration(int64(got) * 24 * 3600 * 1e9)).UTC().Format("2006-01-02")
			if gotStr != tt.wantStr || got != tt.want {
				t.Errorf("TimestampDaysFormat() = %v (%s), want %v (%s)", got, gotStr, tt.want, tt.wantStr)
			}
			convStr := dayStart.Add(time.Duration(int64(tt.want) * 24 * 3600 * 1e9)).UTC().Format("2006-01-02")
			if convStr != tt.wantStr {
				t.Errorf("conversion got %s, want %s", convStr, tt.wantStr)
			}
		})
	}
}

func TestTimestampToDays(t *testing.T) {
	ts := uint32(0)
	end := uint32(time.Now().Unix()) + 473040000 // +15 years
	for ts < end {
		d1 := SlowTimestampToDays(ts)
		d2 := TimestampToDays(ts)
		if d1 != d2 {
			t.FailNow()
		}

		ts += 780 // step 13 minutes
	}
}
func BenchmarkTimestampToDays(b *testing.B) {
	timestamp := uint32(time.Now().Unix())
	x := SlowTimestampToDays(timestamp)

	for i := 0; i < b.N; i++ {
		if TimestampToDays(timestamp) != x {
			b.FailNow()
		}
	}
}

func BenchmarkSlowTimestampToDays(b *testing.B) {
	timestamp := uint32(time.Now().Unix())
	x := SlowTimestampToDays(timestamp)

	for i := 0; i < b.N; i++ {
		if SlowTimestampToDays(timestamp) != x {
			b.FailNow()
		}
	}
}
