package storage

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/agnivade/levenshtein"

	"github.com/arafat/please/schema"
)

type StartOffset int64
type NameIndex map[string]StartOffset

type manifestIterator struct {
	offset          StartOffset
	packageManifest schema.PackageManifest
}

func iterateManifest(manifestPath string) func(yield func(manifestIterator, error) bool) {
	return func(yield func(manifestIterator, error) bool) {
		manifestDecoder, err := NewManifestDecoder(manifestPath)
		if err != nil {
			yield(manifestIterator{}, err)
		}
		defer manifestDecoder.Close()

		// Iterate through objects
		for manifestDecoder.decoder.More() {
			offset := StartOffset(manifestDecoder.decoder.InputOffset())

			var obj schema.PackageManifest
			if err := manifestDecoder.decoder.Decode(&obj); err != nil {
				yield(manifestIterator{}, fmt.Errorf("failed to decode object: %w", err))
				return
			}

			if !yield(manifestIterator{offset: offset, packageManifest: obj}, nil) {
				return
			}
		}
	}
}

const maxFuzzySearchResults = 10

type FuzzySearchResult struct {
	Name     string
	Distance int
	Manifest map[string]interface{}
}

// BuildIndex creates an index mapping object names to their byte offsets in the JSON array
func BuildIndex(manifestPath string) (NameIndex, error) {

	index := make(NameIndex)
	for iter, err := range iterateManifest(manifestPath) {
		if err != nil {
			return nil, fmt.Errorf("iteration failed: %w", err)
		}
		index[iter.packageManifest.Name] = iter.offset
	}
	return index, nil
}

func LookupObject(tarballPath string, index NameIndex, name string) (*schema.PackageManifest, error) {
	// Check if the name exists in the index
	offset, exists := index[name]
	if !exists {
		return nil, fmt.Errorf("object with name '%s' not found in index", name)
	}

	// Open the tarball
	file, err := os.Open(tarballPath)
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

	// Create a decoder and seek to the offset
	decoder := json.NewDecoder(tr)

	// Read opening bracket of array
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to read opening bracket: %w", err)
	}
	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		return nil, fmt.Errorf("expected array, got %v", token)
	}

	// Iterate through objects until we reach the target offset
	for decoder.More() {
		currentOffset := decoder.InputOffset()

		// Decode the object
		var obj schema.PackageManifest
		if err := decoder.Decode(&obj); err != nil {
			return nil, fmt.Errorf("failed to decode object: %w", err)
		}

		// Check if this is the object we're looking for
		if StartOffset(currentOffset) == offset {
			return &obj, nil
		}
	}

	return nil, fmt.Errorf("object with name '%s' not found at expected offset %d", name, offset)
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

type candidate struct {
	name     string
	distance int
	offset   StartOffset
}

func getFuzzyCandidates(index NameIndex, query string) []candidate {
	length := len(query)
	maxDistance := (length * 3) / 10 // 30%, integer division auto-floors
	if maxDistance < 1 {
		maxDistance = 1
	}

	var candidates []candidate
	for name, offset := range index {
		if len(candidates) >= maxFuzzySearchResults {
			break
		}
		distance := levenshtein.ComputeDistance(query, name)
		if distance <= maxDistance {
			candidates = append(candidates, candidate{
				name:     name,
				distance: distance,
				offset:   offset,
			})
		}
	}

	// Best matches first
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].distance < candidates[j].distance
	})
	return candidates
}

// FuzzySearch performs fuzzy search on object names and returns matching objects
func FuzzySearch(manifestPath string, index NameIndex, query string) ([]FuzzySearchResult, error) {

	candidates := getFuzzyCandidates(index, query)
	// If no candidates found, return empty result
	if len(candidates) == 0 {
		return []FuzzySearchResult{}, nil
	}

	// Load only the matching objects from manifest (tarball)
	results := make([]FuzzySearchResult, 0, len(candidates))

	targetOffsets := make(map[StartOffset]candidate)
	for _, c := range candidates {
		targetOffsets[c.offset] = c
	}

	// Create a decoder
	decoder, cleanup, err := createDecoder(manifestPath)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	// Read opening bracket of array
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to read opening bracket: %w", err)
	}
	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		return nil, fmt.Errorf("expected array, got %v", token)
	}

	// Iterate through objects and collect matches
	for decoder.More() {
		currentOffset := StartOffset(decoder.InputOffset())

		// Decode the object
		var obj map[string]interface{}
		if err := decoder.Decode(&obj); err != nil {
			return nil, fmt.Errorf("failed to decode object: %w", err)
		}

		// Check if this offset is one we're looking for
		if c, found := targetOffsets[currentOffset]; found {
			results = append(results, FuzzySearchResult{
				Name:     c.name,
				Distance: c.distance,
				Object:   obj,
			})

			// Remove from map so we know when we've found everything
			delete(targetOffsets, currentOffset)

			// If we've found all matches, we can stop early
			if len(targetOffsets) == 0 {
				break
			}
		}
	}

	// Sort results by distance again (in case they were loaded out of order)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Distance < results[i].Distance {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results, nil
}

func createDecoder(path string) (*json.Decoder, func(), error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open tarball: %w", err)
	}

	gzr, err := gzip.NewReader(file)
	if err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	tr := tar.NewReader(gzr)

	// Read the first (and only) file in the tarball
	_, err = tr.Next()
	if err != nil {
		gzr.Close()
		file.Close()
		return nil, nil, fmt.Errorf("failed to read tar entry: %w", err)
	}

	decoder := json.NewDecoder(tr)

	cleanup := func() {
		gzr.Close()
		file.Close()
	}

	return decoder, cleanup, nil
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
