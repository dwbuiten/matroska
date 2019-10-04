package matroska

import (
	"fmt"
	"io"
)

// fakeSeeker is just used as a lazy way to pass a io.Reader as a
// io.ReadSeeker to a function that will never invoke Seek().
//
// If you think this is gross, you're correct.
type fakeSeeker struct {
	r io.Reader
}

func (f *fakeSeeker) Read(p []byte) (int, error) {
	return f.r.Read(p)
}

func (f *fakeSeeker) Seek(offset int64, whence int) (int64, error) {
	return -1, fmt.Errorf("this is a fake seeker")
}
