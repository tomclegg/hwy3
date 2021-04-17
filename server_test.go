package main

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	check "gopkg.in/check.v1"
)

type ServerSuite struct{}

var _ = check.Suite(&ServerSuite{})

func (s *Suite) TestCloseClients(c *check.C) {
	ln, err := net.Listen("tcp", ":")
	c.Assert(err, check.IsNil)
	lnaddr := ln.Addr().String()
	ln.Close()

	h := &hwy3{
		Listen: lnaddr,
		Channels: map[string]*channel{
			"/test": &channel{
				Command: "sleep 10",
				Chunk:   1,
				Buffers: 1,
			},
		},
	}
	go h.Start()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 10; i++ {
		req, err := http.NewRequestWithContext(ctx, "GET", "http://"+lnaddr+"/test", nil)
		c.Assert(err, check.IsNil)
		go func() {
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				c.Logf("client status %s", resp.Status)
				io.Copy(ioutil.Discard, resp.Body)
			}
		}()
	}

	deadline := time.Now().Add(10 * time.Second)

	for atomic.LoadInt32(&h.clients) < 10 {
		if time.Now().After(deadline) {
			c.Error("timeout waiting for test clients to connect")
			return
		}
		time.Sleep(time.Millisecond)
	}

	cancel()

	for atomic.LoadInt32(&h.clients) > 0 {
		if time.Now().After(deadline) {
			c.Error("timeout waiting for handlers to return")
			return
		}
		time.Sleep(time.Millisecond)
	}
}
