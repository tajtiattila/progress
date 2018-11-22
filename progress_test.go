package progress_test

import (
	"bufio"
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

func testProgressDump(t *testing.T, h []elem) {
	var total int64
	for _, e := range h {
		total += e.v
	}
	tt := h[len(h)-1].t
	epoch := time.Date(2010, 1, 1, 0, 0, 0, 0, time.Local)
	var ct time.Duration

	p := progress.New(total, progress.NowFunc(func() time.Time {
		return epoch.Add(ct)
	}))

	const res = time.Duration(100 * time.Millisecond)

	for _, e := range h {
		ct = e.t
		p.Update(e.v)
		rem := tt - ct
		t.Logf("%s %s\n", rem.Truncate(res), p.Status())
	}
}
