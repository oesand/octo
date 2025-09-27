// cmd/octogen-gen/main.go
package main

import (
	"fmt"
	"github.com/oesand/octo/internal/parse"
	"log"
)

func main() {
	injects, err := parse.ParseInjects()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("PkgDecls: %v \n", injects)
}
