package main

import (
	"bytes"
	"fmt"
	"github.com/oesand/octo/internal"
	"github.com/oesand/octo/internal/parse"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const currentModule = "github.com/oesand/octo/testdata"

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix("[octogen tests]: ")

	entries, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	var failed bool
	errf := func(format string, v ...any) {
		log.Printf("error: "+format+"\n", v...)
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

		log.Printf("run test '%s'...\n", name)

		errsLogsPath := filepath.Join(path, "errs.log")
		wantErrors := internal.IsFileExist(errsLogsPath)

		// TODO: Fix warns check
		packages, warns, errs := parse.ParseInjects(currentModule, path)

		if warns != nil {
			for _, warn := range warns {
				log.Println(warn)
			}
		}

		if wantErrors {
			if errs == nil {
				errf("expected errors, but errors not returned")
				continue
			}

			logsContent, err := os.ReadFile(errsLogsPath)
			if err != nil {
				errf("cannot read errs.log file err: %s", err)
				continue
			}

			expectedErrors := strings.Split(string(logsContent), "\n")
			if len(expectedErrors) != len(errs) {
				errf("expected %d errors", len(expectedErrors))
				for i, e := range expectedErrors {
					log.Printf("[%d] %s \n", i+1, e)
				}

				errf("but got %d errors", len(errs))
				for i, e := range errs {
					log.Printf("[%d] %s \n", i+1, e)
				}

				continue
			}

			var testFailed bool

			for i, expected := range expectedErrors {
				actual := errs[i]
				if strings.HasSuffix(actual.Error(), expected) {
					continue
				}

				testFailed = true
				errf("mismatch error at %d line", i+1)
				log.Printf("expected: %s \n", expected)
				log.Printf("actual: %s \n", actual)
			}

			if !testFailed {
				log.Println("test passed")
			}

			continue
		}

		if errs != nil {
			errf("got unexpected %d errors while parsing", len(errs))
			for _, err := range errs {
				log.Println(err.Error())
			}
		}

		if len(packages) == 0 {
			errf("no packages found!")
			continue
		}

		if len(packages) != 1 {
			failed = true
			var names []string
			for _, pkg := range packages {
				names = append(names, pkg.Name)
			}
			errf("too many packages found: %v", names)
			continue
		}

		pkg := packages[0]
		wantGenPath := filepath.Join(pkg.Path, "want_gen.go")
		if !internal.IsFileExist(wantGenPath) {
			errf("no want_gen file or expected error logs file for package '%s'", pkg.Path)
		}
		wantContent, err := os.ReadFile(wantGenPath)
		if err != nil {
			errf("cannot want_gen file err: %s", err)
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
			log.Println("generated content correct")
		} else {
			errf("unexpected generated content")
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
