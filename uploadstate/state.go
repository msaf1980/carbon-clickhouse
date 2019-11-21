package uploadstate

import (
	"sync/atomic"
)

type State struct {
	fail    int32
	maxFail int32
}

// uploader state
var state State

// Init with maxFailures
func Init(maxFailures int32) {
	state.maxFail = maxFailures
}

// Incr increment failure count and return true if maxFail count is reached
func Incr() bool {
	if atomic.LoadInt32(&state.fail) < state.maxFail {
		return atomic.AddInt32(&state.fail, 1) >= state.maxFail
	}
	return true
}

// Decr decrement failure count, return true if fail count reset to 0
func Decr() bool {
	if atomic.LoadInt32(&state.fail) > 0 {
		atomic.AddInt32(&state.fail, -1)
		return atomic.LoadInt32(&state.fail) <= 0
	}
	return false
}

// IsFail check if failure count reched maxFail
func IsFail() bool {
	return state.maxFail > 0 && atomic.LoadInt32(&state.fail) >= state.maxFail
}
