package main

import (
	"flag"
	"fmt"
	"github.com/oesand/octo/cmd"
	"github.com/oesand/octo/internal"
	"github.com/oesand/octo/internal/parse"
	"golang.org/x/mod/modfile"
	"log"
	"os"
	"path/filepath"
)

const UsageText = "Usage of octogen: \n" +
	"\t octogen # Help - you here ;) \n" +
	"\t octogen version # Print version \n" +
	"\t octogen -gen # Scans packages and generate Injects follows instructions \n" +
	"\t\t ... -name <name:str> # Defines filename for instructions aggregation. Must ends with `.go` extension. Default: octo_gen.go \n" +
	"For more information, see: \n\thttps://github.com/oesand/octo \n"

func PrintUsage() {
	fmt.Print(UsageText)
}

func main() {
	args := os.Args
	if len(args) == 2 && args[1] == "version" {
		fmt.Printf("octogen %s \n", cmd.Version)
		return
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix("[octogen]: ")
	flag.Usage = PrintUsage

	startGeneration := flag.Bool("gen", false, "mark for run generation")
	generationName := flag.String("name", "octo_gen.go", "file name for generation")
	flag.Parse()

	if !*startGeneration {
		PrintUsage()
		return
	}

	if filepath.Ext(*generationName) != ".go" {
		log.Fatalf("'%s' is not a .go file name", *generationName)
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

	genFileName := filepath.Clean(*generationName)

	for _, pkg := range packages {
		filePath := filepath.Join(pkg.Path, genFileName)

		log.Printf("generating package %s \n", pkg.Path)

		err := internal.GenerateFile(filePath, pkg)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("generation finished")
}
