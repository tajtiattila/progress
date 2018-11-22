package progress

import "time"

type Option interface {
	update(*Progress)
}

type optf func(p *Progress)

func (f optf) update(p *Progress) {
	f(p)
}

func NowFunc(now func() time.Time) Option {
	return optf(func(p *Progress) {
		p.nowf = now
	})
}

func SampleWindow(d time.Duration) Option {
	return optf(func(p *Progress) {
		p.window = d
	})
}
