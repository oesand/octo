// cmd/octogen-gen/main.go
package main

import (
	"fmt"
	"github.com/oesand/octo/internal/parse"
	"log"
	"os"
)

func main() {
	injects, errs := parse.ParseInjects()
	if errs != nil {
		for _, err := range errs {
			log.Println(err)
		}
		os.Exit(1)
	}

	fmt.Printf("PkgDecls: %v \n", injects)
}
