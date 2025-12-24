package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/PedroCamargo-dev/rttlog/internal/config"
	"github.com/PedroCamargo-dev/rttlog/internal/probe"
	"github.com/PedroCamargo-dev/rttlog/internal/runner"
)

func main() {
	cfg, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(os.Stdout, "[rttlog] ", log.LstdFlags)

	prober := probe.NewICMPProber(cfg.Timeout)

	r := runner.NewRunner(
		prober,
		cfg.Targets,
		cfg.Interval,
		runner.Options{
			SummaryEvery:   cfg.SummaryEvery,
			Quiet:          cfg.Quiet,
			SpikeThreshold: cfg.SpikeThreshold,
		},
		logger,
	)

	if err := r.Start(); err != nil {
		logger.Fatalf("start: %v", err)
	}

	logger.Printf(
		"running. targets=%v interval=%s timeout=%s summary=%s quiet=%v spike=%s (Ctrl+C to stop)",
		cfg.Targets, cfg.Interval, cfg.Timeout, cfg.SummaryEvery, cfg.Quiet, cfg.SpikeThreshold,
	)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	logger.Println("stopping...")
	_ = r.Stop()
	logger.Println("bye")
}
