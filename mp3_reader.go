package main

import (
	"io"

	"github.com/tcolgate/mp3"
)

type MP3Reader struct {
	d *mp3.Decoder
}

// NewMP3Reader returns an io.Reader that reads one MP3 frame at a
// time from r.
func NewMP3Reader(r io.Reader) *MP3Reader {
	return &MP3Reader{d: mp3.NewDecoder(r)}
}

// Read reads a full mp3 frame. If p is too small to hold the frame,
// return the part that does fit (and io.ErrShortBuffer) and discard
// the rest of the frame.
func (mr *MP3Reader) Read(p []byte) (n int, err error) {
	var f mp3.Frame
	err = mr.d.Decode(&f)
	if err != nil {
		return
	}
	r := f.Reader()

	var x int
	for {
		x, err = r.Read(p)
		n += x
		if len(p) == 0 && err == nil {
			err = io.ErrShortBuffer
		}
		if err != nil {
			break
		}
		p = p[x:]
	}
	if err == io.EOF {
		err = nil
	}
	return
}
