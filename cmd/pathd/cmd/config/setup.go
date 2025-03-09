package config

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	cfgEditor "github.com/buildwithgrove/gdi/config/editor"
	"github.com/buildwithgrove/gdi/log"
	"github.com/go-git/go-git/v5"
	"gopkg.in/yaml.v3"

	"github.com/buildwithgrove/path/cmd/pathd/config"
)

const pathRepo = "https://github.com/buildwithgrove/path"

// RunFirstTimeSetup performs an interactive configuration when the config file does not exist.
func RunFirstTimeSetup() error {
	reader := bufio.NewReader(os.Stdin)

	schema, err := config.LoadSchema()
	if err != nil {
		return fmt.Errorf("failed to load schema: %v", err)
	}

	cfgEditor.ClearTerminal()
	fmt.Println(log.Green + "🌿 Welcome to PATH! It looks like this is the first time you're using it." + log.ResetColor)

	pathRepoPath, err := promptForPathRepoPath(reader)
	if err != nil {
		return err
	}
	fmt.Println(log.Blue + "🌿 Local PATH repo path saved as: " + pathRepoPath + log.ResetColor)

	// Save the config file
	savedConfig, err := saveConfig(pathRepoPath)
	if err != nil {
		fmt.Printf(log.Red+"❌ Failed to save config file: %v"+log.ResetColor, err)
		return fmt.Errorf("failed to save config file: %v", err)
	}

	// Prompt for configuring Morse and Shannon
	if err := promptForMorseAndShannon(reader, savedConfig, schema); err != nil {
		return err
	}

	// (inside RunFirstTimeSetup, after printing the completion message)
	fmt.Println(log.Green + "🌿 PATH configuration completed and saved.\n" + log.Blue + "ℹ️ You may edit the PATH local config file at any time by running 'pathd config'." + log.ResetColor)

	if err := promptForDevelopmentMode(reader); err != nil {
		return err
	}

	return nil
}

// promptForDevelopmentMode prompts the user if they would like to run PATH in development mode and executes the appropriate command.
func promptForDevelopmentMode(reader *bufio.Reader) error {
	fmt.Print(log.Blue + "Would you like to run PATH in development mode now? (y/n): " + log.ResetColor)
	devChoice, _ := reader.ReadString('\n')
	devChoice = strings.TrimSpace(strings.ToLower(devChoice))
	cfgEditor.ClearTerminal()

	if devChoice == "y" {
		fmt.Println(log.Green + "🚀Starting PATH in development mode..." + log.ResetColor)
		cmd := exec.Command("pathd", "develop", "up")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(log.Red + "❌ Failed to run PATH in development mode: " + err.Error() + log.ResetColor)
			return err
		}
	} else {
		fmt.Println(log.Blue + "👋 Goodbye! You can run PATH in development mode at any time by running 'pathd develop up'." + log.ResetColor)
	}
	return nil
}

// promptForPathRepoPath prompts the user to either use an existing local PATH repo
// or clone the PATH repo to a location on their computer.
func promptForPathRepoPath(reader *bufio.Reader) (string, error) {
	for {
		fmt.Println(log.Blue + "❓ Which of the following applies to you?" + log.ResetColor)
		fmt.Println("1. I already have a locally cloned PATH repo checked out to the latest `main` branch.")
		fmt.Println("2. I would like to clone the PATH repo to a location on your computer.")
		fmt.Print(log.Blue + "Enter your choice (1/2): " + log.ResetColor)

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		choice, err := strconv.Atoi(input)
		if err != nil || (choice != 1 && choice != 2) {
			fmt.Println(log.Red + "Invalid selection. Please enter 1 or 2." + log.ResetColor)
			continue
		}

		cfgEditor.ClearTerminal()
		if choice == 1 {
			return promptForLocalPathRepoPath(reader)
		}
		return promptForClonePathRepoPath(reader)
	}
}

// promptForLocalPathRepoPath prompts the user to provide the filepath to their local PATH repo.
func promptForLocalPathRepoPath(reader *bufio.Reader) (string, error) {
	fmt.Print(log.Blue + "📝 Enter the absolute filepath to your local PATH repo: " + log.ResetColor)
	pathRepoPath, _ := reader.ReadString('\n')
	pathRepoPath = strings.TrimSpace(pathRepoPath)
	return pathRepoPath, nil
}

// promptForClonePathRepoPath prompts the user to provide the filepath to where they want to clone the PATH repo.
func promptForClonePathRepoPath(reader *bufio.Reader) (string, error) {
	fmt.Print(log.Blue + "📝 Enter the absolute filepath where you want to clone the PATH repo: " + log.ResetColor)
	clonePath, _ := reader.ReadString('\n')
	clonePath = strings.TrimSpace(clonePath)

	// Check if the provided path already ends with "path"
	if !strings.HasSuffix(clonePath, "path") {
		clonePath += "/path"
	}
	cfgEditor.ClearTerminal()

	if err := validateClonePath(clonePath); err != nil {
		return "", err
	}

	// Clone the PATH repo
	if err := clonePathRepo(clonePath); err != nil {
		return "", err
	}

	return clonePath, nil
}

