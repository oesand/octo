package parse

import "path/filepath"

func GlobFiles() ([]string, error) {
	matches, err := filepath.Glob("*.go")
	if err != nil {
		return nil, err
	}
	return matches, nil
}
