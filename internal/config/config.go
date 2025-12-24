package config

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

type Config struct {
	Targets        []string
	Interval       time.Duration
	Timeout        time.Duration
	SummaryEvery   time.Duration
	Quiet          bool
	SpikeThreshold time.Duration
}

func ParseFlags() (Config, error) {
	var cfg Config
	var targetsCSV string

	flag.StringVar(&targetsCSV, "targets", "1.1.1.1,8.8.8.8", "comma-separated targets")
	flag.DurationVar(&cfg.Interval, "interval", 2*time.Second, "probe interval per target")
	flag.DurationVar(&cfg.Timeout, "timeout", 1*time.Second, "probe timeout")
	flag.DurationVar(&cfg.SummaryEvery, "summary", 10*time.Second, "summary interval (e.g. 10s, 30s)")
	flag.BoolVar(&cfg.Quiet, "quiet", false, "only print summary + errors (+spikes)")
	flag.DurationVar(&cfg.SpikeThreshold, "spike", 0, "spike threshold (e.g. 80ms). 0 disables")

	flag.Parse()

	cfg.Targets = parseTargets(targetsCSV)
	if len(cfg.Targets) == 0 {
		return Config{}, fmt.Errorf("no targets provided")
	}
	if cfg.Interval <= 0 {
		return Config{}, fmt.Errorf("interval must be > 0")
	}
	if cfg.Timeout <= 0 {
		return Config{}, fmt.Errorf("timeout must be > 0")
	}
	if cfg.SummaryEvery <= 0 {
		return Config{}, fmt.Errorf("summary must be > 0")
	}

	return cfg, nil
}

func parseTargets(csv string) []string {
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
