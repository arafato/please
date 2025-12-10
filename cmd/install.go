package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/arafat/please/artifacts"
	"github.com/arafat/please/container"
	"github.com/arafat/please/environment"
	"github.com/arafat/please/utils"
	"github.com/spf13/cobra"
)

var InstallCmd = &cobra.Command{
	Use:   "install [namespace:package:version]",
	Short: "installs a containerized app, default namespace is 'core'.",
	Long:  `installs a containerized app, default namespace is 'core'.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing package name")
		}
		return nil
	},
	// TODO: This entire installation logic needs to be refactored into package appmanagement (installer, deinstaller)
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]
		e := environment.New()
		e.Initialize()

		namespace, pkg, version := parseIdentifier(packageName)

		if namespace == "" {
			namespace = "core"
		}

		// TODO: when multiple namespaces are supported
		// we need to do the according changes here since we
		// need to search in a different manifest archive
		ma := environment.NewManifestArchive(e.ManifestCoreFile)
		pm, err := ma.ExactMatch(pkg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if version == "" {
			version, err = utils.SelectFromOptions(pm.Versions, "Select a version")
			// version = pm.DefaultVersion
		}
		image := pm.Image

		env, err := environment.LoadBundleDefinitions(e)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading environment: %v\n", err)
			return
		}
		activeBundle := env.GetActiveBundle()
		if env.IsPackageInstalled(activeBundle, pkg, version) {
			fmt.Printf("Package %s:%s is already installed in active environment [%s] \n", pkg, version, activeBundle)
			return
		}

		client, err := container.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		platform := selectContainerPlatform(e.Platform, pm.Platforms)
		err = client.Install(context.TODO(), image, version, platform)
		if err != nil {
			if err.Error() == "exit status 2" {
				// NOOP - all good and expected error
			} else {
				fmt.Println(err)
				return
			}
		}

		env.AddPackage(activeBundle, pkg, version)
		env.SaveBundle(e)

		if pm.Script == "standard" {
			stdScript := &artifacts.StandardScript{
				ContainerArgs:   pm.ContainerArgs,
				ApplicationArgs: pm.ApplicationArgs,
				Image:           image,
				Version:         version,
				Application:     pkg,
				Platform:        platform,
			}

			e.DeployScript(stdScript, pkg, version)
			e.CreateSymlink(pkg, version)
		} else {
			fmt.Printf("Script type [%s] is not supported.", pm.Script)
			return
		}

		fmt.Printf("✅ Successfully installed %s:%s in bundle [%s]\n", pkg, version, activeBundle)
	},
}

func selectContainerPlatform(local string, available []string) string {
	fallback := ""
	for _, p := range available {
		if p == local {
			return local
		}
		if fallback == "" && p != local {
			fallback = p
		}
	}

	return fallback
}

func parseIdentifier(s string) (namespace, pkg, version string) {
	parts := strings.Split(s, ":")

	switch len(parts) {
	case 1:
		pkg = parts[0]
	case 2:
		pkg = parts[0]
		version = parts[1]
	case 3:
		namespace = parts[0]
		pkg = parts[1]
		version = parts[2]
	}

	return
}
