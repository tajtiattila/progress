package progress_test

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/tajtiattila/progress"
)

type elem struct {
	v int64
	t time.Duration
}

func TestProgress(t *testing.T) {

	f, err := os.Open("testdata/progress.dump")
	if err != nil {
		t.Fatal("load test data:", err)
	}
	defer f.Close()

	var dumps [][]elem
	start := true

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()
		var tx float64
		var vx int64
		_, err := fmt.Sscanln(l, &tx, &vx)
		if err == nil {
			if start {
				dumps = append(dumps, nil)
				start = false
			}
			last := len(dumps) - 1
			dumps[last] = append(dumps[last], elem{
				v: vx,
				t: time.Duration(tx * float64(time.Second)),
			})
		} else {
			start = true
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal("test data scan:", err)
	}

	for i, d := range dumps {
		if len(d) != 0 {
			t.Logf("test #%d\n", i)
			testProgressDump(t, d)
		}
	}
}

func TestProgressFixed(t *testing.T) {
	testProgressFixed(t, time.Millisecond, 1)
	testProgressFixed(t, 100*time.Millisecond, 100)
}

func testProgressFixed(t *testing.T, timestep time.Duration, value int64) {
	var dump []elem

	var tt time.Duration
	for ; tt < 60*time.Second; tt += timestep {
		dump = append(dump, elem{
			t: tt,
			v: value,
		})
	}

	// append last element to have fixed total length
	dump = append(dump, elem{
		t: tt,
		v: 0,
	})

	t.Logf("test fixed %d/%s", value, timestep)
	testProgressDump(t, dump)
}

func testProgressDump(t *testing.T, h []elem) {
	var total int64
	for _, e := range h {
		total += e.v
	}

	t.Logf("total=%d", total)
	const res = 100 * time.Millisecond

	maxt := h[len(h)-1].t
	epoch := time.Date(2010, 1, 1, 0, 0, 0, 0, time.Local)
	var ct time.Duration

	p := progress.New(total, progress.NowFunc(func() time.Time {
		return epoch.Add(ct)
	}))

	trueend := epoch.Add(maxt)

	lastt := time.Duration(0)
	for _, e := range h {
		ct = e.t
		p.Update(e.v)
		curt := ct.Truncate(res)
		if curt != lastt {
			lastt = curt

			s := p.Status()
			var diffs string
			tr, ok := s.TimeRemaining()
			if ok {
				end := epoch.Add(ct + tr)
				diff := end.Sub(trueend).Truncate(time.Second)
				diffs = diff.String()
				if diff >= 0 {
					diffs = "+" + diffs
				}
			}
			rem := (maxt - curt).Truncate(res)
			t.Logf("%10s %10s %s", rem, diffs, s)
		}
	}
}

func int64SliceString(vv []int64) string {
	buf := &bytes.Buffer{}
	buf.WriteString("[")
	for i, v := range vv {
		if i != 0 {
			buf.WriteString(",")
		}
		fmt.Fprintf(buf, "%d", v)
	}
	buf.WriteString("]")
	return buf.String()
}