// validateClonePath validates the clone path.
func validateClonePath(clonePath string) error {
	// Check if the directory already exists and is not empty
	if _, err := os.Stat(clonePath); !os.IsNotExist(err) {
		files, err := os.ReadDir(clonePath)
		if err != nil {
			fmt.Printf(log.Red+"❌ Failed to read directory: %v"+log.ResetColor, err)
			return fmt.Errorf("failed to read directory: %w", err)
		}
		if len(files) > 0 {
			fmt.Printf(log.Red+"❌ Directory '%s' already exists and is not empty"+log.ResetColor, clonePath)
			return fmt.Errorf("directory '%s' already exists and is not empty", clonePath)
		}
	}
	return nil
}

// clonePathRepo clones the PATH repo to the given path.
func clonePathRepo(clonePath string) error {
	_, err := git.PlainClone(clonePath, false, &git.CloneOptions{
		URL: pathRepo,
	})
	if err != nil {
		return fmt.Errorf("failed to clone PATH repo: %v", err)
	}
	return nil
}

// saveConfig saves the config to the config file.
func saveConfig(pathRepoPath string) (*config.Config, error) {
	configData := &config.Config{
		PATHRepo: pathRepoPath,
	}
	if err := config.SaveConfigToFile(configData); err != nil {
		return nil, fmt.Errorf("failed to save config file: %v", err)
	}
	return configData, nil
}

// promptForMorseAndShannon prompts the user to select a protocol (morse or shannon)
// and then processes that selection by creating the appropriate config file and calling the
// corresponding configuration function.
func promptForMorseAndShannon(reader *bufio.Reader, conf *config.Config, schema *yaml.Node) error {
	fmt.Println(log.Green + "🌿 Configuring PATH: selecting protocol." + log.ResetColor)

	protocols := []string{"morse", "shannon"}

	for {
		fmt.Println(log.Blue + "Select one of the following protocols for configuration (or type 's' to skip):" + log.ResetColor)
		for i, proto := range protocols {
			fmt.Printf("%d. %s\n", i+1, proto)
		}

		fmt.Print(log.Blue + "Enter your choice: " + log.ResetColor)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		cfgEditor.ClearTerminal()

		if input == "s" {
			fmt.Println(log.Yellow + "Skipping protocol configuration setup." + log.ResetColor)
			return nil
		}

		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(protocols) {
			fmt.Println(log.Red + "Invalid choice. Please try again." + log.ResetColor)
			continue
		}

		selectedProtocol := protocols[choice-1]
		return processProtocolSelection(reader, conf, schema, selectedProtocol)
	}
}

// processProtocolSelection creates the appropriate config file for the selected protocol
// and then calls the respective configuration function.
// It uses copyAndStripComments to create the config file based on the example file from the repo.
func processProtocolSelection(reader *bufio.Reader, conf *config.Config, schema *yaml.Node, protocol string) error {
	targetPath := conf.GetPATHConfigFilepath()
	if _, err := os.Stat(targetPath); err == nil {
		fmt.Printf(log.Yellow+"⚠️ File '%s' already exists. Skipping creation to avoid overwriting.\n"+log.ResetColor, targetPath)
		return fmt.Errorf("file '%s' already exists. Skipping creation", targetPath)
	}

	var sourcePath string
	switch protocol {
	case "shannon":
		sourcePath = conf.GetExamplePATHConfigFilepath("shannon")
	case "morse":
		sourcePath = conf.GetExamplePATHConfigFilepath("morse")
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}

	if sourcePath == "" {
		return fmt.Errorf("no example config found for protocol: %s", protocol)
	}

	if err := copyAndStripComments(sourcePath, targetPath); err != nil {
		fmt.Printf(log.Red+"❌ Failed to create config file from example: %v"+log.ResetColor, err)
		return fmt.Errorf("failed to create config file: %v", err)
	}

	fmt.Printf(log.Green+"✅ Created config file for protocol '%s' at '%s'\n"+log.ResetColor, protocol, targetPath)

	// Call the specific configuration function for the selected protocol.
	// Uncomment the following lines when ConfigureShannon and ConfigureMorse are implemented.
	switch protocol {
	case "shannon":
		if err := ConfigureShannon(conf, schema); err != nil {
			return err
		}
	case "morse":
		// if err := ConfigureMorse(); err != nil {
		// 	return err
		// }
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}

	return nil
}

// stripComments reads the file at srcPath and returns its content with every comment stripped out.
func stripComments(srcPath string) ([]byte, error) {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) == 0 || strings.HasPrefix(trimmed, "#") {
			continue
		}
		result = append(result, line)
	}
	return []byte(strings.Join(result, "\n")), nil
}

// copyAndStripComments copies the file from src to dst after stripping out every comment.
func copyAndStripComments(src, dst string) error {
	data, err := stripComments(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
