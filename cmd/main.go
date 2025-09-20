package main

import (
	"fmt"
	"github.com/oesand/octo/internal/parse"
	"go/token"
	"log"
	"os"
	"path/filepath"
)

const UsageText = "??Hello world??"

func PrintUsage() {
	fmt.Print(UsageText)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("go-enumer: ")
	//flag.Usage = PrintUsage
	//flag.Parse()

	files, err := parse.GlobFiles()
	if err != nil {
		panic(err)
	}
	fileSet := token.NewFileSet()
	for _, fileName := range files {
		absolutePath, err := filepath.Abs(fileName)
		if err != nil {
			continue
		}
		if info, err := os.Stat(absolutePath); err != nil || info.IsDir() {
			continue
		}

		_, err = parse.ParseFile(fileSet, absolutePath)
		if err != nil {
			log.Fatal(err)
		}
	}
}
