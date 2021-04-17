package main

import (
	"context"
	"errors"
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

	for deadline := time.Now().Add(time.Second); ; time.Sleep(time.Millisecond) {
		ctx, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()
		conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", lnaddr)
		if err == nil {
			conn.Close()
			break
		}
		if time.Now().After(deadline) {
			c.Error("timed out waiting for server to start")
		}
	}

	nClients := 20

	for i := 0; i < nClients; i++ {
		req, err := http.NewRequestWithContext(ctx, "GET", "http://"+lnaddr+"/test", nil)
		c.Check(err, check.IsNil)
		go func() {
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				c.Logf("client status %s", resp.Status)
				io.Copy(ioutil.Discard, resp.Body)
			} else if !errors.Is(err, context.Canceled) {
				c.Errorf("client err %s", err)
			}
		}()
	}

	deadline := time.Now().Add(time.Second)

	for int(atomic.LoadInt32(&h.clients)) < nClients {
		if time.Now().After(deadline) {
			c.Error("timeout waiting for test clients to connect")
			return
		}
		if c.Failed() {
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
		if c.Failed() {
			return
		}
		time.Sleep(time.Millisecond)
	}
}
