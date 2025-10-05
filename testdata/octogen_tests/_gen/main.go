package main

import (
	"bufio"
	"fmt"
	"github.com/oesand/octo/internal"
	"github.com/oesand/octo/internal/parse"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const currentModule = "github.com/oesand/octo/testdata"

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix("[testgen]: ")

	entries, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Select test for regenerate:")

	var files []os.FileInfo
	for _, entry := range entries {
		if entry.Name()[0] == '_' || !entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, info)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	names := make([]string, len(files))
	for i, f := range files {
		fmt.Printf("[%d] %s \n", i+1, f.Name())
		names[i] = f.Name()
	}

	fmt.Println(strings.Repeat("-", 10))

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter number: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(names) {
		fmt.Println("Invalid choice")
		return
	}

	selected := names[idx-1]
	log.Printf("Selected tests: %s \n", selected)

	path, err := filepath.Abs(selected)
	if err != nil {
		log.Fatalf("error fail to get abs: %s", err.Error())
	}

	packages, warns, errs := parse.ParseInjects(currentModule, path)

	if warns != nil {
		for _, warn := range warns {
			log.Println(warn)
		}
	}

	if errs != nil {
		for _, err := range errs {
			log.Println(err)
		}
		os.Exit(1)
	}

	if len(packages) != 1 {
		var pkgNames []string
		for _, pkg := range packages {
			names = append(names, pkg.Name)
		}
		log.Fatalf("too many packages found: %v\n", pkgNames)
	}

	pkg := packages[0]
	genPath := filepath.Join(pkg.Path, "want_gen.go")
	err = internal.GenerateFile(genPath, pkg)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("generated successfully")
	}
}
