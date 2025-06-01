package localnet

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"github.com/buildwithgrove/gdi/log"
	"github.com/spf13/cobra"

	"github.com/buildwithgrove/path/cmd/pathd/cmd/config"
	pathdConfig "github.com/buildwithgrove/path/cmd/pathd/config"
)

func init() {
	LocalnetCmd.AddCommand(developUpCmd)
	LocalnetCmd.AddCommand(developDownCmd)
}

// LocalnetCmd is the parent command for localnet tasks.
// It provides subcommands to bring the localnet environment up or down.
var LocalnetCmd = &cobra.Command{
	Use:   "localnet",
	Short: "Manage localnet tasks for PATH",
	Long: `The localnet command groups subcommands for managing
the PATH localnet environment.

The localnet environment is a local development environment for PATH.
It runs inside a Docker container to avoid needing to install dependencies
on the host machine.

Subcommands:
  up   : Loads the PATH configuration, installs dependencies if needed,
         and executes "make path_up" to bring the environment up.
  down : Loads the PATH configuration, installs dependencies if needed,
         and executes "make path_down" to bring the environment down.
  install-deps : Checks for required dependencies and installs them if missing.`,
}

// developUpCmd runs "make path_up" to bring up the localnet environment.
var developUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Bring up the localnet environment",
	Long:  "Loads the PATH configuration, and runs 'make path_up' in the local PATH repository.",
	Run: func(cmd *cobra.Command, args []string) {
		// Setup configuration
		setupConfig()

		if err := runMakeTask("path_up"); err != nil {
			fmt.Printf(log.Red+"❌ Failed to run 'make path_up': %v"+log.ResetColor, err)
			os.Exit(1)
		}
	},
}

// developDownCmd runs "make path_down" to bring down the localnet environment.
var developDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Bring down the localnet environment",
	Long:  "Loads the PATH configuration, installs required dependencies, and runs 'make path_down' in the local PATH repository.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runMakeTask("path_down"); err != nil {
			fmt.Println(err)
		}
	},
}

// runMakeTask is a private helper that loads the configuration,
// installs necessary dependencies, and executes a make task (e.g., "path_up" or "path_down")
// in the configured PATH repository directory.
func runMakeTask(task string) error {
	cfg, err := pathdConfig.LoadPATHDConfig()
	if err != nil {
		return fmt.Errorf("❌ Failed to load config: %v", err)
	}

	fmt.Println("Target directory:", cfg.GetPATHRepoFilepath())

	c := exec.Command("make", task)
	c.Dir = cfg.GetPATHRepoFilepath()
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("❌ Failed to run 'make %s': %v", task, err)
	}
	return nil
}

func setupConfig() {
	reader := bufio.NewReader(os.Stdin)

	// Case 1: Neither config files exist
	// 		- PATHD (.pathd) config
	// 		- PATH config (.config.yaml)
	if !pathdConfig.PATHDConfigExists() {
		initialConfig, err := config.RunFirstTimeSetup(reader)
		if err != nil {
			fmt.Println("Error during first-time setup:", err)
			os.Exit(1)
		}

		// Case 1.2: PATH local config doesn't exist so we need to create it
		if !initialConfig.PATHLocalConfigExists() {
			err := config.RunPATHConfigSetup(reader)
			if err != nil {
				fmt.Println("Error during PATH config setup:", err)
				os.Exit(1)
			}
		}
		return
	}

	// Case 2:
	// 		- PATHD (.pathd) config exists
	// 		- PATH config (.config.yaml) doesn't exist

	// Load PATHD config (.pathd)
	pathdConfigStruct, err := pathdConfig.LoadPATHDConfig()
	if err != nil {
		fmt.Println("Error loading PATHD config:", err)
		os.Exit(1)
	}

	// PATH config (.config.yaml) doesn't exist so we need to create it
	if !pathdConfigStruct.PATHLocalConfigExists() {
		fmt.Println("PATH local config file not found. Creating it now...")
		err := config.RunPATHConfigSetup(reader)
		if err != nil {
			fmt.Println("Error during PATH config setup:", err)
			os.Exit(1)
		}
	}
}
