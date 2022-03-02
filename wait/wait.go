package wait

import "time"

var ForeverTestTimeout = time.Second * 30

// NeverStop may be passed to Until to make it never stop.
var NeverStop <-chan struct{} = make(chan struct{})

// Forever calls f every period for ever.
//
// Forever is syntactic sugar on top of Until.
func Forever(f func(), period time.Duration) {
	Until(f, period, NeverStop)
}

// Until loops until stop channel is closed, running f every period.
func Until(f func(), period time.Duration, stopCh <-chan struct{}) {
	JitterUntil(f, period, 0.0, true, stopCh)
}

// JitterUntil loops until stop channel is closed, running f every period.
//
// If jitterFactor is positive, the period is jittered before every run of f.
// If jitterFactor is not positive, the period is unchanged and not jittered.
//
// If sliding is true, true period is computed after f runs. If it is false then
// period includes the runtime for f.
//
// Close stopCh to stop. f may not be invoked if stop channel is already
// closed. Pass NeverStop to if you don't want it stop.
func JitterUntil(f func(), period time.Duration, jitterFactor float64, sliding bool, stopCh <-chan struct{}) {

}
