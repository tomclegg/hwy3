package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/tomclegg/nbtee"
	"gopkg.in/tylerb/graceful.v1"
)

var (
	addr     = flag.String("listen", ":80", "local listening address, :port or host:port")
	buffers  = flag.Int("buffers", 100, "max frames to buffer for each client")
	chunk    = flag.Int("chunk", 0, "send/skip data in chunks of N bytes (0 for any size)")
	grace    = flag.Duration("grace", 0, "on TERM/INT, wait for clients to disconnect")
	graceEOF = flag.Duration("grace-eof", 0, "on EOF, wait for clients to disconnect")
	logTimes = flag.Bool("log-timestamps", true, "prefix log messages with timestamp")
	mp3only  = flag.Bool("mp3", false, "send only full MP3 frames to clients")
)

// signalCloser wraps an io.Writer. When it gets closed, it closes its
// Closed channel, in order to notify other goroutines.
type signalCloser struct {
	io.Writer
	Closed chan struct{}
}

func (sc *signalCloser) Close() error {
	close(sc.Closed)
	return nil
}

// A teeHandler handles http requests by reading data from an nbtee.
type teeHandler struct {
	*nbtee.Writer
	clients int64
}

func (th *teeHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t0 := time.Now()
	if *mp3only {
		w.Header().Set("Content-Type", "audio/mpeg")
	}
	cw := &countingWriter{Writer: w}
	sc := &signalCloser{Writer: cw, Closed: make(chan struct{})}
	log.Printf("%d +%+q %+q", atomic.AddInt64(&th.clients, 1), req.RemoteAddr, req.URL)
	th.Add(sc)

	errs := make(chan error)
	go func() {
		if w, ok := w.(http.CloseNotifier); ok {
			<-w.CloseNotify()
		} else {
			<-sc.Closed
		}
		errs <- th.RemoveAndClose(sc)
	}()
	err := <-errs
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	t := time.Since(t0)
	log.Printf("%d -%+q %+q %s %d =%dB/s %+q", atomic.AddInt64(&th.clients, -1), req.RemoteAddr, req.URL, t, cw.Count(), int64(float64(cw.Count())/t.Seconds()), errStr)
}

func main() {
	flag.Parse()

	if *logTimes {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	} else {
		log.SetFlags(0)
	}

	th := &teeHandler{Writer: nbtee.NewWriter(*buffers).Start()}

	srv := &graceful.Server{
		Timeout: *grace,
		Server: &http.Server{
			Addr:    *addr,
			Handler: th,
		},
	}
	go func() {
		var r io.Reader = bufio.NewReaderSize(os.Stdin, 16384)
		if *chunk > 0 {
			r = &ChunkReader{Reader: r, Size: *chunk}
		}
		if *mp3only {
			r = NewMP3Reader(r)
		}
		n, err := io.Copy(th, r)
		if err != nil {
			log.Println("stdin:", err)
		}
		log.Printf("read %d bytes", n)
		th.Close()
		srv.Stop(*graceEOF)
	}()
	srv.ListenAndServe()
}
