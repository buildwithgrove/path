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
	"golang.org/x/term"
)

const pathRepo = "https://github.com/buildwithgrove/path"

// RunFirstTimeSetup performs an interactive configuration when the config file does not exist.
func RunFirstTimeSetup(reader *bufio.Reader) (*config.Config, error) {
	cfgEditor.ClearTerminal()
	fmt.Println(log.Green + "🌿 Welcome to PATH! It looks like this is the first time you're using it." + log.ResetColor)

	pathRepoPath, err := promptForPathRepoPath(reader)
	if err != nil {
		return nil, err
	}
	fmt.Println("✅ Local PATH repo path saved as: " + pathRepoPath + "\n")

	// Save the config file
	config, err := saveConfig(pathRepoPath)
	if err != nil {
		fmt.Printf(log.Red+"❌ Failed to save config file: %v"+log.ResetColor, err)
		return nil, fmt.Errorf("failed to save config file: %v", err)
	}

	return config, nil
}

func RunPATHConfigSetup(reader *bufio.Reader) error {
	schema, err := config.LoadSchema()
	if err != nil {
		return fmt.Errorf("failed to load schema: %v", err)
	}

	pathdConfig, err := config.LoadPATHDConfig()
	if err != nil {
		return fmt.Errorf("failed to load PATHD config: %v", err)
	}

	// Prompt for configuring Morse and Shannon
	if err := promptForMorseAndShannon(reader, pathdConfig, schema); err != nil {
		return err
	}

	// (inside RunFirstTimeSetup, after printing the completion message)
	fmt.Println(log.Green + "\n🌿 PATH configuration completed and saved.\n" + log.Blue + "\nℹ️  You may edit the PATH local config file at any time by running " + log.ResetColor + "'pathd config'" + log.Blue + ".\n" + log.ResetColor)

	return nil
}

// promptToStartLocalnet prompts the user if they would like to run PATH Localnet and executes the appropriate command.
func promptToStartLocalnet(reader *bufio.Reader) error {
	devChoice, err := prompt(reader, log.Blue+"Would you like to run PATH Localnet now? (y/n):"+log.ResetColor)
	if err != nil {
		return err
	}
	devChoice = strings.TrimSpace(strings.ToLower(devChoice))
	cfgEditor.ClearTerminal()

	switch devChoice {
	case "y":
		fmt.Println(log.Green + "🚀 Starting PATH Localnet ..." + log.ResetColor)
		cmd := exec.Command("pathd", "localnet", "up")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(log.Red + "❌ Failed to run PATH Localnet: " + err.Error() + log.ResetColor)
			return err
		}
	case "n":
		fmt.Println(log.Blue + "👋 Goodbye! You can run PATH Localnet at any time by running 'pathd localnet up'." + log.ResetColor)
	default:
		fmt.Println(log.Blue + "👋 Goodbye! You can run PATH Localnet at any time by running 'pathd localnet up'." + log.ResetColor)
	}
	return nil
}

// promptForPathRepoPath prompts the user to either use an existing local PATH repo
// or clone the PATH repo to a location on their computer.
func promptForPathRepoPath(reader *bufio.Reader) (string, error) {
	for {
		fmt.Println(log.Blue + "❓ Which of the following applies to you?" + log.ResetColor)
		fmt.Println("   1. I already have a locally cloned PATH repo checked out to the latest `main` branch.")
		fmt.Println("   2. I would like to clone the PATH repo to a location on my computer.")
		input, err := prompt(reader, log.Blue+"Enter your choice (1/2): "+log.ResetColor)
		if err != nil {
			return "", err
		}
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
	for {
		input, err := prompt(reader, log.Blue+"📝 Enter the absolute filepath to your local PATH repo"+log.ResetColor+" (e.g. /Users/greg/grove/path):")
		if err != nil {
			return "", err
		}
		if input == "" {
			fmt.Println(log.Red + "Invalid input. Please enter a non-empty path." + log.ResetColor)
			continue
		}

		// Validate that the provided path exists and is a directory
		fi, err := os.Stat(input)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf(log.Red+"❌ The provided path "+log.ResetColor+"%s"+log.Red+" does not exist.\n"+log.ResetColor, input)
				choice, err := prompt(reader, "❔ Would you like to (r)etry entering the path, (c)lone the repo instead, or (e)xit?")
				if err != nil {
					return "", err
				}
				choice = strings.ToLower(strings.TrimSpace(choice))
				if choice == "r" {
					continue
				} else if choice == "c" {
					return promptForClonePathRepoPath(reader)
				} else if choice == "e" {
					return "", fmt.Errorf("exiting setup")
				} else {
					fmt.Println(log.Red + "❌ Invalid selection. Please type 'r', 'c', or 'e'." + log.ResetColor)
					continue
				}
			}
			fmt.Printf(log.Red+"Error checking path: %v"+log.ResetColor, err)
			continue
		}
		if !fi.IsDir() {
			fmt.Println(log.Red + "The provided path is not a directory." + log.ResetColor)
			continue
		}
		cfgEditor.ClearTerminal()
		return input, nil
	}
}

