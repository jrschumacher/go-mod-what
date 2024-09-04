# go mod what

You can see `go mod why` a package exists in the module graph, and `go mod graph` shows the module graph, but
there is no command to show the module in the version of the package in the module graph.

## Install

```bash
go install github.com/jrschumacher/go-mod-what@latest
```

## Usage

```bash
Usage: go-mod-what [options] <package> [<package> ...]

  -help
        show help
  -modfile string
        path to go.mod file (default "./go.mod")
```

This will output the module and version of the package `golang.org/x/mod v0.21.0` within the `go.mod` file.

Example:

```bash
$ go-mod-what golang.org/x/mod
golang.org/x/mod v0.21.0
```

Example with multiple packages:

```bash
$ go-mod-what golang.org/x/mod golang.org/x/tools
golang.org/x/mod v0.21.0
golang.org/x/tools v0.1.0
```

Example with a custom `go.mod` file:

```bash
$ go-mod-what -modfile /path/to/go.mod golang.org/x/mod
golang.org/x/mod v0.21.0
```
