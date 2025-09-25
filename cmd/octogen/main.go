// cmd/octogen-gen/main.go
package main

import (
	"fmt"
	"github.com/oesand/octo/internal/parse"
)

func main() {
	injects := parse.ParseInjects()

	fmt.Printf("PkgDecls: %v \n", injects)
}
