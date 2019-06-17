package staticfile

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
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
// by filepath.Clean, and any leading path separators are discarded.
// This function will panic if path == "" or if the cleaned path has previously
// been registered.
//
// This function is meant to be used from generated code, and should not
// ordinarily be called directly by clients of the library.
func Register(path, data string) {
	if err := register(path, data); err != nil {
		log.Panic(err)
	}
}

func register(path, data string) error {
	registry.Lock()
	defer registry.Unlock()

	clean := strings.TrimLeft(filepath.Clean(path), string(filepath.Separator))
	if path == "" {
		return errors.New("filedata: registered empty path")
	} else if _, ok := registry.data[clean]; ok {
		return fmt.Errorf("filedata: duplicate path registered: %q", clean)
	}
	registry.data[clean] = &fileData{
		Path:    clean,
		Data:    []byte(data),
		Decoded: false,
	}
	return nil
}

// A File is a read-only view of the contents of a static file.
// It implements io.Reader, io.ReaderAt, io.Seeker, and io.WriterTo.
type File struct{ data *bytes.Reader }

// Close implements io.Closer. This implementation never returns an error.
func (*File) Close() error { return nil }

// Size reports the total unencoded size of the file contents, in bytes.
func (f *File) Size() int64 { return f.data.Size() }

// Read implements the io.Reader interface.
func (f *File) Read(data []byte) (int, error) { return f.data.Read(data) }

// ReadAt implements the io.ReaderAt interface.
func (f *File) ReadAt(data []byte, off int64) (int, error) { return f.data.ReadAt(data, off) }

// Seek implements the io.Seeker interface.
func (f *File) Seek(off int64, whence int) (int64, error) { return f.data.Seek(off, whence) }

// WriteTo implements the io.WriterTo interface
func (f *File) WriteTo(w io.Writer) (int64, error) { return f.data.WriteTo(w) }

// Open opens a static file given its clean registered path.
// Returns io.ErrNotExist if no such path is registered.
func Open(path string) (*File, error) {
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

	return &File{bytes.NewReader(d.Data)}, nil
}
