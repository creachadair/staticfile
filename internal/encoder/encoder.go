// Package encoder provides support routines for encoding and decoding data.
// It is part of the filedata compiler.
package encoder

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
)

// Encode converts raw file contents into an encoded form.
func Encode(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(data); err != nil {
		return nil, fmt.Errorf("filepath: encoding error: %v", err)
	} else if err := w.Close(); err != nil {
		return nil, fmt.Errorf("filepath: encoding error: %v", err)
	}
	return buf.Bytes(), nil
}

// ToSource renders the given data as Go source text, which is written to w.
func ToSource(w io.Writer, data []byte) error {
	buf := bufio.NewWriter(w)
	buf.WriteByte('"')

	const maxWidth = 100
	pos := 1
	for _, b := range data {
		if b < ' ' || b > '~' {
			fmt.Fprintf(buf, `\x%02x`, b)
			pos += 4
		} else if b == '"' || b == '\\' {
			buf.WriteByte('\\')
			buf.WriteByte(b)
			pos += 2
		} else {
			buf.WriteByte(b)
			pos++
		}
		if pos > maxWidth {
			fmt.Fprint(buf, "\"+\n\"")
			pos = 1
		}
	}
	buf.WriteByte('"')
	if pos > 1 {
		buf.WriteByte('\n')
	}
	return buf.Flush()
}
