package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func AddToUserPath(newPath string) error {
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	// Convert absolute path to use $HOME if it starts with the user's home directory
	relativePath := newPath
	if strings.HasPrefix(newPath, usr.HomeDir+"/") || newPath == usr.HomeDir {
		relativePath = "$HOME" + strings.TrimPrefix(newPath, usr.HomeDir)
	}

	// Helper function to append export line if not already present
	addPathToFile := func(rcFile string) error {
		f, err := os.OpenFile(rcFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", rcFile, err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			// Check for both absolute and $HOME versions to avoid duplicates
			if strings.Contains(line, newPath) || strings.Contains(line, relativePath) {
				// Path already added
				return nil
			}
		}

		// Append the export line using $HOME-based path
		if _, err := f.WriteString(fmt.Sprintf("\n# added by please\nexport PATH=%s:$PATH\n", relativePath)); err != nil {
			return fmt.Errorf("failed to write to %s: %w", rcFile, err)
		}
		return nil
	}

	// Update ~/.profile for login shells
	profileFile := filepath.Join(usr.HomeDir, ".profile")
	if err := addPathToFile(profileFile); err != nil {
		return err
	}

	// Update shell-specific RC for interactive shells
	shell := os.Getenv("SHELL")
	switch {
	case strings.HasSuffix(shell, "bash"):
		bashRC := filepath.Join(usr.HomeDir, ".bashrc")
		if err := addPathToFile(bashRC); err != nil {
			return err
		}
	case strings.HasSuffix(shell, "zsh"):
		zshRC := filepath.Join(usr.HomeDir, ".zshrc")
		if err := addPathToFile(zshRC); err != nil {
			return err
		}
	}

	return nil
}
