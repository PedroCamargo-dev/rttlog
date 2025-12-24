package runner

import (
	"bytes"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/PedroCamargo-dev/rttlog/internal/probe"
)

type fakeTicker struct{ ch chan time.Time }

func (t *fakeTicker) C() <-chan time.Time { return t.ch }
func (t *fakeTicker) Stop()               {}

type fakeClock struct {
	mu      sync.Mutex
	tickers []*fakeTicker
	now     time.Time
}

func newFakeClock() *fakeClock { return &fakeClock{now: time.Unix(0, 0)} }

func (c *fakeClock) NewTicker(d time.Duration) Ticker {
	c.mu.Lock()
	defer c.mu.Unlock()

	ft := &fakeTicker{ch: make(chan time.Time, 1000)}
	c.tickers = append(c.tickers, ft)
	return ft
}

func (c *fakeClock) tickAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.now = c.now.Add(1 * time.Second)
	for _, t := range c.tickers {
		t.ch <- c.now
	}
}

type fakeProber struct{}

func (fakeProber) Probe(target string, seq int) probe.Result {
	return probe.Result{
		Target: target,
		IP:     target,
		Seq:    seq,
		OK:     true,
		RTT:    10 * time.Millisecond,
	}
}
func (fakeProber) Close() error { return nil }

func TestRunnerSummaryDeterministic(t *testing.T) {
	fc := newFakeClock()
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	r := NewRunnerWithClock(
		fakeProber{},
		[]string{"1.1.1.1", "8.8.8.8"},
		1*time.Second,
		Options{SummaryEvery: 10 * time.Second, Quiet: true},
		logger,
		fc,
	)

	if err := r.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	for i := 0; i < 10; i++ {
		fc.tickAll()
	}

	_ = r.Stop()

	out := buf.String()
	if !strings.Contains(out, "[10s] 1.1.1.1") || !strings.Contains(out, "[10s] 8.8.8.8") {
		t.Fatalf("expected summaries for both targets, got:\n%s", out)
	}
}
