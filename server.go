package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/tomclegg/nbtee2"
)

type channel struct {
	Input       string  // input channel (can be empty)
	Command     string  // shell command to run on input
	Calm        float64 // minimum seconds between restarts
	Buffers     int     // # buffered writes per listener
	Chunk       int     // if > 0, write only fixed-size blocks
	MP3         bool    // write only complete mp3 frames
	ContentType string  // Content-Type response header

	tee       nbtee2.Tee
	setupOnce sync.Once

	hwy3 *hwy3
	name string
}

func (ch *channel) setup() {
	if ch.Buffers == 0 {
		ch.Buffers = 4
	}
	if ch.MP3 && ch.ContentType == "" {
		ch.ContentType = "audio/mpeg"
	}
	go ch.run()
}

var readNothing = ioutil.NopCloser(&bytes.Buffer{})

func (ch *channel) run() {
	log := logrus.WithFields(logrus.Fields{
		"channel": ch.name,
	})

	log.Info("start")
	defer log.Info("stop")

	var tee io.WriteCloser = &ch.tee
	defer tee.Close()

	var src io.ReadCloser
	if ch.Input != "" {
		if inch, ok := ch.hwy3.Channels[ch.Input]; !ok {
			log.Fatalf("bad input channel %q", ch.Input)
		} else {
			src = inch.NewReader()
			defer src.Close()
		}
	} else {
		src = readNothing
	}

	d := time.Duration(ch.Calm * float64(time.Second))
	if d <= 0 {
		d = time.Second
	}
	calm := time.NewTicker(d).C

	for ; ; <-calm {
		func() {
			r, w := src, tee
			var err error
			if ch.Command != "" {
				cmd := exec.Command("sh", "-c", ch.Command)
				cmd.Stdin = r

				stderr, err := cmd.StderrPipe()
				if err != nil {
					log.Fatalf("StderrPipe(): %s", err)
				}
				go func() {
					defer stderr.Close()
					log := log.WithField("Stderr", true)
					r := bufio.NewReader(stderr)
					var s string
					var err error
					for err == nil {
						s, err = r.ReadString('\n')
						if s != "" {
							log.Info(s)
						}
					}
				}()

				r, err = cmd.StdoutPipe()
				if err != nil {
					log.Fatalf("StdoutPipe(): %s", err)
				}
				defer r.Close()

				defer func() {
					go func() {
						err := cmd.Wait()
						log.WithError(err).Error("command exit")
					}()
				}()
				log.WithField("ExecArgs", cmd.Args).Info("command start")
				cmd.Start()
			}
			if r == readNothing {
				log.Error("no input")
				<-context.Background().Done()
				return
			}
			if ch.MP3 {
				w = &MP3Writer{Writer: w}
			}

			var size int64
			if ch.Chunk > 0 {
				buf := make([]byte, ch.Chunk)
				for err == nil {
					var n int
					n, err = io.ReadFull(r, buf)
					if err == nil {
						size += int64(n)
						n, err = w.Write(buf)
					}
				}
			} else {
				size, err = io.Copy(w, r)
			}
			log.WithField("ReadBytes", size).WithError(err).Info("EOF")
		}()
	}
}

func (ch *channel) NewReader() io.ReadCloser {
	ch.setupOnce.Do(ch.setup)
	return ch.tee.NewReader(ch.Buffers)
}

func (ch *channel) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if ch.ContentType != "" {
		w.Header().Set("Content-Type", ch.ContentType)
	}
	rdr := ch.NewReader()
	defer rdr.Close()
	io.Copy(w, rdr)
}

type counter struct {
	http.ResponseWriter
	bytes int64
}

func (w *counter) Write(p []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(p)
	w.bytes += int64(n)
	return
}

type hwy3 struct {
	Listen    string
	LogFormat string
	Channels  map[string]*channel
	clients   int64
	setupOnce sync.Once
}

func (h *hwy3) setup() {
	for name, ch := range h.Channels {
		ch.name = name
		ch.hwy3 = h
	}
	for _, ch := range h.Channels {
		go func(ch *channel) { ch.NewReader().Close() }(ch)
	}
}

func (h *hwy3) Start() {
	h.setupOnce.Do(h.setup)
}

func (h *hwy3) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.setupOnce.Do(h.setup)
	t0 := time.Now()
	log := logrus.WithFields(logrus.Fields{
		"Request": fmt.Sprintf("%x", t0.UnixNano()),
	})
	log.WithFields(logrus.Fields{
		"RemoteAddr":    req.RemoteAddr,
		"XForwardedFor": req.Header.Get("X-Forwarded-For"),
		"Method":        req.Method,
		"Path":          req.URL.Path,
	}).Info("start")
	atomic.AddInt64(&h.clients, 1)
	defer atomic.AddInt64(&h.clients, -1)
	cw := &counter{ResponseWriter: w}
	if ch, ok := h.Channels[req.URL.Path]; ok {
		ch.ServeHTTP(cw, req)
	} else {
		http.Error(cw, "not found", http.StatusNotFound)
	}
	t := time.Since(t0)
	log.WithFields(logrus.Fields{
		"Bytes":          cw.bytes,
		"BytesPerSecond": int64(float64(cw.bytes) / t.Seconds()),
		"Seconds":        t.Seconds(),
	}).Info("end")
}

func main() {
	config := flag.String("config", "config.json", "json configuration `file`")
	flag.Parse()

	var h hwy3
	if *config != "" {
		buf, err := ioutil.ReadFile(*config)
		if err != nil {
			logrus.Fatal(err)
		}
		err = yaml.Unmarshal(buf, &h)
		if err != nil {
			logrus.WithField("config", *config).Fatal(err)
		}
	}

	if h.LogFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	h.Start()

	srv := &http.Server{
		Addr:    h.Listen,
		Handler: &h,
	}
	err := srv.ListenAndServe()
	if err != nil {
		logrus.Error(err)
	}
}
