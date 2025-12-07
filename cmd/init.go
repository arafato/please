package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/arafat/please/schema"
	"github.com/arafat/please/storage"
	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "initializes please for first-time usage",
	Long:  `initializes please for first-time usage`,
	Run: func(cmd *cobra.Command, args []string) {
		s := storage.New()
		if s.IsInitialized() {
			fmt.Println("Please is already initialized.")
			return
		}
		if err := s.Initialize(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing please: %v\n", err)
			os.Exit(1)
		}

		data, err := json.MarshalIndent(schema.NewDefaultEnvironment(), "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling env.json: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(s.EnvironmentPath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write env.json: %v\n", err)
			os.Exit(1)
		}

		manifestURLs, err := s.LoadSources()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading sources: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Updating cache...")
		s.DownloadManifestFiles(manifestURLs)

		fmt.Printf("✅ Initialized please at %s\n", s.PleasePath)
	},
}
