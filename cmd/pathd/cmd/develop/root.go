package develop

import (
	"fmt"
	"os"
	"os/exec"

	pathdConfig "github.com/buildwithgrove/path/cmd/pathd/config"
	"github.com/spf13/cobra"
)

func init() {
	DevelopCmd.AddCommand(developUpCmd)
	DevelopCmd.AddCommand(developDownCmd)
}

// DevelopCmd is the parent command for development tasks.
// It provides subcommands to bring the development environment up or down.
var DevelopCmd = &cobra.Command{
	Use:   "develop",
	Short: "Manage development tasks for PATH",
	Long: `The develop command groups subcommands for managing
the PATH development environment.

Subcommands:
  up   : Loads the PATH configuration, installs dependencies if needed,
         and executes "make path_up" to bring the environment up.
  down : Loads the PATH configuration, installs dependencies if needed,
         and executes "make path_down" to bring the environment down.`,
}

// developUpCmd runs "make path_up" to bring up the development environment.
var developUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Bring up the development environment",
	Long:  "Loads the PATH configuration, installs required dependencies, and runs 'make path_up' in the local PATH repository.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runMakeTask("path_up"); err != nil {
			fmt.Println(err)
		}
	},
}

// developDownCmd runs "make path_down" to bring down the development environment.
var developDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Bring down the development environment",
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
	cfg, err := pathdConfig.LoadConfig()
	if err != nil {
		return fmt.Errorf("❌ Failed to load config: %v", err)
	}

	fmt.Println("Target directory:", cfg.GetPATHRepoFilepath())

	if err := checkAndInstallDependencies(); err != nil {
		return fmt.Errorf("❌ Failed to install dependencies: %v", err)
	}

	c := exec.Command("make", task)
	c.Dir = cfg.GetPATHRepoFilepath()
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("❌ Failed to run 'make %s': %v", task, err)
	}
	return nil
}
