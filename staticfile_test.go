package staticfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/creachadair/staticfile/internal/bits"
)

func reset() {
	registry.Lock()
	defer registry.Unlock()
	registry.data = make(map[string]*fileData)
}

func TestEmptyPathError(t *testing.T) {
	reset()

	err := register("", "whatever")
	if err == nil {
		t.Fatal(`register("") should have failed, but did not`)
	}
}

func TestDuplicatePathError(t *testing.T) {
	reset()

	if err := register("x", "foo"); err != nil {
		t.Errorf(`register("x") #1 unexpectedly failed: %v`, err)
	}
	if err := register("x", "bar"); err == nil {
		t.Error(`register("x") #2 unexpectedly succeeded`)
	}
}

func TestFileOpen(t *testing.T) {
	files := map[string]string{
		"a": "Amy, who fell down the stairs",
		"b": "Basil, assaulted by bears",
		"c": "Clara, who wasted away",
		"d": "Desmond, run down by a sleigh",
	}
	for name, text := range files {
		packed, err := bits.Encode([]byte(text))
		if err != nil {
			t.Fatalf("Encoding %q failed: %v", text, err)
		}
		Register(name, string(packed))
	}

	// Include a "real" file to verify that delegation works.
	const realData = "Ernest, who choked on a peach"
	f, err := ioutil.TempFile("", "real*.txt")
	if err != nil {
		t.Fatalf("Creating temp file: %v", err)
	}
	name := f.Name()
	defer os.Remove(name)
	fmt.Fprint(f, realData)
	f.Close()
	files[name] = realData

	for name := range files {
		base := filepath.Base(name)
		t.Run("Open-"+base, func(t *testing.T) {
			f, err := Open(name)
			if err != nil {
				t.Fatalf("Open(%q) failed: %v", name, err)
			}
			data, err := ioutil.ReadAll(f)
			if err := f.Close(); err != nil {
				t.Errorf("%q.Close() failed: %v", name, err)
			}
			if err != nil {
				t.Errorf("Error reading %q: %v", name, err)
			}
			if got := string(data); got != files[name] {
				t.Errorf("Wrong data for %q: got %q, want %q", name, got, files[name])
			}
		})
		t.Run("ReadFile-"+base, func(t *testing.T) {
			data, err := ReadFile(name)
			if err != nil {
				t.Fatalf("ReadFile(%q) failed: %v", name, err)
			}
			if got := string(data); got != files[name] {
				t.Errorf("Wrong data for %q: got %q, want %q", name, got, files[name])
			}
		})
	}
}
