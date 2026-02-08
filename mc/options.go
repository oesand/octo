package mc

import (
	"iter"
	"time"
)

// Option represents a function that modifies cache configuration.
type Option func(*MemCache)

// WithJanitorInterval sets interval for run background janitor
func WithJanitorInterval(interval time.Duration) Option {
	return func(m *MemCache) {
		m.janitorInterval = interval
	}
}

type usageEvictor interface {
	Used(key string)
	GetExcess() int
	IterWorst() iter.Seq[string]
	Remove(key string)
}
