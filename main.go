package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"golang.org/x/mod/modfile"
)

const usage = `
NAME
  go-mod-what - get the version of a package in a go.mod file

SYNOPSIS
  go-mod-what [options] <package> [<package> ...]

OPTIONS
`

const usageExample = `
EXAMPLES
  To get the version of a package:
      $ go-mod-what github.com/gorilla/mux
      github.com/gorilla/mux v1.8.0

  To get the version of multiple packages:
      $ go-mod-what github.com/gorilla/mux github.com/gorilla/schema
      github.com/gorilla/context v1.1.1
      github.com/gorilla/mux v1.8.0

  To get the version of multiple packages with a wildcard:
      $ go-mod-what github.com/gorilla/*
      github.com/gorilla/context v1.1.1
      github.com/gorilla/mux v1.8.0

  To get the version of a package with a custom go.mod file path:
      $ go-mod-what -modfile ../go.mod github.com/gorilla/mux
      github.com/gorilla/mux v1.8.0

  To get the version of a package with only the version:
      $ go-mod-what -only-version github.com/gorilla/mux
      v1.8.0
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)

		flag.PrintDefaults()

		fmt.Fprint(os.Stderr, usageExample)
	}

	modfilePath := flag.String("modfile", "./go.mod", "path to go.mod file")
	help := flag.Bool("help", false, "show help")
	onlyVersion := flag.Bool("only-version", false, "only print the version")
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *modfilePath == "" {
		printError("go.mod file not provided", nil, true)
	}

	// try stating to see if it's a directory
	if path.Base(*modfilePath) != "go.mod" {
		if fi, err := os.Stat(*modfilePath); err == nil && fi.IsDir() {
			*modfilePath = path.Join(*modfilePath, "go.mod")
		} else {
			if err != nil {
				printError("could not stat go.mod file", err)
			} else {
				printError("invalid go.mod file", nil)
			}
		}
	}

	b, err := os.ReadFile(*modfilePath)
	if err != nil {
		printError("failed to read go.mod file", err)
		return
	}

	m, err := modfile.Parse(*modfilePath, b, nil)
	if err != nil {
		printError("failed to parse go.mod file", err)
	}

	found := make([]bool, len(flag.Args()))
	for _, r := range m.Require {
		for i, p := range flag.Args() {
			if !compareRequire(p, r.Mod.Path) {
				continue
			}

			found[i] = true
			modPath := r.Mod.Path + " "
			if *onlyVersion {
				modPath = ""
			}
			fmt.Fprintln(os.Stdout, modPath+r.Mod.Version)
		}
	}

	for i, f := range found {
		if !f {
			fmt.Fprintf(os.Stderr, "%s not found\n", flag.Args()[i])
		}
	}
}

// compareRequire compares module path with a string
func compareRequire(a string, b string) bool {
	// exact match
	if strings.Compare(a, b) == 0 {
		return true
	}

	// wildcard
	if strings.Contains(a, "*") && strings.HasPrefix(b, strings.TrimSuffix(a, "*")) {
		return true
	}

	return false
}

// printError prints an error message and exits
func printError(s string, err error, with ...bool) {
	if err == nil {
		fmt.Fprintf(os.Stderr, s+"\n\n")
	} else {
		fmt.Fprintf(os.Stderr, s+": %s\n\n", err)
	}

	if len(with) > 0 && with[0] {
		flag.Usage()
	}

	os.Exit(1)
}
