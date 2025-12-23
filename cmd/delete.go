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

func init() {
	DeleteCmd.AddCommand(deleteBundleCmd)
	DeleteCmd.AddCommand(deletePackageCmd)
}

var deleteBundleCmd = &cobra.Command{
	Use:   "bundle <bundlename>",
	Short: "Delete the bundle",
	Long:  "Delete the bundle <bundlename> (must not be active)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		env := environment.New()
		if !env.IsInitialized() {
			fmt.Printf("Please has not been initialized yet. Run please init.")
			return
		}

		bundleName := args[0]

		bDefs, err := environment.LoadBundleDefinitions(env)
		if err != nil {
			fmt.Printf("Error loading bundle definitions: %v\n", err)
			return
		}

		if err := bDefs.DeleteBundle(bundleName); err != nil {
			fmt.Printf("Error deleting bundle %q: %v\n", bundleName, err)
			return
		}

		if err := bDefs.SaveBundle(env); err != nil {
			fmt.Printf("Error saving bundle definitions: %v\n", err)
			return
		}

		fmt.Printf("✅ Bundle %q deleted successfully.\n", bundleName)
	},
}

var deletePackageCmd = &cobra.Command{
	Use:   "package <pkg>",
	Short: "Delete the package",
	Long:  "Delete the package <pkg> from the currently active bundle",
	Args:  cobra.MaximumNArgs(1),
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

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a bundle or a package",
	Long:  "Delete a bundle or a package",
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
