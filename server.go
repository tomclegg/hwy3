package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	_ "github.com/tomclegg/canfs"
	"github.com/tomclegg/mp3dir"
	"github.com/tomclegg/nbtee2"
)

//go:generate make README.md
//go:generate go run $GOPATH/src/github.com/tomclegg/canfs/generate.go -pkg=main -id=sysUI -out=generated_sys_ui.go -dir=./ui_sys
//go:generate go run $GOPATH/src/github.com/tomclegg/canfs/generate.go -pkg=main -id=archiveUI -out=generated_archive_ui.go -dir=./ui_archive

type channel struct {
	Input       string  // input channel (can be empty)
	Command     string  // shell command to run on input
	Calm        float64 // minimum seconds between restarts
	Buffers     int     // max buffered writes per listener
	BufferLow   int     // min buffered writes after underrun
	Chunk       int     // if > 0, write only fixed-size blocks
	MP3         bool    // write only complete mp3 frames
	ContentType string  // Content-Type response header

	MP3Dir    mp3dir.Writer
	archive   http.Handler
	archiveUI http.Handler

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
	if ch.MP3Dir.Root != "" {
		if ch.MP3Dir.SplitOnSize > 0 {
			go ch.hwy3.trackers.Copy(&ch.MP3Dir, ch.tee.NewReader(ch.BufferLow, ch.Buffers), "write:MP3Dir:"+ch.name)
		}
		if ch.MP3Dir.BitRate > 0 {
			ch.archive = http.StripPrefix(ch.name, http.FileServer(&ch.MP3Dir))
			ch.archiveUI = http.StripPrefix(ch.name+"/ui/", http.FileServer(archiveUI))
		}
	}
}

// run data for channel forever.
//
// {inject | input} -> loop { runCommandAndFilter }
func (ch *channel) run() {
	log := logrus.WithFields(logrus.Fields{
		"channel": ch.name,
	})

	log.Info("start")
	defer log.Info("stop")

	var src io.Reader
	src, ch.inject = io.Pipe()
	if inch, ok := ch.hwy3.Channels[ch.Input]; !ok && ch.Input != "" {
		log.Fatalf("bad input channel %q", ch.Input)
	} else if ok {
		go func() {
			r := inch.NewReader()
			defer r.Close()
			ch.hwy3.trackers.Copy(ch.inject, r, "input:"+ch.name)
		}()
	}

	d := time.Duration(ch.Calm * float64(time.Second))
	if d <= 0 {
		d = time.Second
	}
	for calm := time.NewTicker(d).C; ; <-calm {
		ch.runCommandAndFilter(&ch.tee, src, log)
	}
}

// pipe data through command and filters if specified.
//
// {command -> chunk -> mp3 -> tee}
func (ch *channel) runCommandAndFilter(w io.Writer, r io.Reader, log *logrus.Entry) {
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

		if stdout, err := cmd.StdoutPipe(); err != nil {
			log.Fatalf("StdoutPipe(): %s", err)
		} else {
			defer stdout.Close()
			r = stdout
		}

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
	if ch.Chunk > 1 {
		w = bufio.NewWriterSize(w, ch.Chunk)
	}

	size, err := ch.hwy3.trackers.Copy(w, r, "output:"+ch.name)
	log.WithField("ReadBytes", size).WithError(err).Info("EOF")
}

func (ch *channel) NewReader() io.ReadCloser {
	ch.setupOnce.Do(ch.setup)
	return ch.tee.NewReader(ch.BufferLow, ch.Buffers)
}

func (ch *channel) Inject(w http.ResponseWriter, req *http.Request) {
	if sv := req.Context().Value(http.ServerContextKey); sv == nil || sv != ch.hwy3.ctlServer {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	if ch.inject == nil {
		http.Error(w, "cannot inject", http.StatusBadRequest)
	}
	// TODO: prevent concurrent injects
	inj := ch.inject
	if xc := req.Header.Get("X-Chunk"); xc != "" {
		chunk, err := strconv.Atoi(xc)
		if err != nil || chunk > 1<<24 || chunk < 0 {
			http.Error(w, fmt.Sprintf("bad x-chunk header %q", xc), http.StatusBadRequest)
			return
		}
		if chunk > 1 {
			inj = bufio.NewWriterSize(inj, chunk)
		}
	}
	ch.hwy3.trackers.Copy(inj, req.Body, "input:"+ch.name)
}

func (ch *channel) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != ch.name {
		if ch.archive == nil {
			code := http.StatusNotFound
			http.Error(w, http.StatusText(code), code)
			return
		}
		if req.Method != "GET" && req.Method != "HEAD" {
			code := http.StatusMethodNotAllowed
			http.Error(w, http.StatusText(code), code)
			return
		}
		if strings.HasPrefix(req.URL.Path, ch.name+"/ui/") {
			ch.archiveUI.ServeHTTP(w, req)
			return
		}
		if fnm := req.FormValue("filename"); fnm != "" {
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fnm))
		} else {
			w.Header().Set("Content-Disposition", "attachment")
		}
		w.Header().Set("Content-Type", ch.ContentType)
		ch.archive.ServeHTTP(w, req)
		return
	}
	if req.Method == "POST" {
		ch.Inject(w, req)
		return
	}
	if ch.ContentType != "" {
		w.Header().Set("Content-Type", ch.ContentType)
	}
	rdr := ch.NewReader()
	defer rdr.Close()
	ch.hwy3.trackers.Copy(w, rdr, "client:"+ch.name+","+req.Header.Get("X-Request-Id")+","+req.Header.Get("X-Forwarded-For")+","+req.RemoteAddr)
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
	ListenTLS     string
	CertFile      string
	KeyFile       string
	LogFormat     string
	Channels      map[string]*channel

	Theme struct {
		Title string `json:"title"`
	}

	clients   int32
	ctlServer *http.Server
	cert      chan *tls.Certificate // server/updater can safely borrow/replace a *cert

	trackers trackers
}

