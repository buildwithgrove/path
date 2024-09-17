/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "path-cli",
	Short: color.GreenString("🌿 PATH CLI - Path API & Toolkit Harness"),
	Long: color.BlueString(`PATH CLI is a command-line tool designed to help you interact with the PATH service, 
a framework for enabling access to a decentralized supply network.

This tool provides various commands to streamline the integration and interaction with decentralized protocols. 
Use the 'quickstart' command to get started quickly, and refer to the documentation for more detailed usage examples.`),
	Run: func(cmd *cobra.Command, args []string) {
		color.Green("🌿 PATH CLI - Path API & Toolkit Harness")
		color.Blue("This CLI tool helps you interact with the PATH service, a framework for enabling access to a decentralized supply network.")
		color.Cyan("Use the 'quickstart' command to get started quickly.")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.path-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
