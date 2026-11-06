/*
author: deadc0de6 (https://github.com/deadc0de6)
Copyright (c) 2024, deadc0de6
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

// baseExt returns the catalog extension, ignoring any compression extension
func baseExt(path string) string {
	if isCompressedPath(path) {
		path = path[:len(path)-len(gzipExt)]
	}
	return filepath.Ext(path)
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
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	br := bufio.NewReader(fd)
	magic, err := br.Peek(len(gzipMagic))
	if err != nil || !bytes.Equal(magic, gzipMagic) {
		// not compressed (or too short to tell), let the decoder handle it
		return &readCloser{Reader: br, closers: []io.Closer{fd}}, nil
	}

	gz, err := gzip.NewReader(br)
	if err != nil {
		_ = fd.Close()
		return nil, err
	}
	return &readCloser{Reader: gz, closers: []io.Closer{gz, fd}}, nil
}

// writeFile writes the catalog, compressing it if the path ends with .gz
func writeFile(path string, content []byte) error {
	if !isCompressedPath(path) {
		return os.WriteFile(path, content, os.ModePerm)
	}

	fd, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	gz := gzip.NewWriter(fd)
	_, err = gz.Write(content)
	return errors.Join(err, gz.Close(), fd.Close())
}
