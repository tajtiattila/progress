package progress

import (
	"fmt"
	"io"
	"os"
)

// Reader returns a reader that reads from r
// and writes progress to output.
func Reader(output io.Writer, r io.Reader, size int64, opt ...Option) io.Reader {
	return &limitReader{
		r: &progressReader{
			r:      r,
			output: output,
			p:      New(size, opt...),
		},
		limit: 1 << 20, // 1 MiB
	}
}

// StderrReader returns a reader that reads from r
// and displays progress on os.Stderr.
//
// Progress is written only if os.Stderr
// is a device (eg. console).
func StderrReader(r io.Reader, size int64, opt ...Option) io.Reader {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return r
	}
	if (fi.Mode() & os.ModeDevice) != 0 {
		// don't write progress to file or pipe
		return r
	}
	return Reader(os.Stderr, r, size, opt...)
}

type progressReader struct {
	r io.Reader

	output io.Writer
	p      *Progress
}

func (r *progressReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	changed := r.p.Update(int64(n))
	if err == nil && changed {
		const cr = "\r"            // move cursor to start of current line
		const clearline = "\x1b[K" // escape sequence to clear rest of line
		fmt.Fprint(r.output, cr, r.p.Status(), clearline)
	}
	return n, err
}

// limitReader limits the amount of bytes
// read from the underlying reader
type limitReader struct {
	r io.Reader

	limit int
}

func (l *limitReader) Read(p []byte) (n int, err error) {
	for len(p) > l.limit {
		m, err := l.r.Read(p[:l.limit])
		n += m
		if err != nil {
			return n, err
		}
		p = p[l.limit:]
	}
	m, err := l.r.Read(p)
	return n + m, err
}
