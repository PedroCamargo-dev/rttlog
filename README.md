<div align="center">

# 🚀 rttlog

### A lightweight RTT (latency) monitor via ICMP

_Simple CLI • Low overhead • Useful metrics_

[![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey?style=for-the-badge)](https://github.com/PedroCamargo-dev/rttlog)

</div>

---

## ✨ Features

- **🌐 ICMP Echo** - IPv4 support (IPv6 when available)
- **⚡ Parallel monitoring** - One worker per target
- **📊 Time-windowed summaries** - Default 10s windows with comprehensive metrics
- **🔇 Quiet mode** - Clean output with summaries only
- **🚨 Spike detection** - Alerts for high latency
- **🧪 Comprehensive testing** - Unit, race, and integration tests

### 📈 Example Output

```bash
[rttlog] 2025/12/24 13:54:34 [10s] 1.1.1.1: sent=9 ok=9 loss=0.0% avg=8.5ms p95=9.2ms min=7.2ms max=9.2ms
```

---

## 📥 Installation

### Option 1: Pre-built Binaries (Recommended)

Download the latest release for your platform from the [Releases page](https://github.com/PedroCamargo-dev/rttlog/releases).

### Option 2: Build from Source

```bash
go build -o rttlog ./cmd/rttlog
```

---

## 🚀 Usage

### Command Line Flags

| Flag         | Type     | Description                            | Example           |
| ------------ | -------- | -------------------------------------- | ----------------- |
| `--targets`  | string   | Comma-separated targets (IP or domain) | `1.1.1.1,8.8.8.8` |
| `--interval` | duration | Interval between probes per target     | `1s`, `250ms`     |
| `--timeout`  | duration | Probe timeout                          | `1s`              |
| `--summary`  | duration | Summary window duration                | `10s`, `30s`      |
| `--quiet`    | flag     | Only show summaries + errors + spikes  | -                 |
| `--spike`    | duration | Spike threshold (0 disables)           | `80ms`            |

### 💡 Quick Examples

**Basic monitoring with 10s summaries:**

```bash
./rttlog --targets 1.1.1.1,8.8.8.8 --interval 1s --timeout 1s --summary 10s --quiet
```

**Advanced: 30s summaries + spike detection:**

```bash
./rttlog --targets 1.1.1.1,8.8.8.8 --interval 1s --timeout 1s --summary 30s --quiet --spike 80ms
```

---

## 🔐 ICMP Permissions

> ⚠️ **Important:** ICMP requires raw socket permissions

### 🐧 Linux (Recommended)

```bash
sudo setcap cap_net_raw+ep ./rttlog
./rttlog --targets 1.1.1.1 --quiet
```

### 🍎 macOS

```bash
sudo ./rttlog --targets 1.1.1.1 --quiet
```

### 🪟 Windows

Run PowerShell/CMD as **Administrator**:

```cmd
rttlog.exe --targets 1.1.1.1,8.8.8.8 --quiet
```

---

## 🧪 Testing

### Unit Tests

```bash
go test ./... -count=1
```

### Race Detection

```bash
go test ./... -race -count=1
```

### Integration Tests (Optional)

```bash
# Basic integration
RTTLOG_INTEGRATION=1 go test -tags=integration ./... -count=1

# Linux with real raw sockets
sudo -E env RTTLOG_INTEGRATION=1 go test -tags=integration ./... -count=1
```

---

## 🏗️ Multi-Platform Build

Create binaries for all platforms in `dist/` directory:

```bash
rm -rf dist && mkdir -p dist

# Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/rttlog-linux-amd64 ./cmd/rttlog
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/rttlog-linux-arm64 ./cmd/rttlog

# Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/rttlog-windows-amd64.exe ./cmd/rttlog

# macOS
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/rttlog-darwin-amd64 ./cmd/rttlog
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/rttlog-darwin-arm64 ./cmd/rttlog
```

---

## 🔧 Troubleshooting

### ❌ "socket: operation not permitted"

**Cause:** Missing ICMP raw socket permissions

**Solutions:**

- **Linux:** `sudo setcap cap_net_raw+ep ./rttlog`
- **macOS:** Run with `sudo`
- **Windows:** Run as Administrator

### 📡 High packet loss / timeouts

**Possible causes:**

- ICMP blocked by firewall/router/ISP
- Network instability
- Target doesn't respond to ping
- Incorrect target address

---

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

<div align="center">

</div>
