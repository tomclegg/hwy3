package main

import (
	"io"
)

// ChunkReader always reads exactly the specified number of bytes in
// each call to Read.
type ChunkReader struct {
	Reader io.Reader
	// Number of bytes to read in each Read()
	Size int
}

// Read reads exactly cr.Size bytes if possible, otherwise an
// error. If len(p) < cr.Size it returns io.ErrShortBuffer.
func (cr *ChunkReader) Read(p []byte) (int, error) {
	if len(p) < cr.Size {
		return 0, io.ErrShortBuffer
	}
	return io.ReadFull(cr.Reader, p[:cr.Size])
}
