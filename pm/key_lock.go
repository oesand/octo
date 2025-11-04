package pm

import "sync"

// KeyLock provides per-key locking similar to singleflight.Group,
// but without requiring a function callback.
// Each key has its own internal *sync.Mutex.
// Once all waiters for a key are done, its mutex is removed automatically.
type KeyLock[K comparable] struct {
	mu    sync.Mutex
	locks map[K]*lockEntry
}

type lockEntry struct {
	mu   sync.Mutex
	wait int
}

// Lock acquires the lock for the given key.
func (kl *KeyLock[K]) Lock(key K) {
	kl.mu.Lock()
	if kl.locks == nil {
		kl.locks = make(map[K]*lockEntry)
	}

	entry, ok := kl.locks[key]
	if !ok {
		entry = &lockEntry{}
		kl.locks[key] = entry
	}
	entry.wait++
	kl.mu.Unlock()

	entry.mu.Lock()
}

// TryLock attempts to acquire the lock for the given key without blocking.
// It returns true if the lock was successfully acquired, and false otherwise.
func (kl *KeyLock[K]) TryLock(key K) bool {
	kl.mu.Lock()
	if kl.locks == nil {
		kl.locks = make(map[K]*lockEntry)
	}

	entry, ok := kl.locks[key]
	if !ok {
		// create new lock entry
		entry = &lockEntry{}
		kl.locks[key] = entry
	}
	entry.wait++
	kl.mu.Unlock()

	// Attempt to acquire the lock without blocking.
	locked := entry.mu.TryLock()
	if !locked {
		// Could not acquire — roll back wait count
		kl.mu.Lock()
		entry.wait--
		if entry.wait == 0 {
			delete(kl.locks, key)
		}
		kl.mu.Unlock()
	}
	return locked
}

// Unlock releases the lock for the given key.
func (kl *KeyLock[K]) Unlock(key K) {
	kl.mu.Lock()
	if kl.locks == nil {
		kl.mu.Unlock()
		panic("keylock: unlock of unlocked key")
	}

	entry, ok := kl.locks[key]
	if !ok {
		kl.mu.Unlock()
		panic("keylock: unlock of unlocked key")
	}

	entry.wait--
	if entry.wait == 0 {
		delete(kl.locks, key)
	}
	kl.mu.Unlock()

	entry.mu.Unlock()
}
