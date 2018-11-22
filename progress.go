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

	sumv int64 // running sum of []elem.v
	done int64 // update sum

	favg weightValueAvg // average finished time calculator

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
		favg: weightValueAvg{
			v: make([]weightValue, 8),
		},
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

	p.sumv += v
	p.done += v
}

func (p *Progress) add(t time.Duration, v int64) {
	if t <= p.elem[p.w].t {
		p.elem[p.w].v += v
		return
	}

	p.updateStatus(p.elem[p.w].t, t)

	p.incw()
	p.elem[p.w] = elem{t: t, v: v}
	st := t - p.window
	for p.r != p.w && p.elem[p.r].t <= st {
		p.incr()
	}
}

func (p *Progress) updateStatus(lastt, t time.Duration) {
	p.status.Done = p.done
	p.status.V = p.sumv
	if t >= p.window {
		p.status.Acc = true
		p.status.Dt = p.window
	} else {
		p.status.Dt = t
	}

	if p.status.Total <= 0 {
		return
	}

	// calculate finished time
	left, ok := p.status.timeLeft()
	if !ok {
		return
	}

	finish := int64((t + left) / resolution)
	confidence := int64((t - lastt) / resolution)
	p.favg.add(confidence, finish)

	if t < p.window {
		return
	}

	if finish, ok = p.favg.value(); ok {
		p.status.TimeLeft = time.Duration(finish)*resolution - t
		p.status.ETA = p.start.Add(p.status.TimeLeft)
	}
}

func (p *Progress) incw() {
	p.w = (p.w + 1) % len(p.elem)
	if p.w == p.r {
		p.incr()
	}
}

func (p *Progress) incr() {
	p.sumv -= p.elem[p.r].v
	p.r = (p.r + 1) % len(p.elem)
}

type Status struct {
	Done  int64
	Total int64

	// Acc is set when accuracy should be adequate.
	// It is set once enough progress data is collected.
	Acc bool

	TimeLeft time.Duration // remaining time
	ETA      time.Time     // expected time when Done reaches Total

	// V/Dt is throughput
	V  int64
	Dt time.Duration
}

func (s Status) timeLeft() (time.Duration, bool) {
	if s.Total <= 0 || s.Dt <= 0 {
		return 0, false // unknown
	}

	left := float64(s.Total - s.Done)
	throughput := float64(s.V) / float64(s.Dt.Seconds())
	if throughput < 1 {
		return 0, false // inaccurate/unknown
	}
	t := time.Duration(left / throughput * float64(time.Second))
	return t, true
}

func (s Status) String() string {
	buf := new(bytes.Buffer)
	if s.Total > 0 {
		r := float64(s.Done) / float64(s.Total)
		fmt.Fprintf(buf, "%7.2f%%  ", float64(r)*100)
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
		buf.WriteString("0 ")
	case tp < 10:
		fmt.Fprintf(buf, "%.1f %s", tp, sbsuffix[sfxi])
	default:
		fmt.Fprintf(buf, "%.0f %s", tp, sbsuffix[sfxi])
	}
	buf.WriteString("B/s")

	if s.TimeLeft > 0 {
		fmt.Fprintf(buf, "  ETA %s", s.TimeLeft.Truncate(time.Second))
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

type weightValue struct {
	weight int64
	value  int64
}

type weightValueAvg struct {
	v    []weightValue
	w, r int

	sum weightValue
}

func (x *weightValueAvg) add(weight, value int64) {
	if weight <= 0 {
		return
	}

	x.w = (x.w + 1) % len(x.v)
	if x.w == x.r {
		elem := x.v[x.r]
		x.sum.weight -= elem.weight
		x.sum.value -= elem.value * elem.weight
		x.r = (x.r + 1) % len(x.v)
	}

	x.v[x.w].weight = weight
	x.v[x.w].value = value

	x.sum.weight += weight
	x.sum.value += value * weight
}

func (x *weightValueAvg) value() (int64, bool) {
	if x.sum.weight <= 0 {
		return 0, false
	}
	return x.sum.value / x.sum.weight, true
}
