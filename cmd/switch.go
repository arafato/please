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

		bundle, err := environment.LoadBundleDefinitions(env)
		if err != nil {
			fmt.Println("Error loading bundle definitions:", err)
			return
		}

		ma := environment.NewManifestArchive(env.ManifestCoreFile)

		toBeRemovedPkgs := bundle.GetInstalledPackages()
		for pkg, _ := range toBeRemovedPkgs {
			pm, err := ma.ExactMatch(pkg)
			if err != nil {
				fmt.Println(err)
				return
			}
			env.DeleteSymlink(pm.Exec)
		}

		bundle.SetActiveBundle(bundleName)
		packages := bundle.GetInstalledPackages()

		for pkg, version := range packages {
			pm, err := ma.ExactMatch(pkg)
			if err != nil {
				fmt.Println(err)
				return
			}

			if err := env.CreateSymlink(pkg, pm.Exec, version); err != nil {
				fmt.Println("Error creating symlink: ", err)
				return
			}
		}
		bundle.SaveBundle(env)
		fmt.Printf("✅ Switched to bundle %q\n", bundleName)
	},
}
