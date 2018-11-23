package progress

import (
	"fmt"
	"io"
	"os"
)

// Writer returns a writer that writes to w,
// and displays progress output.
func Writer(output, w io.Writer, size int64) io.Writer {
	return &limitWriter{
		w: &progressWriter{
			w:      w,
			output: output,
			p:      New(size),
		},
		limit: 1 << 20, // 1 MiB
	}
}

// StderrWriter returns a writer that writes to w,
// and displays progress on os.Stderr.
//
// Progress is written only if os.Stderr
// is a device (eg. console).
func StderrWriter(w io.Writer, size int64) io.Writer {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return w
	}
	if (fi.Mode() & os.ModeDevice) != 0 {
		// don't write progress to file or pipe
		return w
	}
	return Writer(os.Stderr, w, size)
}

type progressWriter struct {
	w io.Writer

	output io.Writer
	p      *Progress
}

func (w *progressWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	changed := w.p.Update(int64(n))
	if err != nil && changed {
		const cr = "\r"            // move cursor to start of current line
		const clearline = "\x1b[K" // escape sequence to clear rest of line
		fmt.Fprint(w.output, cr, w.p.Status(), clearline)
	}
	return n, err
}

// limitWriter writes at most limit bytes at once
type limitWriter struct {
	w io.Writer

	limit int
}

func (w *limitWriter) Write(p []byte) (n int, err error) {
	for len(p) > w.limit {
		m, err := w.w.Write(p[:w.limit])
		n += m
		if err != nil {
			return n, err
		}
		p = p[w.limit:]
	}
	m, err := w.w.Write(p)
	return n + m, err
}
