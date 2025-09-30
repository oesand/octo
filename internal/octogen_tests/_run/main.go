package main

import (
	"bytes"
	"fmt"
	"github.com/oesand/octo/internal"
	"github.com/oesand/octo/internal/parse"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix("[octogen tests]: ")

	entries, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	var failed bool
	errf := func(format string, args ...any) {
		log.Printf(format+"\n", args...)
		failed = true
	}

	for _, entry := range entries {
		name := entry.Name()
		if name[0] == '_' || !entry.IsDir() {
			continue
		}

		if !entry.IsDir() {
			errf("error: not dir '%s'", name)
			continue
		}

		path, err := filepath.Abs(name)
		if err != nil {
			errf("error fail to get abs: %s", err)
			continue
		}

		log.Printf("run test %s...\n", name)

		packages, errs := parse.ParseInjects(path)
		if errs != nil {
			for _, err := range errs {
				errf(err.Error())
			}
		}

		if len(packages) != 1 {
			failed = true
			var names []string
			for _, pkg := range packages {
				names = append(names, pkg.Name)
			}
			errf("too many packages found: %v\n", names)
			continue
		}

		pkg := packages[0]

		wantPath := filepath.Join(pkg.Path, "want_gen.go")
		wantContent, err := os.ReadFile(wantPath)
		if err != nil {
			errf("open want gen file err: %s \n", err)
			continue
		}

		var buf bytes.Buffer
		err = internal.Generate(&buf, pkg)
		if err != nil {
			errf("fail to generate: %s", err)
			continue
		} else {
			log.Println("generated successfully")
		}

		actualContent := buf.Bytes()
		if bytes.Equal(wantContent, actualContent) {
			log.Printf("generated content correct \n")
		} else {
			errf("unexpected generated content\n")
			fmt.Print(string(actualContent))
		}
	}

	if failed {
		log.Println("--- FAIL ---")
		os.Exit(1)
	} else {
		log.Println("--- OK ---")
	}
}
