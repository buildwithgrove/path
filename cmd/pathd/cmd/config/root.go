package config

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	cfgEditor "github.com/buildwithgrove/gdi/config/editor"

	"github.com/buildwithgrove/path/cmd/pathd/config"
)

var (
	show           bool
	editor         string
	configFilePath string
)

// init sets up flags for the config command.
func init() {
	var defaultConfigFilePath string
	if pathdConfig, _ := config.LoadConfig(); pathdConfig != nil {
		defaultConfigFilePath = pathdConfig.GetPATHConfigFilepath()
	}

	ConfigCmd.Flags().BoolVarP(&show, "show", "s", false, "Show the configuration.")
	ConfigCmd.Flags().StringVarP(&editor, "editor", "e", "", "Edit the configuration in the given text editor.")
	ConfigCmd.Flags().StringVarP(&configFilePath, "config", "c", defaultConfigFilePath, "The path to the configuration file.")
}

// ConfigCmd represents the interactive configuration command.
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Edit the configuration for the application.",
	Long: `Edit the configuration for the application.

This command is used to modify the YAML configuration file for PATH.
It uses an interactive command-line interface to traverse and update configuration fields,
using the schema defined in ./config/config.schema.yaml.
	  
Flags:
  --show (-s)   : Show the current config file.
  --editor (-e) : Open the config file in a specified text editor.
  --config (-c) : The path to the configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle the --show flag: print the configuration file.
		if show {
			showConfig()
			return
		}

		// Handle the --editor flag: open the file in the given text editor.
		if editor != "" {
			editConfig(editor)
			return
		}

		// Otherwise, start the interactive configuration editor.
		schema, err := config.LoadSchema()
		if err != nil {
			fmt.Printf("Failed to load schema: %v", err)
			os.Exit(1)
		}

		// Define custom field handlers.
		customHandlerFuncs := []cfgEditor.WithCustomFieldHandlerFunc{}

		yamlEditor, err := cfgEditor.NewYAMLEditor(configFilePath, schema, customHandlerFuncs...)
		if err != nil {
			fmt.Printf("Failed to create editor: %v", err)
			os.Exit(1)
		}

		// Start the interactive editor.
		yamlEditor.InteractiveEditConfig()
	},
}

// showConfig prints the current configuration file to stdout.
func showConfig() {
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		fmt.Printf("Failed to read config file: %v", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// editConfig opens the configuration file in the user's preferred text editor.
func editConfig(editor string) {
	cmd := exec.Command(editor, configFilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
}
