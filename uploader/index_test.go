package uploader

import (
	"testing"
)

func Test_indexNextDay(t *testing.T) {
	indexes := []indexRow{
		{days: 1},
		{days: 1},
		{days: 3},
		{days: 5},
		{days: 5},
		{days: 5},
		{days: 7},
	}
	tests := []int{
		2, 3, 6, len(indexes),
	}
	pos := 0
	for n, tt := range tests {
		if got := indexNextDay(indexes, pos, 4); got != tt {
			t.Fatalf("step %d nextDay(%d) = %d, want %d", n, pos, got, tt)
		} else {
			pos = got
			if pos == len(indexes) && n != len(tests)-1 {
				t.Fatalf("step %d nextDay pos out of range", n)
			}
		}
	}
}
