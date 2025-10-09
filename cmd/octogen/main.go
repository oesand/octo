package main

import (
	"github.com/oesand/octo/internal"
	"github.com/oesand/octo/internal/parse"
	"golang.org/x/mod/modfile"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix("[octogen]: ")

	if !internal.IsFileExist("./go.mod") {
		log.Fatalln("go.mod not found, must run only in workspace directory")
	}

	modData, err := os.ReadFile("go.mod")
	if err != nil {
		log.Fatalf("failed to read go.mod: %s", err)
	}

	currentModule := modfile.ModulePath(modData)
	if currentModule == "" {
		log.Fatalln("unknown module name in go.mod")
	}

	packages, warns, errs := parse.ParseInjects(currentModule, ".")

	if warns != nil {
		for _, warn := range warns {
			log.Println("[WARN]:", warn)
		}
	}

	if errs != nil {
		for _, err := range errs {
			log.Println(err)
		}
		os.Exit(1)
	}

	for _, pkg := range packages {
		filePath := filepath.Join(pkg.Path, "octo_gen.go")

		log.Printf("generating package %s \n", pkg.Path)

		err := internal.GenerateFile(filePath, pkg)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("generated successfully")
}
