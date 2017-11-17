package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/tomclegg/nbtee2"

	"gopkg.in/tylerb/graceful.v1"
)

var (
	addr     = flag.String("listen", ":80", "local listening address, :port or host:port")
	buffers  = flag.Int("buffers", 100, "max frames to buffer for each client")
	chunk    = flag.Int("chunk", 0, "send/skip data in chunks of N bytes (0 for any size)")
	grace    = flag.Duration("grace", 0, "on TERM/INT, wait for clients to disconnect")
	graceEOF = flag.Duration("grace-eof", 0, "on EOF, wait for clients to disconnect")
	logTimes = flag.Bool("log-timestamps", true, "prefix log messages with timestamp")
	mimeType = flag.String("mime-type", "", "send given MIME type as Content-Type header")
	mp3only  = flag.Bool("mp3", false, "send only full MP3 frames to clients, default -mime-type audio/mpeg")
)

type handler struct {
	newReader func() io.ReadCloser
	clients   int64
}

func (th *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t0 := time.Now()
	if *mimeType != "" {
		w.Header().Set("Content-Type", *mimeType)
	} else if *mp3only {
		w.Header().Set("Content-Type", "audio/mpeg")
	}
	log.Printf("%d +%+q", atomic.AddInt64(&th.clients, 1), req.RemoteAddr)
	rdr := th.newReader()
	defer rdr.Close()
	n, err := io.Copy(w, rdr)

	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	t := time.Since(t0)
	log.Printf("%d -%+q %s %d =%dB/s %+q", atomic.AddInt64(&th.clients, -1), req.RemoteAddr, t, n, int64(float64(n)/t.Seconds()), errStr)
}

func main() {
	flag.Parse()

	if *logTimes {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	} else {
		log.SetFlags(0)
	}

	tee := &nbtee2.Tee{}
	srv := &graceful.Server{
		Timeout: *grace,
		Server: &http.Server{
			Addr: *addr,
			Handler: &handler{newReader: func() io.ReadCloser {
				return tee.NewReader(*buffers)
			}},
		},
	}
	go func() {
		var size int64
		var err error
		var r io.Reader = os.Stdin
		var w io.WriteCloser = tee
		if *mp3only {
			w = &MP3Writer{Writer: w}
		}
		if *chunk > 0 {
			buf := make([]byte, *chunk)
			for err == nil {
				var n int
				n, err = io.ReadFull(r, buf)
				if n > 0 {
					n, err = w.Write(buf)
					size += int64(n)
				}
			}
		} else {
			size, err = io.Copy(w, r)
		}
		if err != nil {
			log.Println("stdin:", err)
		}
		log.Printf("read %d bytes", size)
		err = w.Close()
		if err != nil {
			log.Print(err)
		}
		srv.Stop(*graceEOF)
	}()
	err := srv.ListenAndServe()
	if err != nil {
		log.Print(err)
	}
}
