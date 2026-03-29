package flow

import (
	"fmt"
	"math/rand"
	"time"
)

func NewUID() string {
	sugar := make([]byte, 2)
	for i := 0; i < len(sugar); i++ {
		sugar[i] = byte(rand.Int63() & 0xff)
	}
	now := time.Now().UnixMilli()
	return fmt.Sprintf("%x_%x", sugar, now)
}
