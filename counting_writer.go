package main

import (
	"io"
	"sync/atomic"
)

type countingWriter struct {
	io.Writer
	count uint64
}

func (cw *countingWriter) Write(p []byte) (n int, err error) {
	n, err = cw.Writer.Write(p)
	atomic.AddUint64(&cw.count, uint64(n))
	return
}

func (cw *countingWriter) Count() uint64 {
	return atomic.LoadUint64(&cw.count)
}
