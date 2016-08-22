package filedata

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
func Register(path, data string) {
	registry.Lock()
	defer registry.Unlock()

	clean := strings.TrimLeft(filepath.Clean(path), string(filepath.Separator))
	if path == "" {
		log.Panic("filedata: registered empty path")
	} else if _, ok := registry.data[clean]; ok {
		log.Panicf("filedata: duplicate path registered: %q", clean)
	}
	registry.data[clean] = &fileData{
		Path:    clean,
		Data:    []byte(data),
		Decoded: false,
	}
}

// A File is a read-only view of the contents of a static file.
// It implements io.Reader, io.ReaderAt, io.Seeker, and io.WriterTo.
type File struct{ data *bytes.Reader }

// Close implements io.Closer. This implementation never returns an error.
func (*File) Close() error { return nil }

// Size reports the total unencoded size of the file contents, in bytes.
func (f *File) Size() int64 { return f.data.Size() }

func (f *File) Read(data []byte) (int, error)              { return f.data.Read(data) }
func (f *File) ReadAt(data []byte, off int64) (int, error) { return f.data.ReadAt(data, off) }
func (f *File) Seek(off int64, whence int) (int64, error)  { return f.data.Seek(off, whence) }
func (f *File) WriteTo(w io.Writer) (int64, error)         { return f.data.WriteTo(w) }

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
		dec, err := decode(d.Data)
		if err != nil {
			return nil, err
		}
		d.Data = dec
		d.Decoded = true
	}

	return &File{bytes.NewReader(d.Data)}, nil
}

// decode returns the raw file contents denoted by data.
func decode(data []byte) ([]byte, error) {
	rc, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("filedata: decoding error: %v", err)
	}
	defer rc.Close()
	return ioutil.ReadAll(rc)
}