package timers

import "math/rand"
import "time"

type Gaussian struct {
	C <-chan time.Time

	t        *time.Timer
	lastMean time.Duration
	lastDev  time.Duration
}

// creates Gaussian distributed Duration (only possitive)
func stdDuration(m time.Duration, d time.Duration) time.Duration {
	res := time.Duration(rand.NormFloat64()*float64(d) + float64(m))
	if res < 0 {
		return 0
	} else {
		return res
	}
}

// NewGaussian createos a timer (similar from standard library package time)
// The duration for timer to trigger isn't fully defined.
// It is a Gaussian distribution with mean and standard deviation.
func NewGaussian(mean time.Duration, stddev time.Duration) *Gaussian {
	t := time.NewTimer(stdDuration(mean, stddev))
	return &Gaussian{
		C:        t.C,
		t:        t,
		lastMean: mean,
		lastDev:  stddev,
	}
}

// Calls Stop method on underlying timer
func (g *Gaussian) Stop() bool {
	return g.t.Stop()
}

// Calls Reset on underlying timer with duration defined by Gaussian distribution
// defined in parameters to the functon.
func (g *Gaussian) Reset(mean time.Duration, stddev time.Duration) bool {
	g.lastMean = mean
	g.lastDev = stddev
	return g.t.Reset(stdDuration(mean, stddev))
}

// Calls Reset on underlying timer with duration defined by last Gaussian distribution.
func (g *Gaussian) ResetLast() bool {
	return g.t.Reset(stdDuration(g.lastMean, g.lastDev))
}
