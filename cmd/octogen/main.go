package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/oesand/octo/internal"
	"github.com/oesand/octo/internal/octogen/parse"
	"golang.org/x/mod/modfile"
)

const UsageText = "Usage of octogen: \n" +
	"\t octogen # Help - you here ;) \n" +
	"\t octogen gen # Scans packages and generate Injects follows instructions \n" +
	"\t\t ... -name <name:str> # Defines filename for instructions aggregation. Must ends with `.go` extension. Default: octo_gen.go \n" +
	"For more information, see: \n\thttps://github.com/oesand/octo \n"

func PrintUsage() {
	fmt.Print(UsageText)
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix("[octogen]: ")
	flag.Usage = PrintUsage

	generationName := flag.String("name", "octo_gen.go", "file name for generation")
	flag.Parse()

	args := flag.Args()
	switch {
	case len(args) == 1 && args[0] == "gen":
		runGen(*generationName)
	default:
		PrintUsage()
	}
}

func runGen(genName string) {
	if filepath.Ext(genName) != ".go" {
		genName = genName + ".go"
	}

	if !internal.IsFileExist("./go.mod") {
		log.Fatalln("go.mod not found, must run only in module directory")
	}

	modData, err := os.ReadFile("go.mod")
	if err != nil {
		log.Fatalf("failed to read go.mod: %s", err)
	}

	currentModule := modfile.ModulePath(modData)
	if currentModule == "" {
		log.Fatalln("unknown module name in go.mod")
	}

	packages, warns, errs := parse.Parse(currentModule, ".")

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

	genFileName := filepath.Clean(genName)

	for _, pkg := range packages {
		filePath := filepath.Join(pkg.Dir(), genFileName)

		log.Printf("generating package '%s'...\n", pkg.Path())

		err = os.WriteFile(filePath, pkg.Render(), 0666)
		if err != nil {
			log.Printf("failed to generate: %s\n", err)
		} else {
			log.Println("generated successfully")
		}
	}

	log.Println("generation finished")
}
