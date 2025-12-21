package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/arafat/please/artifacts"
	"github.com/arafat/please/environment"
	"github.com/arafat/please/utils"
	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [namespace:package]",
	Short: "Delete an app from the current bundle, default namespace is 'core'",
	Long:  "Delete an app from the current bundle, default namespace is 'core'",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing package name")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		pkg := args[0]
		e := environment.New()
		e.Initialize()

		ma := environment.NewManifestArchive(e.ManifestCoreFile)
		pm, err := ma.ExactMatch(pkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding package:%v", err)
			return
		}

		bundle, err := environment.LoadBundleDefinitions(e)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading bundle definitions:%v", err)
			return
		}

		version, err := bundle.GetPackageVersion(pkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting package version:%v", err)
			return
		}

		replacer := utils.MakeRuntimeReplacer(version)
		replacer(pm.ContainerArgs.ContainerEnvVars)
		replacer(pm.HostEnvVars)

		e.DeleteSymlink(pm.Exec)
		e.DeleteArtifact(pkg, version)

		// Delete the package from the bundle
		if err := bundle.DeletePackage(pkg); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting package:%v", err)
			return
		}

		// Save the updated bundle
		if err := bundle.SaveBundle(e); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving bundle:%v", err)
			return
		}

		hooks, err := ma.LoadScriptHooksFromManifest(pkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading script hooks: %v\n", err)
			return
		}

		preHook := artifacts.NewShellHook(hooks.PostHook, pm.HostEnvVars)
		if err := preHook.Execute(context.TODO()); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing post-hook: %v\n", err)
			return
		}

		fmt.Printf("✅ Package '%s' deleted successfully from bundle [%s]\n", pkg, bundle.GetActiveBundle())
	},
}
