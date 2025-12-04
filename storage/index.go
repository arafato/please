package storage

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type StartOffset int64
type NameIndex map[string]StartOffset

// BuildIndex creates an index mapping object names to their byte offsets in the JSON array
func BuildIndex(inputPath string) (NameIndex, error) {
	// Open the tarball
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tarball: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Read the first (and only) file in the tarball
	_, err = tr.Next()
	if err != nil {
		return nil, fmt.Errorf("failed to read tar entry: %w", err)
	}

	return buildIndexFromReader(tr)
}

// buildIndexFromReader builds the index from any io.Reader containing JSON array
func buildIndexFromReader(r io.Reader) (NameIndex, error) {
	index := make(NameIndex)
	decoder := json.NewDecoder(r)

	// Read opening bracket of array
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to read opening bracket: %w", err)
	}
	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		return nil, fmt.Errorf("expected array, got %v", token)
	}

	// Iterate through array elements
	for decoder.More() {
		// Record the offset before decoding the next object
		// decoder.InputOffset() gives us the position in the stream
		offset := decoder.InputOffset()

		// Decode just enough to get the name field
		var obj struct {
			Name string `json:"name"`
		}

		if err := decoder.Decode(&obj); err != nil {
			return nil, fmt.Errorf("failed to decode object: %w", err)
		}

		if obj.Name == "" {
			return nil, fmt.Errorf("object at offset %d has no name field", offset)
		}

		// Store the offset where this object begins
		index[obj.Name] = StartOffset(offset)
	}

	// Read closing bracket
	token, err = decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to read closing bracket: %w", err)
	}
	if delim, ok := token.(json.Delim); !ok || delim != ']' {
		return nil, fmt.Errorf("expected closing bracket, got %v", token)
	}

	return index, nil
}

// SaveIndex saves the index to a file
func SaveIndex(index NameIndex, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create index file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(index); err != nil {
		return fmt.Errorf("failed to encode index: %w", err)
	}

	return nil
}

// LoadIndex loads the index from a file
func LoadIndex(inputPath string) (NameIndex, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open index file: %w", err)
	}
	defer file.Close()

	var index NameIndex
	decoder := json.NewDecoder(file)

	if err := decoder.Decode(&index); err != nil {
		return nil, fmt.Errorf("failed to decode index: %w", err)
	}

	return index, nil
}
