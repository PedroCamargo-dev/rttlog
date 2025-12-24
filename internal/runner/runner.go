package runner

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/PedroCamargo-dev/rttlog/internal/probe"
	"github.com/PedroCamargo-dev/rttlog/internal/stats"
)

type Options struct {
	SummaryEvery   time.Duration
	Quiet          bool
	SpikeThreshold time.Duration
}

type Runner interface {
	Start() error
	Stop() error
}

type Ticker interface {
	C() <-chan time.Time
	Stop()
}

type Clock interface {
	NewTicker(d time.Duration) Ticker
}

type realClock struct{}

func (realClock) NewTicker(d time.Duration) Ticker {
	return &realTicker{t: time.NewTicker(d)}
}

type realTicker struct {
	t *time.Ticker
}

func (rt *realTicker) C() <-chan time.Time { return rt.t.C }
func (rt *realTicker) Stop()               { rt.t.Stop() }

type defaultRunner struct {
	prober   probe.Prober
	targets  []string
	interval time.Duration
	logger   *log.Logger
	opt      Options
	clock    Clock

	ctx    context.Context
	cancel context.CancelFunc

	workersWG  sync.WaitGroup
	consumerWG sync.WaitGroup

	mu      sync.Mutex
	started bool

	results chan probe.Result
}

func NewRunner(p probe.Prober, targets []string, interval time.Duration, opt Options, logger *log.Logger) Runner {
	return NewRunnerWithClock(p, targets, interval, opt, logger, realClock{})
}

func NewRunnerWithClock(p probe.Prober, targets []string, interval time.Duration, opt Options, logger *log.Logger, clock Clock) Runner {
	if opt.SummaryEvery <= 0 {
		opt.SummaryEvery = 10 * time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())

	return &defaultRunner{
		prober:   p,
		targets:  targets,
		interval: interval,
		logger:   logger,
		opt:      opt,
		clock:    clock,
		ctx:      ctx,
		cancel:   cancel,
		results:  make(chan probe.Result, 4096),
	}
}

func (r *defaultRunner) Start() error {
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return fmt.Errorf("runner already started")
	}
	r.started = true
	r.mu.Unlock()

	r.consumerWG.Add(1)
	go r.consume()

	for _, target := range r.targets {
		t := target

		r.workersWG.Add(1)
		go func() {
			defer r.workersWG.Done()

			ticker := r.clock.NewTicker(r.interval)
			defer ticker.Stop()

			seq := 1

			r.results <- r.prober.Probe(t, seq)
			seq++

			for {
				select {
				case <-r.ctx.Done():
					return
				case <-ticker.C():
					r.results <- r.prober.Probe(t, seq)
					seq++
				}
			}
		}()
	}

	return nil
}

func (r *defaultRunner) Stop() error {
	r.mu.Lock()
	alreadyStopped := !r.started
	r.mu.Unlock()
	if alreadyStopped {
		return nil
	}

	r.cancel()

	r.workersWG.Wait()
	close(r.results)
	r.consumerWG.Wait()

	_ = r.prober.Close()
	return nil
}

func (r *defaultRunner) consume() {
	defer r.consumerWG.Done()

	wins := make(map[string]*stats.Window, len(r.targets))
	for _, t := range r.targets {
		wins[t] = stats.NewWindow()
	}

	summaryTicker := r.clock.NewTicker(r.opt.SummaryEvery)
	defer summaryTicker.Stop()

	for {
		select {
		case res, ok := <-r.results:
			if !ok {

				for _, t := range r.targets {
					r.logger.Print(wins[t].SummaryLine(t, r.opt.SummaryEvery))
				}
				return
			}

			w := wins[res.Target]
			if w == nil {
				w = stats.NewWindow()
				wins[res.Target] = w
			}
			w.Add(res)

			if res.Err != nil {
				r.logger.Printf("Probe to %s failed: %v", res.Target, res.Err)
				continue
			}

			if r.opt.SpikeThreshold > 0 && res.OK && res.RTT >= r.opt.SpikeThreshold {
				r.logger.Printf("[spike] %s (%s): seq=%d time=%v", res.Target, res.IP, res.Seq, res.RTT)
			}

			if !r.opt.Quiet {
				r.logger.Printf("Probe to %s (%s): seq=%d time=%v", res.Target, res.IP, res.Seq, res.RTT)
			}

		case <-summaryTicker.C():
			for _, t := range r.targets {
				r.logger.Print(wins[t].SummaryLine(t, r.opt.SummaryEvery))
				wins[t].Reset()
			}
		}
	}
}
