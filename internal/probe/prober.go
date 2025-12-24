package probe

type Prober interface {
	Probe(target string, seq int) Result
	Close() error
}
