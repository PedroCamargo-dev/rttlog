// internal/stats/window_test.go
package stats

import (
	"strings"
	"testing"
	"time"

	"github.com/PedroCamargo-dev/rttlog/internal/probe"
)

type errSentinel struct{}

func (errSentinel) Error() string { return "boom" }

func TestWindowSummaryAndReset(t *testing.T) {
	w := NewWindow()

	for i := 0; i < 8; i++ {
		w.Add(probe.Result{OK: true, RTT: time.Duration(10+i) * time.Millisecond})
	}
	w.Add(probe.Result{OK: false, Err: errSentinel{}})
	w.Add(probe.Result{OK: false, Err: errSentinel{}})

	line := w.SummaryLine("1.1.1.1", 10*time.Second)
	if !strings.Contains(line, "sent=10") || !strings.Contains(line, "ok=8") || !strings.Contains(line, "loss=20.0%") {
		t.Fatalf("unexpected summary: %s", line)
	}

	w.Reset()
	line2 := w.SummaryLine("1.1.1.1", 10*time.Second)
	if !strings.Contains(line2, "no samples") {
		t.Fatalf("expected no samples after reset: %s", line2)
	}
}
