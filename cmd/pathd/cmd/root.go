package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/buildwithgrove/path/cmd/pathd/cmd/config"
	"github.com/buildwithgrove/path/cmd/pathd/cmd/develop"
	pathdConfig "github.com/buildwithgrove/path/cmd/pathd/config"
)

var rootCmd = &cobra.Command{
	Use:   "cli",
	Short: "PATH CLI",
	Long: `The PATH CLI is a command-line interface for PATH (PATH API & Toolkit Harness).
It provides a set of commands to help you PATH local development and deployment.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(config.ConfigCmd)
	rootCmd.AddCommand(develop.DevelopCmd)

	if !pathdConfig.ConfigExists() {
		reader := bufio.NewReader(os.Stdin)

		err := config.RunFirstTimeSetup(reader)
		if err != nil {
			fmt.Println("Error during first-time setup:", err)
			os.Exit(1)
		}

		if err := config.PromptForDevelopmentMode(reader); err != nil {
			fmt.Println("Error during development mode prompt:", err)
			os.Exit(1)
		}
		return
	}
}
