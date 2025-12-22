package cmd

import (
	"fmt"
	"os"

	"github.com/arafat/please/utils/buildinfo"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the current Please version.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := buildinfo.PrintVersion(os.Stdout); err != nil {
			fmt.Printf("%v", err)
		}
		fmt.Print("\n")
	},
}
