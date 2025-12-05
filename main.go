package main

import (
	"fmt"
	"os"

	"github.com/arafat/please/storage"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var rootCmd = &cobra.Command{
	Use:   "please",
	Short: "Packaged Lightweight Environments As Sandboxed Executions - A Package Manager for MacOS",
	Long: `please - Packaged Lightweight Environments As Sandboxed Executions

	A modern package manager for macOS that leverages Apple's native OCI container
	support to run CLI tools in isolated Linux environments. Each tool runs in its
	own container, eliminating dependency conflicts while providing:

	• Multiple versions side-by-side
	• Reproducible development environments
	• Zero system pollution
	• Automatic version switching per project

	Install tools from curated container images and switch between versions
	seamlessly without affecting your system installation.`,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates the local cache",
	Long:  `Updates the local cache by pulling the latest manifests defined in .please/sources`,
	Run: func(cmd *cobra.Command, args []string) {
		s := storage.New()
		if err := s.Initialize(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing please: %v\n", err)
			os.Exit(1)
		}
		manifestURLs, err := s.LoadSources()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading sources: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Updating cache...")
		s.DownloadManifestFiles(manifestURLs)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
