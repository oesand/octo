package flow

import (
	"fmt"
	"math/rand"
	"time"
)

// NewUid returns a short unique identifier used to identify flow instances.
// It composes a small random prefix and the current millisecond timestamp.
func NewUid() string {
	sugar := make([]byte, 2)
	for i := 0; i < len(sugar); i++ {
		sugar[i] = byte(rand.Int63() & 0xff)
	}
	now := time.Now().UnixMilli()
	return fmt.Sprintf("%x_%x", sugar, now)
}
