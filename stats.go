package main

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

const (
	bucketDuration = 2 * time.Second
	bucketCount    = 180
)

type rateTracker struct {
	io.Writer
	buckets []int64
	last    time.Time
	sync.Mutex
}

func (rt *rateTracker) update(delta int64, current []int64) {
	rt.Lock()
	defer rt.Unlock()
	if rt.buckets == nil {
		rt.buckets = make([]int64, bucketCount)
		rt.last = time.Now()
	} else if shift := int(time.Since(rt.last) / bucketDuration); shift > 0 {
		rt.last = rt.last.Add(bucketDuration * time.Duration(shift))
		copy(rt.buckets[shift:], rt.buckets)
		for i := 0; i < shift; i++ {
			rt.buckets[i] = 0
		}
	}
	rt.buckets[0] += delta
	if current != nil {
		copy(current, rt.buckets)
	}
}

func (rt *rateTracker) Write(p []byte) (n int, err error) {
	n, err = rt.Writer.Write(p)
	rt.update(int64(n), nil)
	return
}

type trackers struct {
	rts map[*rateTracker]string
	sync.Mutex
}

func (t *trackers) Copy(w io.Writer, r io.Reader, label string) (n int64, err error) {
	rt := t.Add(w, label)
	defer t.Remove(rt)
	return io.Copy(rt, r)
}

func (t *trackers) Add(w io.Writer, label string) *rateTracker {
	rt := &rateTracker{Writer: w}
	t.Lock()
	defer t.Unlock()
	if t.rts == nil {
		t.rts = make(map[*rateTracker]string)
	}
	t.rts[rt] = label
	return rt
}

func (t *trackers) Remove(rt *rateTracker) {
	t.Lock()
	defer t.Unlock()
	delete(t.rts, rt)
}

func (t *trackers) MarshalJSON() ([]byte, error) {
	t.Lock()
	defer t.Unlock()
	chBytes := make(map[string][]int64, len(t.rts))
	for rt, name := range t.rts {
		chBytes[name] = make([]int64, bucketCount)
		rt.update(0, chBytes[name])
	}
	return json.Marshal(map[string]interface{}{
		"SampleIntervalNS": bucketDuration.Nanoseconds(),
		"ChannelBytes":     chBytes,
	})
}
