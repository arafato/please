package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/arafat/please/environment"
	"github.com/spf13/cobra"
)

func init() {
	ShowCmd.AddCommand(showBundleCmd)
	ShowCmd.AddCommand(showPackageCmd)
}

var ShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show information about the current project",
	Long:  "Show information about bundles or packages in the current project",
}

var showBundleCmd = &cobra.Command{
	Use:   "bundle <bundlename>",
	Short: "Show information about bundles",
	Long:  "Show all bundles or details about a specific bundle",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		env := environment.New()
		if !env.IsInitialized() {
			fmt.Printf("Please has not been initialized yet. Run please init.")
			return
		}

		bDefs, err := environment.LoadBundleDefinitions(env)
		if err != nil {
			fmt.Printf("Error loading bundle definitions: %v\n", err)
			return
		}

		if len(args) == 0 {
			bundles := bDefs.ListBundles()
			fmt.Printf("%d available bundle(s): \n", len(bundles))
			for _, bundle := range bundles {
				fmt.Printf("- %s\n", bundle)
			}
		} else {
			bundleName := args[0]
			packages := bDefs.GetInstalledPackages(bundleName)
			fmt.Printf("%d installed package(s) in bundle [%s]\n", len(packages), bundleName)
			for pkg, version := range packages {
				fmt.Printf("- %s, Version: %s\n", pkg, version)
			}
		}
	},
}

// Subcommand for showing packages
var showPackageCmd = &cobra.Command{
	Use:   "package <packagename>",
	Short: "Show information about a specific package",
	Args:  cobra.ExactArgs(1), // Require exactly one argument
	Run: func(cmd *cobra.Command, args []string) {
		env := environment.New()
		if !env.IsInitialized() {
			fmt.Printf("Please has not been initialized yet. Run please init.")
			return
		}

		pkg := args[0]

		fmt.Printf("Package information: %s\n", pkg)
		ma := environment.NewManifestArchive(env.ManifestCoreFile)
		pm, err := ma.ExactMatch(pkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding package:%v", err)
			return
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "Name:\t%s\n", pm.Name)
		fmt.Fprintf(w, "Description:\t%s\n", pm.Description)
		fmt.Fprintf(w, "Homepage:\t%s\n", pm.Homepage)
		fmt.Fprintf(w, "License:\t%s\n", pm.License)
		fmt.Fprintf(w, "Categories:\t%s\n", strings.Join(pm.Categories, ", "))
		fmt.Fprintf(w, "Image:\t%s\n", pm.Image)
		fmt.Fprintf(w, "Platforms:\t%s\n", strings.Join(pm.Platforms, ", "))
		if len(pm.Versions) > 0 {
			fmt.Fprintf(w, "Versions:\t%s\n", strings.Join(pm.Versions, ", "))
		} else {
			fmt.Fprintf(w, "Versions:\tauto-discover\n")
		}
		fmt.Fprintf(w, "Platforms:\t%s\n", strings.Join(pm.Platforms, ", "))

		w.Flush()
	},
}
