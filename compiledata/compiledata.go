// Binary compiledata generates Go source text containing encoded file data,
// for use with the github.com/creachadair/staticfile package.
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
	"text/template"

	"github.com/creachadair/staticfile/internal/bits"
)

var (
	pkgName    = flag.String("pkg", "", "Output package name (required)")
	trimPrefix = flag.String("trim", "", "Trim this prefix from each input path")
	addPrefix  = flag.String("add", "", "Join this prefix to each registered path")
	baseOnly   = flag.Bool("base", false, "Take only the base name of each input path")
	outputPath = flag.String("out", "", "Output path (required)")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: %s [options] <file-glob>...

Compile the files specified by the glob patterns into a Go source package.
Each file is registered to the staticfile package on import under its original
path, discarding any leading path separators.

Use -trim to discard a common prefix from each path before registration.
Use -add to join a prefix before each registered path.

The compiled files can be accessed via the staticfile package:

    import "github.com/creachadair/staticfile"
    ...
    f, err := staticfile.Open("path/to/my/file.txt")

Options:
`, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	switch {
	case flag.NArg() == 0:
		log.Fatal("You must specify at least one input glob")
	case *pkgName == "":
		log.Fatal("You must specify an output package name")
	case *outputPath == "":
		log.Fatal("You must specify an output filename")
	}

	// Resolve all the files to be compiled.
	inputs, err := expandGlobs(flag.Args())
	if err != nil {
		log.Fatalf("Error expanding globs: %v", err)
	}

	// Set up the output directory.
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0755); err != nil {
		log.Fatalf("Creating output directory: %v", err)
	}

	if err := compileFiles(inputs); err != nil {
		log.Fatalf("File compilation failed: %v", err)
	}
}

// compileFiles generates a source file containing the contents of each
// specified file registered under its stipulated path.
func compileFiles(paths []string) error {
	type file struct {
		Path string
		Name string
		Data string
		Var  string
		Len  int
	}
	v := struct {
		Pkg   string
		Args  string
		Files []file
	}{Pkg: *pkgName, Args: strings.Join(os.Args[1:], " ")}

	for i, name := range paths {
		data, err := ioutil.ReadFile(name)
		if err != nil {
			return fmt.Errorf("reading file contents: %v", err)
		}
		packed, err := bits.Encode(data)
		if err != nil {
			return fmt.Errorf("encoding file contents: %v", err)
		}
		trimmed := strings.TrimPrefix(name, *trimPrefix)
		added := filepath.Join(*addPrefix, trimmed)
		if *baseOnly {
			added = filepath.Join(*addPrefix, filepath.Base(trimmed))
		}
		var src bytes.Buffer
		if err := bits.ToSource(&src, packed); err != nil {
			return fmt.Errorf("packing file contents: %v", err)
		}
		v.Files = append(v.Files, file{
			Path: trimmed,
			Name: added,
			Var:  fmt.Sprintf("_fileData%d", i+1),
			Data: src.String(),
			Len:  len(data),
		})
	}

	buf := new(bytes.Buffer)
	if err := fileTemplate.Execute(buf, v); err != nil {
		return fmt.Errorf("generating source: %v", err)
	}
	code, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting source: %v", err)
	}

	return ioutil.WriteFile(*outputPath, code, 0644)
}

// expandGlobs returns the paths all matching ordinary files from the specified
// globs. Non-files are silently skipped.
func expandGlobs(globs []string) ([]string, error) {
	var inputs []string
	for _, arg := range flag.Args() {
		match, err := filepath.Glob(arg)
		if err != nil {
			log.Fatalf("Invalid glob pattern %q: %v", arg, err)
		}
		for _, path := range match {
			fi, err := os.Stat(path)
			if err != nil {
				return nil, err
			} else if fi.Mode().IsRegular() {
				inputs = append(inputs, path)
			}
		}
	}
	return inputs, nil
}

const templateSource = `package {{.Pkg}}
// This file was generated by compiledata {{.Args}}
// DO NOT EDIT

import "github.com/creachadair/staticfile"

func init() { {{- range .Files}}
  staticfile.Register("{{.Name}}", {{.Var}}){{end -}}
}

const ({{range .Files}}
// {{.Len}} bytes generated from {{.Path}}
{{.Var}} = ""+
{{.Data}}
{{end}})

// END OF GENERATED DATA`

var fileTemplate = template.Must(template.New("static").Parse(templateSource))
