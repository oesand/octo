package mc

import (
	"container/list"
	"iter"
	"sync"
)

// WithLruEviction returns an Option that configures the cache to use an
// LRU (least-recently-used) usage evictor with the specified capacity.
//
// It panics:
//
// - if `capacity` is less than 1
//
// - if a usage evictor was already configured on the cache.
func WithLruEviction(capacity int) Option {
	return func(cache *MemCache) {
		if capacity < 1 {
			panic("capacity must be greater than zero")
		}

		if cache.usageEvictor != nil {
			panic("evictor by usage already set")
		}

		cache.usageEvictor = newLruEvictor(capacity)
	}
}

// newLruEvictor creates and initializes an LRU-based usage evictor with
// the provided capacity. The returned evictor tracks usage order so that
// oldest (least recently used) keys can be identified for eviction.
func newLruEvictor(capacity int) *lruEvictor {
	return &lruEvictor{
		capacity:     capacity,
		evictList:    list.New(),
		evictEntries: make(map[string]*list.Element, capacity),
	}
}

type lruEvictor struct {
	mu           sync.Mutex
	capacity     int
	evictList    *list.List
	evictEntries map[string]*list.Element
}

func (lru *lruEvictor) Used(key string) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if ent, has := lru.evictEntries[key]; has {
		lru.evictList.MoveToFront(ent)
	} else {
		lru.evictEntries[key] = lru.evictList.PushFront(key)
	}
}

func (lru *lruEvictor) GetExcess() int {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if ll := lru.evictList.Len(); ll > lru.capacity {
		return ll - lru.capacity
	}
	return 0
}

func (lru *lruEvictor) IterWorst() iter.Seq[string] {
	return func(yield func(string) bool) {
		lru.mu.Lock()
		defer lru.mu.Unlock()

		ent := lru.evictList.Back()
		for ent != nil {
			if !yield(ent.Value.(string)) {
				break
			}
			ent = ent.Prev()
		}
	}
}

func (lru *lruEvictor) Remove(key string) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if ent, has := lru.evictEntries[key]; has {
		lru.evictList.Remove(ent)
		delete(lru.evictEntries, key)
	}
}
