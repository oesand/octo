package main

import (
	"github.com/oesand/octo/internal"
	"github.com/oesand/octo/internal/parse"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix("[octogen]: ")

	packages, errs := parse.ParseInjects(".")
	if errs != nil {
		for _, err := range errs {
			log.Println(err)
		}
		os.Exit(1)
	}

	for _, pkg := range packages {
		filePath := filepath.Join(pkg.Path, "octo_gen.go")

		log.Printf("generating package %s", pkg.Path)

		err := internal.GenerateFile(filePath, pkg)
		//err := internal.Generate(os.Stdin, pkg)
		if err != nil {
			log.Fatal(err)
		}
	}
}
