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
	pw         *io.PipeWriter
	closed    chan struct{}	// closes after final Writer.Write() returns
}

func (mw *MP3Writer) setup() {
	pr, pw := io.Pipe()
	mw.pw = pw
	mw.closed = make(chan struct{})
	go func() {
		var ignoreSkipped int
		defer close(mw.closed)
		dec := mp3.NewDecoder(pr)
		buf := make([]byte, 8192)
		for {
			var f mp3.Frame
			err := dec.Decode(&f, &ignoreSkipped)
			if err != nil {
				pr.CloseWithError(err)
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
				pr.CloseWithError(err)
				return
			}
		}
	}()
}

// Write implements io.Writer.
func (mw *MP3Writer) Write(p []byte) (n int, err error) {
	mw.setupOnce.Do(mw.setup)
	return mw.pw.Write(p)
}

// Close waits for any buffered frames to finish writing.
func (mw *MP3Writer) Close() error {
	mw.setupOnce.Do(mw.setup)
	err := mw.pw.Close()
	<-mw.closed
	return err
}
