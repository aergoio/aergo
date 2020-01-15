package p2p

import (
	"testing"
	"time"
)

func Test_getNextInterval(t *testing.T) {
	prev := time.Second
	for i := 0; i < 20; i++ {
		got := getNextInterval(i)
		if got < time.Second {
			t.Fatalf("getNextInterval() too short duration %v ", got.String())
		}
		if got < prev {
			t.Fatalf("getNextInterval() next trial is shorter than prev next %v , prev %v", got, prev)
		}
		prev = got
	}
}
