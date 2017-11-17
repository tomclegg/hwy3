package main

import (
	"io"
	"sync"

	"github.com/tcolgate/mp3"
)

// MP3Writer wraps an io.Writer, passing through each MP3 frame in a
// single Write() call and dropping everything else.
type MP3Writer struct {
	Writer    io.Writer
	setupOnce sync.Once
	w         io.Writer
	closed    chan struct{}
}

func (mw *MP3Writer) setup() {
	r, w := io.Pipe()
	mw.w = w
	mw.closed = make(chan struct{})
	go func() {
		defer close(mw.closed)
		defer r.Close()
		dec := mp3.NewDecoder(r)
		buf := make([]byte, 8192)
		for {
			var f mp3.Frame
			err := dec.Decode(&f)
			if err != nil {
				return
			}
			r := f.Reader()
			var got int
			for {
				n, err := r.Read(buf[got:])
				got += n
				if err != nil {
					break
				}
				if got == len(buf) {
					newbuf := make([]byte, len(buf)*2)
					copy(newbuf, buf)
					buf = newbuf
				}
			}
			_, err = mw.Writer.Write(buf[:got])
			if err != nil {
				return
			}
		}
	}()
}

// Write implements io.Writer.
func (mw *MP3Writer) Write(p []byte) (n int, err error) {
	mw.setupOnce.Do(mw.setup)
	return mw.w.Write(p)
}

// Close waits for any buffered frames to finish writing, then closes
// the wrapped writer (if it's an io.Closer) and returns.
func (mw *MP3Writer) Close() (err error) {
	mw.setupOnce.Do(mw.setup)
	if c, ok := mw.w.(io.Closer); ok {
		err = c.Close()
	}
	<-mw.closed
	return
}
