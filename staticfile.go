package staticfile

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/creachadair/staticfile/internal/bits"
)

var registry = struct {
	sync.Mutex
	data map[string]*fileData
}{data: make(map[string]*fileData)}

type fileData struct {
	Path    string
	Data    []byte
	Decoded bool
}

// Register the contents of a file under the given path.  The path is cleaned
// by filepath.Clean.  This function will panic if path == "" or if the cleaned
// path has previously been registered.
//
// This function is meant to be used from generated code, and should not
// ordinarily be called directly by clients of the library.
func Register(path, data string) {
	if err := register(path, data); err != nil {
		log.Panic(err)
	}
}

func register(path, data string) error {
	if path == "" {
		return errors.New("filedata: registered empty path")
	}

	registry.Lock()
	defer registry.Unlock()
	clean := filepath.Clean(path)
	if _, ok := registry.data[clean]; ok {
		return fmt.Errorf("filedata: duplicate path registered: %q", clean)
	}
	registry.data[clean] = &fileData{
		Path:    clean,
		Data:    []byte(data),
		Decoded: false,
	}
	return nil
}

// View is a read-only view of the contents of a static file.  It implements
// the File interface.
type View struct{ data *bytes.Reader }

// Close implements io.Closer. This implementation never returns an error, and
// no resources are leaked if a *View is not closed.
func (*View) Close() error { return nil }

// Size reports the total unencoded size of the file contents, in bytes.
func (v *View) Size() int64 { return v.data.Size() }

// Read implements the io.Reader interface.
func (v *View) Read(data []byte) (int, error) { return v.data.Read(data) }

// ReadAt implements the io.ReaderAt interface.
func (v *View) ReadAt(data []byte, off int64) (int, error) { return v.data.ReadAt(data, off) }

// Seek implements the io.Seeker interface.
func (v *View) Seek(off int64, whence int) (int64, error) { return v.data.Seek(off, whence) }

// File is the interface satisfied by files opened by the Open function.  It is
// satisfied by the *os.File and *View types.
type File interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}

// Open opens the specified file path for reading. If path is not a registered
// static file path, Open delegates to os.Open. Otherwise, the concrete type of
// the result is *View.
func Open(path string) (File, error) {
	data, err := openData(path)
	if err == os.ErrNotExist {
		return os.Open(path)
	} else if err != nil {
		return nil, err
	}
	return &View{bytes.NewReader(data)}, nil
}

// ReadFile reads the complete contents of the specified file path. If path is
// not a registered static file path, ReadFile delegates to ioutil.ReadFile.
func ReadFile(path string) ([]byte, error) {
	data, err := openData(path)
	if err == os.ErrNotExist {
		return ioutil.ReadFile(path)
	} else if err != nil {
		return nil, err
	}
	return data, nil
}

func openData(path string) ([]byte, error) {
	registry.Lock()
	defer registry.Unlock()

	d, ok := registry.data[path]
	if !ok {
		return nil, os.ErrNotExist
	}

	// The first time we open a file, decode the bits.
	if !d.Decoded {
		dec, err := bits.Decode(d.Data)
		if err != nil {
			return nil, err
		}
		d.Data = dec
		d.Decoded = true
	}

	return d.Data, nil
}

// MustReadFile returns the full content of the specified static file or
// panics.  It is intended for use during program initialization.  Unlike
// ReadFile, this function does not delegate to the real filesystem.
func MustReadFile(path string) []byte {
	data, err := openData(path)
	if err != nil {
		log.Panicf("reading %q: %v", path, err)
	}
	return data
}
