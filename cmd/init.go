package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/arafat/please/environment"
	"github.com/arafat/please/schema"
	"github.com/arafat/please/utils"
	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "initializes please for first-time usage",
	Long:  `initializes please for first-time usage`,
	Run: func(cmd *cobra.Command, args []string) {
		e := environment.New()
		if e.IsInitialized() {
			fmt.Println("Please is already initialized.")
			return
		}
		if err := e.Initialize(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing please: %v\n", err)
			os.Exit(1)
		}

		data, err := json.MarshalIndent(schema.NewDefaultBundle(), "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling env.json: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(e.EnvironmentPath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write env.json: %v\n", err)
			os.Exit(1)
		}

		manifestURLs, err := e.LoadSources()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading sources: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Updating cache...")
		e.DownloadManifestFiles(manifestURLs)

		fmt.Printf("Adding %s to $PATH\n", e.BinPath)
		utils.AddToUserPath(e.BinPath)
		fmt.Println("Please close and reopen your terminal to apply the changes.")

		fmt.Printf("✅ Initialized please at %s\n", e.PleasePath)
	},
}
