package internal

import "os"

type CtxKey struct {
	Key string
}

func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true // exists
	}
	if os.IsNotExist(err) {
		return false // definitely does not exist
	}
	// some other error (e.g., permission denied)
	return false
}
