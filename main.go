package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/mod/modfile"
)

var Version = "v0.0.0"

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
	modfilePath := flag.String("modfile", "./go.mod", "path to go.mod file")
	help := flag.Bool("help", false, "show help")
	version := flag.Bool("version", false, "show version")
	onlyVersion := flag.Bool("only-version", false, "only print the version")
	flag.Parse()

	if *help {
		flag.Usage = printUsage(os.Stdout)
		flag.Usage()
		return
	}

	if *version {
		fmt.Fprint(os.Stdout, Version+"\n")
		return
	}

	if len(flag.Args()) == 0 {
		printError("no package provided", nil, true)
	}

	if *modfilePath == "" {
		printError("go.mod file not provided", nil, true)
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

func printUsage(w io.Writer) func() {
	return func() {
		flag.CommandLine.SetOutput(w)
		fmt.Fprint(w, usage)
		flag.PrintDefaults()
		fmt.Fprint(w, usageExample)
	}
}

// printError prints an error message and exits
func printError(s string, err error, with ...bool) {
	if err == nil {
		fmt.Fprint(os.Stderr, s+"\n\n")
	} else {
		fmt.Fprintf(os.Stderr, s+": %s\n\n", err)
	}

	if len(with) > 0 && with[0] {
		flag.Usage = printUsage(os.Stderr)
		flag.Usage()
	}

	// panic if in test mode to simulate os.Exit
	if os.Getenv("TEST_MODE") == "true" {
		panic(s)
	}

	// exit with status 1
	os.Exit(1)
}
