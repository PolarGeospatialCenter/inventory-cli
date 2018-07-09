package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	apiBaseUrl   string
	awsRegion    string
	awsVaultRole string
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
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.inventory-cli.yml)")
	rootCmd.PersistentFlags().StringVarP(&apiBaseUrl, "api-url", "u", "", "base url for api calls")
	rootCmd.PersistentFlags().StringVarP(&awsRegion, "aws-region", "r", "us-east-2", "aws region to use for authentication")
	rootCmd.PersistentFlags().StringVar(&awsVaultRole, "aws-vault-role", "", "if set, attempt to obtain sts credentials from vault")
	viper.BindPFlag("baseurl", rootCmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag("aws.region", rootCmd.PersistentFlags().Lookup("aws-region"))
	viper.BindPFlag("aws.vault_role", rootCmd.PersistentFlags().Lookup("aws-vault-role"))
	cmdNode.AddCommand(cmdNodeList)
	cmdNode.AddCommand(cmdNodeInteractiveCreate)
	cmdNode.AddCommand(cmdNodeDetectNetworks)
	rootCmd.AddCommand(cmdNode)
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".inventory-cli")
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}
}
