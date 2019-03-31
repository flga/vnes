package asset

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type List []*asset

func (a List) Open(path string) (io.ReadCloser, error) {
	for i := 0; i < len(a); i++ {
		if a[i].path == path {
			return a[i], nil
		}
	}

	return nil, &os.PathError{
		Op:   "open",
		Path: path,
		Err:  os.ErrNotExist,
	}
}

type asset struct {
	path string
	data string

	buf        *gzip.Reader
	err        error
	decodeOnce sync.Once
}

func New(args ...string) *asset {
	a := &asset{
		path: filepath.Join(args[:len(args)-1]...),
		data: args[len(args)-1],
	}
	return a
}

func Encode(data []byte) (string, error) {
	buf := &bytes.Buffer{}
	encoder := base64.NewEncoder(base64.StdEncoding, buf)

	w := gzip.NewWriter(encoder)
	if _, err := w.Write(data); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	if err := encoder.Close(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (a *asset) decode() {
	decoder := base64.NewDecoder(base64.RawStdEncoding, strings.NewReader(a.data))
	reader, err := gzip.NewReader(decoder)
	if err != nil {
		a.err = err
		return
	}
	a.buf = reader
}

func (a *asset) Close() error {
	return a.buf.Close()
}

func (a *asset) Read(p []byte) (n int, err error) {
	a.decodeOnce.Do(a.decode)
	if a.err != nil {
		return 0, err
	}

	return a.buf.Read(p)
}