func (h *hwy3) Inject(channel string, rdr io.Reader, chunk int) error {
	hc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", h.ControlSocket)
			},
		},
	}
	u, err := url.Parse("http://localhost")
	if err != nil {
		return err
	}
	addr, err := u.Parse(channel)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", addr.String(), rdr)
	if err != nil {
		return err
	}
	req.Header.Set("X-Chunk", strconv.FormatInt(int64(chunk), 10))
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP status %d", resp.StatusCode)
	}
	return nil
}

func (h *hwy3) serveTheme(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.Theme)
}

func (h *hwy3) serveChannels(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	type chaninfo struct {
		Archive bool `json:"archive"`
	}
	channels := make(map[string]chaninfo, len(h.Channels))
	for name, ch := range h.Channels {
		channels[name] = chaninfo{
			Archive: ch.MP3Dir.BitRate > 0,
		}
	}
	json.NewEncoder(w).Encode(channels)
}

func (h *hwy3) serveStats(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&h.trackers)
}

func (h *hwy3) middleware(mux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("X-Request-Id") == "" {
			req.Header.Set("X-Request-Id", fmt.Sprintf("%x", time.Now().UnixNano()))
		}
		cw := &counter{ResponseWriter: w}
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
		atomic.AddInt32(&h.clients, 1)
		defer atomic.AddInt32(&h.clients, -1)

		mux.ServeHTTP(cw, req)

		t := time.Since(t0)
		log.WithFields(logrus.Fields{
			"Bytes":          cw.bytes,
			"BytesPerSecond": int64(float64(cw.bytes) / t.Seconds()),
			"Seconds":        t.Seconds(),
		}).Info("end")
	})
}

func (h *hwy3) Start() error {
	for name, ch := range h.Channels {
		ch.name = name
		ch.hwy3 = h
	}
	for _, ch := range h.Channels {
		go func(ch *channel) { ch.NewReader().Close() }(ch)
	}

	mux := http.NewServeMux()
	mux.Handle("/sys/ui/", http.StripPrefix("/sys/ui/", http.FileServer(sysUI)))
	mux.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(archiveUI)))
	mux.HandleFunc("/sys/stats", h.serveStats)
	mux.HandleFunc("/sys/channels", h.serveChannels)
	mux.HandleFunc("/sys/theme", h.serveTheme)
	mux.HandleFunc("/", h.serveHTTP)

	stack := h.middleware(mux)

	errs := make(chan error)

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
			Handler: stack,
		}
		go func() {
			errs <- h.ctlServer.Serve(ln)
		}()
	}

	if h.Listen != "" {
		srv := &http.Server{
			Addr:    h.Listen,
			Handler: stack,
		}
		go func() {
			errs <- srv.ListenAndServe()
		}()
	}

	if h.ListenTLS != "" {
		h.cert = make(chan *tls.Certificate, 1)
		go h.ensureCurrentCertificate()

		srv := &http.Server{
			Addr:    h.ListenTLS,
			Handler: stack,
			TLSConfig: &tls.Config{
				GetCertificate: h.getCertificate,
				//NextProtos: []string{"h2", "http/1.1"},
				PreferServerCipherSuites: true,
				CurvePreferences: []tls.CurveID{
					tls.CurveP256,
					tls.X25519,
				},
				MinVersion: tls.VersionTLS12,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				},
			},
		}
		go func() {
			errs <- srv.ListenAndServeTLS("", "")
		}()
	}

	return <-errs
}

func (h *hwy3) getCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	if h.cert == nil {
		panic("no cert chan")
	}
	cert := <-h.cert
	h.cert <- cert
	return cert, nil
}

func (h *hwy3) ensureCurrentCertificate() {
	for first := true; ; first = false {
		cert, err := tls.LoadX509KeyPair(h.CertFile, h.KeyFile)
		if err != nil {
			if first {
				logrus.WithError(err).Fatal("error loading TLS certificate")
			}
			logrus.WithError(err).Warn("error loading TLS certificate")
			time.Sleep(time.Second)
			continue
		}
		if !first {
			<-h.cert
		}
		h.cert <- &cert
		time.Sleep(time.Minute)
	}
}

func (h *hwy3) serveHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		http.Redirect(w, req, "/ui/", http.StatusFound)
		return
	}
	for last := len(req.URL.Path); last >= 0; last = strings.LastIndexByte(req.URL.Path[:last], '/') {
		if ch, ok := h.Channels[req.URL.Path[:last]]; ok {
			ch.ServeHTTP(w, req)
			break
		} else if last == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			break
		}
	}
}

func main() {
	config := flag.String("config", "hwy3.yaml", "yaml or json configuration `file`")
	inject := flag.String("inject", "", "inject stdin to specified `channel`")
	chunk := flag.Int("chunk", 0, "use `n`-byte chunks when injecting")
	flag.Parse()

	var h hwy3

	buf, err := ioutil.ReadFile(*config)
	if err != nil {
		logrus.Fatal(err)
	}
	err = yaml.Unmarshal(buf, &h)
	if err != nil {
		logrus.WithField("config", *config).Fatal(err)
	}

	if h.LogFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	if *inject != "" {
		logrus.Fatal(h.Inject(*inject, os.Stdin, *chunk))
	}

	logrus.Fatal(h.Start())
}
