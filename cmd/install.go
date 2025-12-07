package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/arafat/please/container"
	"github.com/arafat/please/storage"
	"github.com/spf13/cobra"
)

var version string

func init() {
	InstallCmd.Flags().StringVar(&version, "version", "", "Version of the package to install")
}

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
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]
		s := storage.New()
		s.Initialize()

		namespace, pkg, version := parseIdentifier(packageName)

		if namespace == "" {
			namespace = "core"
		}

		// TODO: when multiple namespaces are supported
		// we need to do the according changes here since we
		// need to search in a different manifest archive
		ma := storage.NewManifestArchive(s.ManifestCoreFile)
		pm, err := ma.ExactMatch(pkg)
		if err != nil {
			fmt.Println(err)
			return
		}
		client, err := container.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		if version == "" {
			version = pm.DefaultVersion
		}

		image := pm.Image

		err = client.Install(context.TODO(), image, version)
		if err != nil {
			if err.Error() == "exit status 2" {
				return
			}
			fmt.Println(err)
			return
		}
	},
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
