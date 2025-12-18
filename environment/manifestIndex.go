package environment

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"iter"
	"os"
	"sort"
	"strings"

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
		m.Namespace = manifestDecoder.namespace

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

func (m *ManifestArchive) ExactMatch(name string) (*schema.PackageManifest, error) {
	for iter, err := range m.iterateManifest() {
		if err != nil {
			return nil, fmt.Errorf("iteration failed: %w", err)
		}
		if iter.packageManifest.Name == name {
			return &iter.packageManifest, nil
		}
	}

	return nil, fmt.Errorf("package with name '%s' not found", name)
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

// ScriptHooks contains the pre and post hook scripts
type ScriptHooks struct {
	PreHook  string
	PostHook string
}

// LoadScriptHooksFromManifest reads the package name from JSON and returns
// both prehook and posthook scripts if they exist
func (m *ManifestArchive) LoadScriptHooksFromManifest(packageName string) (*ScriptHooks, error) {
	preHookName := fmt.Sprintf("%s_prehook.sh", packageName)
	postHookName := fmt.Sprintf("%s_posthook.sh", packageName)

	hooks, err := extractScriptsFromTarball(m.Path, preHookName, postHookName)
	if err != nil {
		return nil, err
	}

	return hooks, nil
}

// extractScriptsFromTarball extracts both prehook and posthook scripts from the tarball
func extractScriptsFromTarball(manifestPath, preHookName, postHookName string) (*ScriptHooks, error) {
	file, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tarball: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	hooks := &ScriptHooks{}
	foundCount := 0

	for {
		header, err := tr.Next()
		if err == io.EOF {
			// Finished reading tarball
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Only process regular files in the scripts directory
		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Check if this is in the hooks directory
		// Path will be like "hooks/<pkg>_prehook.sh"
		if strings.HasPrefix(header.Name, "hooks/") {
			fileName := strings.TrimPrefix(header.Name, "hooks/")

			if fileName == preHookName {
				contents, err := io.ReadAll(tr)
				if err != nil {
					return nil, fmt.Errorf("failed to read prehook script: %w", err)
				}
				hooks.PreHook = string(contents)
				foundCount++
			} else if fileName == postHookName {
				contents, err := io.ReadAll(tr)
				if err != nil {
					return nil, fmt.Errorf("failed to read posthook script: %w", err)
				}
				hooks.PostHook = string(contents)
				foundCount++
			}

			// If we found both scripts, we can stop early
			if foundCount == 2 {
				break
			}
		}
	}

	return hooks, nil
}
