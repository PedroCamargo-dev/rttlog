//go:build integration

package probe

import (
	"errors"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestICMPProbeIntegration(t *testing.T) {
	if os.Getenv("RTTLOG_INTEGRATION") == "" {
		t.Skip("set RTTLOG_INTEGRATION=1 to run integration tests")
	}

	p := NewICMPProber(1 * time.Second)
	defer p.Close()

	res := p.Probe("1.1.1.1", 1)
	if res.Err != nil {
		if errors.Is(res.Err, syscall.EPERM) || errors.Is(res.Err, syscall.EACCES) ||
			strings.Contains(res.Err.Error(), "operation not permitted") {
			t.Skip("ICMP raw socket requires privileges. Run with sudo, or build test binary and setcap it.")
		}
		t.Fatalf("probe failed: %v", res.Err)
	}

	if !res.OK {
		t.Fatalf("expected OK=true, got false")
	}
	if res.RTT <= 0 {
		t.Fatalf("expected RTT > 0, got %v", res.RTT)
	}
}
