/*
 * Core Initialization
 */
package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultSourceURL = "https://please.oarafat.workers.dev/manifest-core.tar.gz"
	sourcesFile      = "sources"
	pleaseDir        = ".please"
)

type Storage struct {
	homeDir   string
	pleaseDir string
}

func New() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	return &Storage{
		homeDir:   homeDir,
		pleaseDir: fmt.Sprintf("%s/%s", homeDir, pleaseDir),
	}, nil

}

func (s *Storage) SourcesPath() string {
	return filepath.Join(s.pleaseDir, sourcesFile)
}

func (s *Storage) Initialize() (bool, error) {
	if stat, err := os.Stat(s.pleaseDir); err == nil && stat.IsDir() {
		return false, s.createSources() // guarantee a sources file
	}

	fmt.Println("Initializing please for the first time...")

	if err := os.MkdirAll(s.pleaseDir, 0755); err != nil {
		return false, err
	}

	if err := s.createSources(); err != nil {
		return false, err
	}

	fmt.Printf("✓ Initialized please at %s\n", s.pleaseDir)
	fmt.Printf("✓ Configured default source: %s\n", defaultSourceURL)
	return true, nil
}

func (s *Storage) createSources() error {
	path := s.SourcesPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		content := fmt.Sprintf("# please package sources\n# Add one URL per line\n\n%s\n", defaultSourceURL)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create sources file: %w", err)
		}
	}

	return nil
}
