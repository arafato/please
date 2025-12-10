package cmd

import (
	"fmt"
	"os"

	"github.com/arafat/please/environment"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "init" {
			return nil
		}
		s := environment.New()
		if s.IsInitialized() {
			return nil
		}

		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		red := color.New(color.FgRed, color.Bold).SprintFunc()
		yellow := color.New(color.FgHiYellow).SprintFunc()
		fmt.Printf("%s %s\n", red("❌"), red("please is not initialized yet."))
		fmt.Printf("%s %s\n", yellow("💡"), yellow("Run: please init"))

		os.Exit(1)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(UpdateCmd)
	RootCmd.AddCommand(SearchCmd)
	RootCmd.AddCommand(InstallCmd)
	RootCmd.AddCommand(InitCmd)
}
