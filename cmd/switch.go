package cmd

import (
	"fmt"

	"github.com/arafat/please/environment"
	"github.com/spf13/cobra"
)

var SwitchCmd = &cobra.Command{
	Use:   "switch <bundle>",
	Short: "Switch to a different bundle and make it active",
	Long:  "Switch to a different bundle and make it active",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Usage: please switch <bundle>")
			return
		}

		bundleName := args[0]

		env := environment.New()
		if !env.IsInitialized() {
			fmt.Println("Please initialize the tool first with the init command.")
			return
		}

		if err := cleanupCurrentBundle(env); err != nil {
			fmt.Printf("Error cleaning up current bundle: %v\n", err)
		}

		bundle, err := activateBundle(bundleName, env)
		if err != nil {
			fmt.Printf("Error switching to bundle %q: %v\n", bundleName, err)
			return
		}

		if err := bundle.SaveBundle(env); err != nil {
			fmt.Printf("Error saving bundle: %w", err)
			return
		}

		fmt.Printf("✅ Switched to bundle %q\n", bundleName)
	},
}

func cleanupCurrentBundle(env *environment.Environment) error {
	bundle, err := environment.LoadBundleDefinitions(env)
	if err != nil {
		return fmt.Errorf("Error loading bundle definitions: %w", err)
	}

	ma := environment.NewManifestArchive(env.ManifestCoreFile)

	toBeRemovedPkgs := bundle.GetInstalledPackages()
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
	bundle, err := environment.LoadBundleDefinitions(env)
	if err != nil {
		return nil, fmt.Errorf("Error loading bundle definitions: %w", err)
	}

	ma := environment.NewManifestArchive(env.ManifestCoreFile)
	bundle.SetActiveBundle(bundleName)
	packages := bundle.GetInstalledPackages()
	for pkg, version := range packages {
		pm, err := ma.ExactMatch(pkg)
		if err != nil {
			return nil, fmt.Errorf("Error finding package %q: %w", pkg, err)
		}

		if err := env.CreateSymlink(pkg, pm.Exec, version); err != nil {
			return nil, fmt.Errorf("Error creating symlink for %q: %w", pkg, err)
		}
	}

	return bundle, nil
}
