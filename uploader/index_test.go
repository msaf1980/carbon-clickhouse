package uploader

import (
	"testing"
)

func Test_indexNextDay(t *testing.T) {
	indexes := []indexRow{
		{day: 1},
		{day: 1},
		{day: 3},
		{day: 5},
		{day: 5},
		{day: 5},
		{day: 7},
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
