// Binary compiledata generates Go source text containing encoded file data,
// for use with the bitbucket.org/creachadair/filedata package.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/creachadair/filedata/internal/encoder"
)

var (
	pkgName    = flag.String("pkg", "", "Output package name (required)")
	trimPrefix = flag.String("trim", "", "Trim this prefix from each input path")
	baseOnly   = flag.Bool("baseonly", false, "Use only the base name of each input path")
	addPrefix  = flag.String("add", "", "Add this prefix to each registered path")
	outputDir  = flag.String("dir", "", "Output directory (default is $PWD)")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: %s [options] <file-glob>...

Compile all the input files specified by the given glob patterns into a Go
source package. Each input file is compiled to a separate Go source file.

By default, each file is registered to the filedata package on import under its
original path, less any leading path separators. Use --trim to discard a common
prefix from each input path.  Use --add to prepend a prefix to each registered
name.

To include the compiled data in a program, import the generated package.
File contents may be accessed via the filedata package, for example, if
you ran

    compiledata -pkg staticdata -add path/to/my file.txt

then you could access "file.txt" by writing:

    import "bitbucket.org/creachadair/filedata"
    import "./staticdata"  // or wherever you put the package
    ...
    f, err := filedata.Open("path/to/my/file.txt")

Options:
`, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	switch {
	case flag.NArg() == 0:
		log.Fatal("You must specify at least one input file")
	case *pkgName == "":
		log.Fatal("You must specify an output package name")
	}

	dir := filepath.Join(*outputDir, *pkgName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Error setting up output directory: %v", err)
	}

	for i, arg := range flag.Args() {
		path, err := compileFile(dir, arg, i+1)
		if err != nil {
			log.Fatalf("Error compiling %q: %v", arg, err)
		}
		fmt.Fprintln(os.Stderr, "Compiled", arg, "to", path)
	}
}

// compileFile generates a source file under dir to represent the contents of
// name, using index as a nonce to generate a unique identifier.
func compileFile(dir, name string, index int) (string, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("reading file contents: %v", err)
	}
	packed, err := encoder.Encode(data)
	if err != nil {
		return "", fmt.Errorf("encoding file contents: %v", err)
	}

	trimmed := strings.TrimPrefix(name, *trimPrefix)
	added := filepath.Join(*addPrefix, trimmed)
	if *baseOnly {
		added = filepath.Join(*addPrefix, filepath.Base(trimmed))
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, `package %[1]s
// This file was generated from %[2]q.
// Input size: %[5]d bytes; encoded size: %[6]d bytes.
// DO NOT EDIT

import "bitbucket.org/creachadair/filedata"

func init() { filedata.Register(%[3]q, file%[4]ddata) }

const file%[4]ddata = ""+
`, *pkgName, trimmed, added, index, len(data), len(packed))

	if err := encoder.ToSource(&buf, packed); err != nil {
		return "", err
	}
	code, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("formatting source: %v", err)
	}

	path := filepath.Join(dir, fmt.Sprintf("file%d.go", index))
	return path, ioutil.WriteFile(path, code, 0644)
}
