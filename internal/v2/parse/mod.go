package parse

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/mod/modfile"
)

func ParseModule() (string, error) {
	if !isFileExist("go.mod") {
		return "", errors.New("go.mod not found, must run only in module directory")
	}

	modData, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod: %s", err)
	}

	currentModule := modfile.ModulePath(modData)
	if currentModule == "" {
		return "", errors.New("unknown module name in go.mod")
	}
	return currentModule, nil
}

func isFileExist(path string) bool {
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
