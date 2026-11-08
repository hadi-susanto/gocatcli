/*
Fork-only file (main-ex branch of hadi-susanto/gocatcli-ex).
Gzip-compressed catalogs. Upstream backends only swap a call:
os.WriteFile -> writeFile, os.Open -> openReader,
toml.DecodeFile -> decodeTOMLFile, filepath.Ext(p) -> filepath.Ext(trimCompressExt(p)).
*/

package catalog

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/deadc0de6/gocatcli/internal/log"

	"github.com/BurntSushi/toml"
)

const (
	gzipExt = ".gz"
)

var (
	gzipMagic = []byte{0x1f, 0x8b}
)

// isCompressedPath returns true if the catalog should be saved compressed
func isCompressedPath(path string) bool {
	return strings.EqualFold(filepath.Ext(path), gzipExt)
}

// trimCompressExt removes the compression extension, if any,
// so the backend is picked from the extension before it
func trimCompressExt(path string) string {
	if isCompressedPath(path) {
		return path[:len(path)-len(gzipExt)]
	}
	return path
}

// readCloser closes the decompressor and the underlying file
type readCloser struct {
	io.Reader
	closers []io.Closer
}

func (r *readCloser) Close() error {
	var errs []error
	for _, c := range r.closers {
		errs = append(errs, c.Close())
	}
	return errors.Join(errs...)
}

// openReader opens the catalog, transparently decompressing it if gzipped
func openReader(path string) (io.ReadCloser, error) {
	var fd, err = os.Open(path)
	if err != nil {
		return nil, err
	}

	var br = bufio.NewReader(fd)
	var magic []byte
	magic, err = br.Peek(len(gzipMagic))
	if err != nil || !bytes.Equal(magic, gzipMagic) {
		// not compressed (or too short to tell), let the decoder handle it
		return &readCloser{Reader: br, closers: []io.Closer{fd}}, nil
	}

	var gz *gzip.Reader
	gz, err = gzip.NewReader(br)
	if err != nil {
		_ = fd.Close()
		return nil, err
	}
	return &readCloser{Reader: gz, closers: []io.Closer{gz, fd}}, nil
}

// writeFile is os.WriteFile, compressing when the path ends with .gz
func writeFile(path string, content []byte, perm os.FileMode) error {
	if !isCompressedPath(path) {
		return os.WriteFile(path, content, perm)
	}

	var fd, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	var gz = gzip.NewWriter(fd)
	_, err = gz.Write(content)
	return errors.Join(err, gz.Close(), fd.Close())
}

// decodeTOMLFile is toml.DecodeFile, transparently decompressing
func decodeTOMLFile(path string, v any) (toml.MetaData, error) {
	var fd, err = openReader(path)
	if err != nil {
		return toml.MetaData{}, err
	}
	defer func() {
		var cerr = fd.Close()
		if cerr != nil {
			log.Error(cerr)
		}
	}()
	return toml.NewDecoder(fd).Decode(v)
}
