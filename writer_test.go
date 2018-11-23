package progress_test

func TestWriter(t *testing.T) {
	h0 := sha1.New()
	_, err := io.Copy(h0, testReader())
	if err != nil {
		t.Fatal("error copying:", err)
	}

	h1 := sha1.New()
}

type fakeTimeReader struct {
	t time.Time // current time

	bytesPerSec int
}

func (ftr *fakeTimeReader) Read(p []byte) (n int, err error) {

}

func (ftr *fakeTimeReader) Time() time.Time {
	return ftr.t
}

func testReader() io.Reader {
	return rand.New(rand.NewSource(0))
}
