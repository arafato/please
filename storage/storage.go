/*
 * Core Initialization
 */
package storage

import (
	"fmt"
	"os"
)

const (
	defaultSourceURL = "https://please.oarafat.workers.dev/manifest-core.tar.gz"
	sourcesFile      = "sources"
	pleaseDir        = ".please"
)

func Initialize() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	pleaseDirPath := fmt.Sprintf("%s/%s", homeDir, pleaseDir)
	err = os.Mkdir(pleaseDirPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create .please directory: %w", err)
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", pleaseDirPath, sourcesFile))
	if err != nil {
		return fmt.Errorf("failed to create sources file: %w", err)
	}

	_, err = file.Write([]byte(defaultSourceURL))
	if err != nil {
		return fmt.Errorf("failed to write to sources file: %w", err)
	}

	return nil
}
