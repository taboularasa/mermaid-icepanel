// Package parser provides functionality to parse Mermaid diagrams.
package parser

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ErrEmptyPath is returned when an empty file path is provided.
var ErrEmptyPath = errors.New("empty file path")

// ErrInvalidPath is returned when a suspicious file path is provided.
var ErrInvalidPath = errors.New("invalid or suspicious file path")

// OsFileReader reads files from the filesystem using os package.
type OsFileReader struct{}

// ReadFile implements FileReader interface.
func (r *OsFileReader) ReadFile(path string) (ReadCloser, error) {
	// Validate file path
	if path == "" {
		return nil, ErrEmptyPath
	}

	// Prevent path traversal attacks by removing ".." and ensuring path is clean
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return nil, ErrInvalidPath
	}

	// Only allow files with .mmd extension
	if filepath.Ext(cleanPath) != ".mmd" {
		return nil, ErrInvalidPath
	}

	return os.Open(cleanPath)
}

// ReadCloser combines io.Reader and io.Closer.
type ReadCloser interface {
	Read(p []byte) (n int, err error)
	Close() error
}
