package cmd

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/arafat/please/schema"
	"github.com/arafat/please/storage"
	"github.com/spf13/cobra"
)

var fuzzySearch bool = false

func init() {
	SearchCmd.Flags().BoolVar(&fuzzySearch, "fuzzy", true, "Fuzzy search (true, default) or exact match search (false)")
}

var SearchCmd = &cobra.Command{
	Use:   "search [package]",
	Short: "performs a package search from local cache",
	Long:  `performs a package search from local cache`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing package name")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]
		s := storage.New()
		s.Initialize()

		manifestPaths, err := s.GetManifestPaths()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var wg sync.WaitGroup
		results := make(map[string][]schema.PackageManifest)
		var resultsMutex sync.Mutex
		errChan := make(chan error, len(manifestPaths))

		for _, path := range manifestPaths {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()

				archive := storage.NewManifestArchive(path)
				var res []schema.PackageManifest
				var err error
				if fuzzySearch {
					res, err = archive.FuzzySearch(packageName, storage.MaxFuzzySearchResults)
				} else {
					match, err := archive.ExactMatch(packageName)
					if err == nil {
						res = []schema.PackageManifest{*match}
					}
				}

				if err != nil {
					errChan <- fmt.Errorf("manifest %s: %w", p, err)
				}

				resultsMutex.Lock()
				results[archive.Namespace] = append(results[archive.Namespace], res...)
				resultsMutex.Unlock()

			}(path)
		}
		wg.Wait()
		close(errChan)

		var errors []error
		for err := range errChan {
			errors = append(errors, err)
		}

		if len(errors) > 0 {
			fmt.Println("Errors encountered during search operation:")
			for _, err := range errors {
				fmt.Printf("  - %v\n", err)
			}
		}

		// Print out search results
		namespaces := make([]string, 0, len(results))
		for namespace := range results {
			namespaces = append(namespaces, namespace)
		}

		// Sort with "core" first, then alphabetically
		sort.Slice(namespaces, func(i, j int) bool {
			if namespaces[i] == "core" {
				return true
			}
			if namespaces[j] == "core" {
				return false
			}
			return namespaces[i] < namespaces[j]
		})

		// Print in sorted order
		for _, namespace := range namespaces {
			fmt.Printf("Namespace: %s\n", namespace)
			for _, manifest := range results[namespace] {
				fmt.Printf("  - %v\n", manifest.Name)
			}
		}
	},
}
