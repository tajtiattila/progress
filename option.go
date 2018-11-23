package progress

import "time"

// Option is an option for New.
type Option interface {
	update(*Progress)
}

type optf func(p *Progress)

func (f optf) update(p *Progress) {
	f(p)
}

// NowFunc uses the provided function
// instead of time.Now to get the current time.
//
// It is intended for testing and debugging.
func NowFunc(now func() time.Time) Option {
	return optf(func(p *Progress) {
		p.nowf = now
	})
}

// SampleWindow sets the sample window duration to d.
func SampleWindow(d time.Duration) Option {
	return optf(func(p *Progress) {
		p.window = d
	})
}

// StatusFunc calls f after every status change.
func StatusFunc(f func(Status)) Option {
	return optf(func(p *Progress) {
		p.statusCallback = f
	})
}
