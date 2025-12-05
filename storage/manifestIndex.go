package storage

import (
	"fmt"
	"iter"
	"sort"

	"github.com/agnivade/levenshtein"

	"github.com/arafat/please/schema"
)

type manifestIterator struct {
	packageManifest schema.PackageManifest
}

type ManifestArchive struct {
	Path      string
	Namespace string
	Count     int
}

func NewManifestArchive(path string) *ManifestArchive {
	m := &ManifestArchive{
		Path:  path,
		Count: 0,
	}
	for range m.iterateManifest() {
		m.Count++
	}

	return m
}

// Iterates through entire manifest (zipped tarball) with streaming json
func (m *ManifestArchive) iterateManifest() iter.Seq2[manifestIterator, error] {
	return func(yield func(manifestIterator, error) bool) {
		manifestDecoder, err := NewManifestDecoder(m.Path)
		if err != nil {
			yield(manifestIterator{}, err)
		}
		defer manifestDecoder.Close()

		// Iterate through objects
		for manifestDecoder.decoder.More() {
			var obj schema.PackageManifest
			if err := manifestDecoder.decoder.Decode(&obj); err != nil {
				yield(manifestIterator{}, fmt.Errorf("failed to decode object: %w", err))
				return
			}

			if !yield(manifestIterator{packageManifest: obj}, nil) {
				return
			}
		}
	}
}

const MaxFuzzySearchResults = 10

func (m *ManifestArchive) Lookup(name string) (*schema.PackageManifest, error) {
	for iter, err := range m.iterateManifest() {
		if err != nil {
			return nil, fmt.Errorf("iteration failed: %w", err)
		}
		if iter.packageManifest.Name == name {
			return &iter.packageManifest, nil
		}
	}

	return nil, fmt.Errorf("object with name '%s' not found at expected offset %d", name, offset)
}

type candidate struct {
	manifest schema.PackageManifest
	distance int
}

func (m *ManifestArchive) getFuzzyCandidates(query string, maxResults int) []candidate {
	length := len(query)
	maxDistance := (length * 3) / 10
	if maxDistance < 1 {
		maxDistance = 1
	}

	var candidates []candidate
	for iter, _ := range m.iterateManifest() {
		if len(candidates) >= maxResults {
			break
		}

		distance := levenshtein.ComputeDistance(query, iter.packageManifest.Name)
		if distance <= maxDistance {
			candidates = append(candidates, candidate{
				manifest: iter.packageManifest,
				distance: distance,
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
func (m *ManifestArchive) FuzzySearch(query string, maxResults int) ([]schema.PackageManifest, error) {
	candidates := m.getFuzzyCandidates(query, maxResults)

	if len(candidates) == 0 {
		return []schema.PackageManifest{}, nil
	}
	manifests := make([]schema.PackageManifest, 0, len(candidates))
	for _, c := range candidates {
		manifests = append(manifests, c.manifest)
	}

	return manifests, nil
}
