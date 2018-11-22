package progress

import (
	"bytes"
	"fmt"
	"time"
)

type Progress struct {
	start time.Time

	nowf func() time.Time

	window time.Duration // sampling window

	elem []elem // collected samples
	w, r int    // sample write/read index

	status Status
}

type elem struct {
	v int64
	t time.Duration
}

const resolution = time.Duration(100 * time.Millisecond)

func New(total int64, option ...Option) *Progress {
	p := &Progress{
		nowf:   time.Now,
		window: 5 * time.Second,
		status: Status{
			Total: total,
		},
	}

	for _, o := range option {
		o.update(p)
	}

	p.window = p.window.Truncate(resolution)
	if p.window <= 0 {
		p.window = resolution
	}

	p.elem = make([]elem, p.window/resolution+1)

	p.start = p.nowf()

	return p
}

func (p *Progress) Status() Status {
	return p.status
}

func (p *Progress) Update(v int64) {
	t := p.nowf().Sub(p.start).Truncate(resolution)
	p.add(t, v)
}

func (p *Progress) add(t time.Duration, v int64) {
	//log.Printf("add %s %d", t, v)
	p.status.V += v
	p.status.Done += v
	if t <= p.elem[p.w].t {
		p.elem[p.w].v += v
		return
	}
	p.incw()
	p.elem[p.w] = elem{t: t, v: v}
	st := t - p.window
	for p.r != p.w && p.elem[p.r].t < st {
		p.incr()
	}
	p.status.Dt = t - p.elem[p.r].t
	//log.Printf("%d/%s", p.status.V, p.status.Dt)
}

func (p *Progress) incw() {
	p.w = (p.w + 1) % len(p.elem)
	if p.w == p.r {
		p.incr()
	}
}

func (p *Progress) incr() {
	p.status.V -= p.elem[p.r].v
	p.r = (p.r + 1) % len(p.elem)
}

type Status struct {
	Done  int64
	Total int64

	// V/Dt is throughput
	V  int64
	Dt time.Duration
}

func (s Status) TimeRemaining() (time.Duration, bool) {
	if s.Total <= 0 || s.Dt <= 0 {
		return 0, false // unknown
	}

	left := float64(s.Total - s.Done)
	throughput := float64(s.V) / float64(s.Dt.Seconds())
	if throughput < 1 {
		return 0, false // inaccurate/unknown
	}
	t := time.Duration(left / throughput * float64(time.Second))
	return t.Truncate(time.Second), true
}

func (s Status) String() string {
	buf := new(bytes.Buffer)
	if s.Total > 0 {
		r := float64(s.Done) / float64(s.Total)
		fmt.Fprintf(buf, "%7.2f%% ", float64(r)*100)
	}

	var throughput float64
	if s.Dt > 0 {
		throughput = float64(s.V) / float64(s.Dt.Seconds())
	}

	tp := throughput
	sfxi := 0
	for tp >= 1000 && sfxi+1 < len(sbsuffix) {
		sfxi++
		tp /= 1000
	}

	switch {
	case tp <= 0:
		buf.WriteString("0")
	case tp < 10:
		fmt.Fprintf(buf, "%.1f%s", tp, sbsuffix[sfxi])
	default:
		fmt.Fprintf(buf, "%.0f%s", tp, sbsuffix[sfxi])
	}
	buf.WriteString("B/s")

	r, ok := s.TimeRemaining()
	if ok {
		fmt.Fprintf(buf, " ETA %s", r)
	}

	return buf.String()
}

var sbsuffix = []string{
	"",
	"ki",
	"Mi",
	"Gi",
	"Ti",
	"Ei",
}
