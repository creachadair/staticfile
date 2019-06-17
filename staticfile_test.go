package staticfile

import (
	"io/ioutil"
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

	for name := range files {
		f, err := Open(name)
		if err != nil {
			t.Errorf("Open(%q) failed: %v", name, err)
			continue
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
	}
}
