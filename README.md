# filedata

http://godoc.org/bitbucket.org/creachadair/filedata

This repository provides tools to compile static file data into a Go binary,
and to access those files via a file-like interface.

## Overview

For the sake of discussion, let's assume you have a directory structure that
looks like this:

```
src/
  myrepo/
    data/
      static.html
      main.css
    server/
      main.go
```

Use the `compiledata` command to create a Go source package representing a set
of static files:

    compiledata -pkg staticdata data/*

Assuming this is run from within the `src/myrepo/` directory, you now have:

```
src/
  myrepo/
    data/
      static.html
      main.css
    server/
      main.go
    staticdata/
      file1.go
      file2.go
```

You can build this package in the usual way:

    go build myrepo/staticdata

To use these files, import the `staticdata` package in your main:

```go
import (
   "bitbucket.org/creachadair/filedata"

  _ "myrepo/staticdata"
)
```

The generated package has the side effect of registering the files under their
original paths, and you can open them by using `filedata.Open`:

```go
f, err := filedata.Open("data/main.css")
if err != nil {
  // ...
}
```

## Using "go generate"

You can also use `compiledata` with the `go generate` subcommand. Going back to
the original layout, add a `staticdata/gen.go` file like:

```go
package staticdata
//go:generate compiledata -pkg staticdata -here ../data/*
```

Now you can run

    go generate myrepo/staticdata

and get the same effect.
