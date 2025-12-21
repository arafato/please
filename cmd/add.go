package cmd

import (
	"fmt"

	"github.com/arafat/please/environment"
	"github.com/spf13/cobra"
)

var (
	addSFlag bool
	descFlag string
)

func init() {
	AddCmd.Flags().BoolVarP(&addSFlag, "s", "s", false, "Switch to new bundle")
	AddCmd.Flags().StringVarP(&descFlag, "desc", "d", "", "Description of the bundle")
}

var AddCmd = &cobra.Command{
	Use:   "add <bundlename>",
	Short: "Add a new bundle",
	Long:  "Add a new bundle",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Usage: please add <bundle>")
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

		if err := bundle.AddBundle(bundleName, descFlag); err != nil {
			fmt.Println("Error adding bundle:", err)
			return
		}

		if addSFlag {
			bundle.SetActiveBundle(bundleName)
		}

		bundle.SaveBundle(env)

		fmt.Printf("✅ Successfully created bundle [%s]\n", bundleName)
		if addSFlag {
			fmt.Printf("✅ Active bundle set to [%s]\n", bundleName)
		}

	},
}
