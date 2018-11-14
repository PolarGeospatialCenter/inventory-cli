package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	inventoryCfgProfile string
)

var rootCmd = &cobra.Command{
	Use:   "inventory-cli",
	Short: "CLI for interacting with inventory api",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&inventoryCfgProfile, "profile", "", "inventory cli config profile name")
}
