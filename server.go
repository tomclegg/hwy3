package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
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

	inject    io.Writer
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

// start data pipeline.
//
// {inject | input} -> command -> chunk -> mp3 -> tee
func (ch *channel) run() {
	log := logrus.WithFields(logrus.Fields{
		"channel": ch.name,
	})

	log.Info("start")
	defer log.Info("stop")

	var tee io.WriteCloser = &ch.tee
	defer tee.Close()

	var src io.ReadCloser
	if ch.Input == "" {
		src, ch.inject = io.Pipe()
	} else {
		if inch, ok := ch.hwy3.Channels[ch.Input]; !ok {
			log.Fatalf("bad input channel %q", ch.Input)
		} else {
			src = inch.NewReader()
			defer src.Close()
		}
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

func (ch *channel) Inject(w http.ResponseWriter, req *http.Request) {
	if ch.hwy3.ctlServer == nil || req.Context().Value(http.ServerContextKey) != ch.hwy3.ctlServer {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	if ch.inject == nil {
		http.Error(w, "cannot inject", http.StatusBadRequest)
	}
	// TODO: prevent concurrent injects
	io.Copy(ch.inject, req.Body)
}

func (ch *channel) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		ch.Inject(w, req)
		return
	}
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
	ControlSocket string
	Listen        string
	LogFormat     string
	Channels      map[string]*channel
	clients       int64
	ctlServer     *http.Server
}

func (h *hwy3) Inject(channel string, rdr io.Reader) error {
	hc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", h.ControlSocket)
			},
		},
	}
	u, err := url.Parse("http://" + h.Listen)
	if err != nil {
		return err
	}
	addr, err := u.Parse(channel)
	if err != nil {
		return err
	}
	resp, err := hc.Post(addr.String(), "application/octet-stream", rdr)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP status %d", resp.StatusCode)
	}
	return nil
}

func (h *hwy3) Start() error {
	errs := make(chan error)

	for name, ch := range h.Channels {
		ch.name = name
		ch.hwy3 = h
	}
	for _, ch := range h.Channels {
		go func(ch *channel) { ch.NewReader().Close() }(ch)
	}

	if h.ControlSocket != "" {
		if err := os.Remove(h.ControlSocket); err != nil && !os.IsNotExist(err) {
			return err
		}
		ln, err := net.Listen("unix", h.ControlSocket)
		if err != nil {
			return err
		}
		err = os.Chmod(h.ControlSocket, 0777)
		if err != nil {
			return err
		}
		h.ctlServer = &http.Server{
			Handler: h,
		}
		go func() {
			errs <- h.ctlServer.Serve(ln)
		}()
	}

	if h.Listen != "" {
		srv := &http.Server{
			Addr:    h.Listen,
			Handler: h,
		}
		go func() {
			errs <- srv.ListenAndServe()
		}()
	}

	return <-errs
}

func (h *hwy3) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
	config := flag.String("config", "config.yaml", "yaml or json configuration `file`")
	inject := flag.String("inject", "", "inject stdin to specified `channel`")
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

	if *inject != "" {
		logrus.Fatal(h.Inject(*inject, os.Stdin))
	}

	logrus.Fatal(h.Start())
}
