# rttlog

A lightweight RTT (latency) monitor via ICMP (ping) with time-windowed summary.
Focus: Simple CLI, low overhead, and useful metrics (avg/p95/min/max/loss) per target.

Example (quiet mode + summary):
[rttlog] 2025/12/24 13:54:34 [10s] 1.1.1.1: sent=9 ok=9 loss=0.0% avg=8.5ms p95=9.2ms min=7.2ms max=9.2ms

---

## FEATURES

- ICMP Echo (IPv4; IPv6 when supported).
- Multiple targets in parallel (one worker per target).
- Summary per window (default: 10s):
  sent, ok, loss%, avg, p95, min, max
- --quiet: prints only summary + errors (+ spikes).
- --spike: alert when RTT >= threshold.
- Tests: unit + race + optional integration (ICMP).

---

## INSTALLATION

1. Via Releases (recommended)

- Download the package for your system from the repository's Releases page and extract.

2. Local build (dev)
   go build -o rttlog ./cmd/rttlog

---

## USAGE

Flags:
--targets string Targets separated by comma (IP or domain).
--interval duration Interval between probes per target. Ex: 1s, 250ms.
--timeout duration Probe timeout. Ex: 1s.
--summary duration Summary window. Ex: 10s, 30s.
--quiet Only summary + errors (+ spikes).
--spike duration Spike threshold. Ex: 80ms. 0 disables.

Examples:

Summary every 10s (quiet):
./rttlog --targets 1.1.1.1,8.8.8.8 --interval 1s --timeout 1s --summary 10s --quiet

Summary every 30s + spike >= 80ms:
./rttlog --targets 1.1.1.1,8.8.8.8 --interval 1s --timeout 1s --summary 30s --quiet --spike 80ms

---

## ICMP PERMISSIONS (IMPORTANT)

ICMP normally requires permission to open raw socket.

Linux (recommended: setcap)
sudo setcap cap_net_raw+ep ./rttlog
./rttlog --targets 1.1.1.1 --interval 1s --timeout 1s --summary 10s --quiet

macOS
sudo ./rttlog --targets 1.1.1.1 --interval 1s --timeout 1s --summary 10s --quiet

Windows

- Run terminal as Administrator:
  rttlog.exe --targets 1.1.1.1,8.8.8.8 --interval 1s --timeout 1s --summary 10s --quiet

---

## TESTS

Unit tests:
go test ./... -count=1

Race detector:
go test ./... -race -count=1

ICMP integration (optional):
RTTLOG_INTEGRATION=1 go test -tags=integration ./... -count=1

Linux (real integration, raw socket):
sudo -E env RTTLOG_INTEGRATION=1 go test -tags=integration ./... -count=1

---

## MULTI-PLATFORM BUILD (OUTPUT IN dist/)

rm -rf dist && mkdir -p dist

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/rttlog-linux-amd64 ./cmd/rttlog
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/rttlog-linux-arm64 ./cmd/rttlog
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/rttlog-windows-amd64.exe ./cmd/rttlog
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/rttlog-darwin-amd64 ./cmd/rttlog
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/rttlog-darwin-arm64 ./cmd/rttlog

---

## TROUBLESHOOTING

Error: "socket: operation not permitted"

- Missing ICMP raw socket permission.
  Linux: sudo setcap cap_net_raw+ep ./rttlog
  macOS: run with sudo
  Windows: run as Administrator

High loss / timeouts

- Possible causes:
  - ICMP blocked by firewall/router/ISP
  - network instability
  - target doesn't respond to ping

---

## LICENSE

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
