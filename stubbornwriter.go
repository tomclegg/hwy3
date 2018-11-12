package main

import (
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

// stubbornWriter wraps a writer, retrying indefinitely on errors.
type stubbornWriter struct {
	writer io.Writer
	logger logrus.FieldLogger
	retry  time.Duration
}

func newStubbornWriter(w io.Writer, lgr logrus.FieldLogger, retry time.Duration) stubbornWriter {
	return stubbornWriter{
		writer: w,
		logger: lgr,
		retry:  retry,
	}
}

func (sw stubbornWriter) Write(p []byte) (int, error) {
	done := 0
	retry := sw.retry
	for {
		n, err := sw.writer.Write(p[done:])
		done += n
		if err == nil {
			return done, nil
		}
		if n > 0 {
			retry = sw.retry
		}
		sw.logger.WithError(err).Warn("write error")
		time.Sleep(retry)
		if retry < sw.retry*1000 {
			retry = retry * 11 / 10
		}
	}
}
