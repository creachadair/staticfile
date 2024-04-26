# staticfile

> This package is obsolete. New programs should use https://pkg.go.dev/embed instead.
> This repository remains to support existing use.

This repository provides a tool to compile static data into a Go binary, and to
access those data via a file-like interface.

The `compiledata` program generates a Go source file in the specified package
that embeds the contents of the named file globs:

    compiledata -pkg staticdata -out static.go data/*

The resulting file can be compiled into a package in the usual way.  This tool
can also be invoked from `go generate` rules.

In common use, the main package will blank import the static data package, and
other packages access the files via the `staticfile` package:

```go
import "github.com/creachadair/staticfile"

f, err := staticfile.Open("data/main.css")
...
defer f.Close()
doStuffWith(f)
```
