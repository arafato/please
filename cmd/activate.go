package cmd

import (
	"fmt"

	"github.com/arafat/please/environment"
	"github.com/spf13/cobra"
)

var ActivateCmd = &cobra.Command{
	Use:   "activate <bundle>",
	Short: "activate bundle",
	Long:  "activate bundle",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Usage: please activate <bundle>")
			return
		}

		bundleName := args[0]

		env := environment.New()
		if !env.IsInitialized() {
			fmt.Println("Please initialize the tool first with the init command.")
			return
		}

		bDefs, err := environment.LoadBundleDefinitions(env)
		if err != nil {
			fmt.Printf("Error loading bundle definitions: %v", err)
		}

		if !bDefs.BundleExists(bundleName) {
			fmt.Printf("Bundle %q does not exist\n", bundleName)
			return
		}

		if err := cleanupCurrentBundle(env, bDefs); err != nil {
			fmt.Printf("Error cleaning up current bundle: %v\n", err)
		}

		bundle, err := activateBundle(bundleName, env)
		if err != nil {
			fmt.Printf("Error activating bundle %q: %v\n", bundleName, err)
			return
		}

		if err := bundle.SaveBundle(env); err != nil {
			fmt.Printf("Error saving bundle: %v", err)
			return
		}

		fmt.Printf("✅ Switched to bundle %q\n", bundleName)
	},
}

func cleanupCurrentBundle(env *environment.Environment, bDefs *environment.Bundle) error {
	ma := environment.NewManifestArchive(env.ManifestCoreFile)
	bundleName := bDefs.GetActiveBundle()
	toBeRemovedPkgs := bDefs.GetInstalledPackages(bundleName)
	for pkg, _ := range toBeRemovedPkgs {
		pm, err := ma.ExactMatch(pkg)
		if err != nil {
			return fmt.Errorf("Error finding package %q: %w", pkg, err)
		}
		env.DeleteSymlink(pm.Exec)
	}

	return nil
}

func activateBundle(bundleName string, env *environment.Environment) (*environment.Bundle, error) {
	bDefs, err := environment.LoadBundleDefinitions(env)
	if err != nil {
		return nil, fmt.Errorf("Error loading bundle definitions: %w", err)
	}

	ma := environment.NewManifestArchive(env.ManifestCoreFile)
	bDefs.SetActiveBundle(bundleName)
	packages := bDefs.GetInstalledPackages(bundleName)
	for pkg, version := range packages {
		pm, err := ma.ExactMatch(pkg)
		if err != nil {
			return nil, fmt.Errorf("Error finding package %q: %w", pkg, err)
		}

		if err := env.CreateSymlink(pkg, pm.Exec, version); err != nil {
			return nil, fmt.Errorf("Error creating symlink for %q: %w", pkg, err)
		}
	}

	return bDefs, nil
}
