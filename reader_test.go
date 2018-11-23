package progress_test

import (
	"bytes"
	"crypto/sha1"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/tajtiattila/progress"
)

func TestReader(t *testing.T) {
	const (
		bytesPerTick = 100
		tickLen      = time.Duration(100 * time.Millisecond)

		testLen = 60 * time.Second
		nbytes  = int64(testLen / tickLen * bytesPerTick)
	)
	h0 := sha1.New()
	if _, err := io.Copy(h0, testReader(nbytes)); err != nil {
		t.Fatal("initial copy:", err)
	}

	epoch := time.Date(2010, 1, 1, 12, 0, 0, 0, time.Local)
	wantfinish := epoch.Add(testLen)

	h1 := sha1.New()
	r := &fakeTimeReader{
		r:            testReader(nbytes),
		t:            epoch,
		bytesPerTick: bytesPerTick,
		tickLen:      tickLen,
	}

	statusok := false
	buf := &bytes.Buffer{}

	pr := progress.Reader(buf, r, nbytes,
		progress.NowFunc(r.Time),
		progress.StatusFunc(func(s progress.Status) {
			// wait for status to become accurate
			if r.t.Before(epoch.Add(7 * time.Second)) {
				return
			}
			if !s.Acc {
				t.Error("status inaccurate")
				return
			}
			statusok = true
			d := s.ETA.Sub(wantfinish)
			dp := d
			if dp < 0 {
				dp = -dp
			}
			if dp > time.Second/2 {
				t.Errorf("invalid ETA after reading %d bytes, diff=%s", r.nread, d)
			}
		}))

	if _, err := io.Copy(h1, pr); err != nil {
		t.Fatal("copying with progress:", err)
	}

	if !statusok {
		t.Error("status callback does not work")
	}

	if !bytes.Equal(h0.Sum(nil), h1.Sum(nil)) {
		t.Error("hashes differ")
	}

	if buf.Len() == 0 {
		t.Error("nothing printed")
	} else {
		t.Log("output:", string(bytes.Replace(buf.Bytes(), []byte{'\r'}, []byte{'\n'}, -1)))
	}
}

type fakeTimeReader struct {
	r io.Reader

	nread int64

	t     time.Time // current time
	ntick int       // data read so far in this tick

	bytesPerTick int
	tickLen      time.Duration
}

func (r *fakeTimeReader) Read(p []byte) (n int, err error) {
	if m := r.bytesPerTick - r.ntick; len(p) > m {
		p = p[:m]
	}
	n, err = r.r.Read(p)
	r.nread += int64(n)
	r.ntick += n
	if r.ntick >= r.bytesPerTick {
		r.ntick -= r.bytesPerTick
		r.t = r.t.Add(r.tickLen)
	}
	return n, err
}

func (r *fakeTimeReader) Time() time.Time {
	return r.t
}

func testReader(size int64) io.Reader {
	return &io.LimitedReader{
		R: rand.New(rand.NewSource(0)),
		N: size,
	}
}
