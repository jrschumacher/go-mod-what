package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"golang.org/x/mod/modfile"
)

var Version = "v0.0.0"

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <package> [<package> ...]\n\n", os.Args[0])

		flag.PrintDefaults()
	}

	modfilePath := flag.String("modfile", "./go.mod", "path to go.mod file")
	help := flag.Bool("help", false, "show help")
	version := flag.Bool("version", false, "show version")
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *version {
		fmt.Fprintf(os.Stdout, Version+"\n")
		return
	}

	if *modfilePath == "" {
		printError("go.mod file not provided", nil)
		return
	}

	if flag.NArg() < 1 {
		printError("package name not provided", nil)
		return
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

	found := false
	for _, p := range flag.Args() {
		for _, r := range m.Require {
			if r.Mod.Path == p {
				found = true
				fmt.Fprintf(os.Stdout, "%s %s\n", r.Mod.Path, r.Mod.Version)
			}
		}
	}

	if !found {
		printError("module not found", nil)
	}
}

func printError(s string, err error) {
	if err == nil {
		fmt.Fprintf(os.Stderr, s+"\n\n")
	} else {
		fmt.Fprintf(os.Stderr, s+": %s\n\n", err)
	}
	flag.Usage()
	os.Exit(1)
}
