package storage

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type ManifestDecoder struct {
	file    *os.File
	gzr     *gzip.Reader
	decoder *json.Decoder
}

func (t *ManifestDecoder) Close() error {
	if t.gzr != nil {
		t.gzr.Close()
	}
	if t.file != nil {
		return t.file.Close()
	}
	return nil
}

func NewManifestDecoder(manifestPath string) (*ManifestDecoder, error) {
	// Open the tarball
	file, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tarball: %w", err)
	}

	// Create gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Find the JSON file at the root of the archive
	var jsonFound bool
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			gzr.Close()
			file.Close()
			return nil, fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Check if this is a regular file at the root with .json extension
		if header.Typeflag == tar.TypeReg &&
			!strings.Contains(header.Name, "/") &&
			strings.HasSuffix(header.Name, ".json") {
			jsonFound = true
			break
		}
	}

	if !jsonFound {
		gzr.Close()
		file.Close()
		return nil, fmt.Errorf("no JSON file found at root of tarball")
	}

	// Create a decoder
	decoder := json.NewDecoder(tr)

	// Read opening bracket of array
	token, err := decoder.Token()
	if err != nil {
		gzr.Close()
		file.Close()
		return nil, fmt.Errorf("failed to read opening bracket: %w", err)
	}
	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		gzr.Close()
		file.Close()
		return nil, fmt.Errorf("expected array, got %v", token)
	}

	return &ManifestDecoder{
		file:    file,
		gzr:     gzr,
		decoder: decoder,
	}, nil
}
