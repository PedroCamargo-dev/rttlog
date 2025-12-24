package stats

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/PedroCamargo-dev/rttlog/internal/probe"
)

type Window struct {
	sent int
	ok   int
	rtts []time.Duration
	min  time.Duration
	max  time.Duration
}

func NewWindow() *Window {
	return &Window{rtts: make([]time.Duration, 0, 1024)}
}

func (w *Window) Add(res probe.Result) {
	w.sent++
	if !res.OK || res.Err != nil {
		return
	}
	w.ok++
	w.rtts = append(w.rtts, res.RTT)

	if w.ok == 1 {
		w.min = res.RTT
		w.max = res.RTT
		return
	}
	if res.RTT < w.min {
		w.min = res.RTT
	}
	if res.RTT > w.max {
		w.max = res.RTT
	}
}

func (w *Window) Reset() {
	w.sent = 0
	w.ok = 0
	w.rtts = w.rtts[:0]
	w.min = 0
	w.max = 0
}

func (w *Window) SummaryLine(target string, window time.Duration) string {
	if w.sent == 0 {
		return fmt.Sprintf("[%s] %s: no samples", window, target)
	}

	loss := float64(w.sent-w.ok) / float64(w.sent) * 100.0
	if w.ok == 0 {
		return fmt.Sprintf("[%s] %s: sent=%d ok=%d loss=%.1f%% (no replies)", window, target, w.sent, w.ok, loss)
	}

	avg := w.avg()
	p95 := w.p95()

	return fmt.Sprintf(
		"[%s] %s: sent=%d ok=%d loss=%.1f%% avg=%s p95=%s min=%s max=%s",
		window, target, w.sent, w.ok, loss,
		durMS(avg), durMS(p95), durMS(w.min), durMS(w.max),
	)
}

func (w *Window) avg() time.Duration {
	var sum time.Duration
	for _, d := range w.rtts {
		sum += d
	}
	return time.Duration(int64(sum) / int64(len(w.rtts)))
}

func (w *Window) p95() time.Duration {
	tmp := make([]time.Duration, len(w.rtts))
	copy(tmp, w.rtts)
	sort.Slice(tmp, func(i, j int) bool { return tmp[i] < tmp[j] })

	n := len(tmp)
	i := int(math.Ceil(0.95*float64(n))) - 1
	if i < 0 {
		i = 0
	}
	if i >= n {
		i = n - 1
	}
	return tmp[i]
}

func durMS(d time.Duration) string {
	return fmt.Sprintf("%.1fms", float64(d)/float64(time.Millisecond))
}
