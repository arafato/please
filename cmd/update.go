package cmd

import (
	"fmt"
	"os"

	"github.com/arafat/please/storage"
	"github.com/spf13/cobra"
)

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates the local cache",
	Long:  `Updates the local cache by pulling the latest manifests defined in .please/sources`,
	Run: func(cmd *cobra.Command, args []string) {
		s := storage.New()
		manifestURLs, err := s.LoadSources()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading sources: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Updating cache...")
		s.DownloadManifestFiles(manifestURLs)
	},
}