// promptForClonePathRepoPath prompts the user to provide the filepath to where they want to clone the PATH repo.
func promptForClonePathRepoPath(reader *bufio.Reader) (string, error) {
	clonePath, err := prompt(reader, log.Blue+"📝 Enter the absolute filepath where you want to clone the PATH repo:"+log.ResetColor)
	if err != nil {
		return "", err
	}
	clonePath = strings.TrimSpace(clonePath)
	if !strings.HasSuffix(clonePath, "path") {
		clonePath += "/path"
	}
	if err := validateClonePath(clonePath); err != nil {
		return "", err
	}
	if err := clonePathRepo(clonePath); err != nil {
		return "", err
	}
	cfgEditor.ClearTerminal()
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
			fmt.Printf("   %d."+log.Purple+" %s"+log.ResetColor+"\n", i+1, proto)
		}

		input, err := prompt(reader, log.Blue+"Enter your choice:"+log.ResetColor)
		if err != nil {
			return err
		}
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
	switch protocol {
	case "shannon":
		// All shannon-specific code has been moved to ConfigureShannon.
		return ConfigureShannon(conf, schema)
	case "morse":
		// TODO_IMPROVE(@commoddity): uncomment this when ConfigureMorse is implemented.
		// targetPath := conf.GetPATHConfigFilepath()
		// if _, err := os.Stat(targetPath); err == nil {
		// 	fmt.Printf(log.Yellow+"⚠️ File '%s' already exists. Skipping creation to avoid overwriting.\n"+log.ResetColor, targetPath)
		// 	return fmt.Errorf("file '%s' already exists. Skipping creation", targetPath)
		// }
		// sourcePath := conf.GetExamplePATHConfigFilepath("morse")
		// if sourcePath == "" {
		// 	return fmt.Errorf("no example config found for protocol: morse")
		// }
		// if err := copyAndStripComments(sourcePath, targetPath); err != nil {
		// 	fmt.Printf(log.Red+"❌ Failed to create config file from example: %v"+log.ResetColor, err)
		// 	return fmt.Errorf("failed to create config file: %v", err)
		// }
		// fmt.Printf(log.Green+"✅ Created config file for protocol 'morse' at '%s'\n"+log.ResetColor, targetPath)
		// // Uncomment the following lines when ConfigureMorse is implemented.
		// // if err := ConfigureMorse(conf, schema); err != nil {
		// //	   return err
		// // }
		return nil
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
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

// prompt ensures that the promtp shows `>` on a new line.
func prompt(reader *bufio.Reader, message string) (string, error) {
	fmt.Println(message)
	fmt.Print("> ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// promptHidden reads input from the terminal without echoing it.
// Ensure that "golang.org/x/term" is imported in this file.
func promptHidden(reader *bufio.Reader, message string) (string, error) {
	fmt.Println(message)
	fmt.Print("> ")
	byteInput, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println("") // Print a newline after hidden input.
	return strings.TrimSpace(string(byteInput)), nil
}
