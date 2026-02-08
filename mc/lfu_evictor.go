package mc

import (
	"container/list"
	"iter"
	"slices"
	"sync"
)

// WithLfuEviction returns an Option that configures the cache to use an
// LFU (least-frequently-used) usage evictor with the specified capacity.
//
// It panics:
//
// - if `capacity` is less than 1
//
// - if a usage evictor was already configured on the cache.
func WithLfuEviction(capacity int) Option {
	return func(cache *MemCache) {
		if capacity < 1 {
			panic("capacity must be greater than zero")
		}

		if cache.usageEvictor != nil {
			panic("evictor by usage already set")
		}

		cache.usageEvictor = newLfuEvictor(capacity)
	}
}

func newLfuEvictor(capacity int) *lfuEvictor {
	return &lfuEvictor{
		capacity:  capacity,
		entries:   make(map[string]*lfuEntry),
		freqLists: make(map[int]*list.List),
	}
}

type lfuEvictor struct {
	mu       sync.Mutex
	capacity int

	entries   map[string]*lfuEntry
	freqLists map[int]*list.List
}

type lfuEntry struct {
	key  string
	freq int
	elem *list.Element
}

func (lfu *lfuEvictor) Used(key string) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	// Existing entry
	if ent, ok := lfu.entries[key]; ok {
		lfu.bump(ent)
		return
	}

	// New entry
	ent := &lfuEntry{
		key:  key,
		freq: 1,
	}

	ll := lfu.freqLists[1]
	if ll == nil {
		ll = list.New()
		lfu.freqLists[1] = ll
	}

	ent.elem = ll.PushBack(ent)
	lfu.entries[key] = ent
}

func (lfu *lfuEvictor) bump(ent *lfuEntry) {
	oldFreq := ent.freq
	ll := lfu.freqLists[oldFreq]
	ll.Remove(ent.elem)

	if ll.Len() == 0 {
		delete(lfu.freqLists, oldFreq)
	}

	ent.freq++

	newLL := lfu.freqLists[ent.freq]
	if newLL == nil {
		newLL = list.New()
		lfu.freqLists[ent.freq] = newLL
	}

	ent.elem = newLL.PushBack(ent)
}

func (lfu *lfuEvictor) GetExcess() int {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if excess := len(lfu.entries) - lfu.capacity; excess > 0 {
		return excess
	}
	return 0
}

func (lfu *lfuEvictor) IterWorst() iter.Seq[string] {
	return func(yield func(string) bool) {
		lfu.mu.Lock()
		defer lfu.mu.Unlock()

		// Collect existing frequencies
		freqs := make([]int, 0, len(lfu.freqLists))
		for f := range lfu.freqLists {
			freqs = append(freqs, f)
		}

		// Lowest frequency = worst
		slices.Sort(freqs)

		for _, freq := range freqs {
			ll := lfu.freqLists[freq]

			// Oldest → newest within the same frequency
			for e := ll.Front(); e != nil; e = e.Next() {
				ent := e.Value.(*lfuEntry)
				if !yield(ent.key) {
					return
				}
			}
		}
	}
}

func (lfu *lfuEvictor) Remove(key string) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	ent, ok := lfu.entries[key]
	if !ok {
		return
	}

	ll := lfu.freqLists[ent.freq]
	ll.Remove(ent.elem)
	delete(lfu.entries, key)

	if ll.Len() == 0 {
		delete(lfu.freqLists, ent.freq)
	}
}
