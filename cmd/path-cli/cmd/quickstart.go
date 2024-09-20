/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/config/shannon"
	"github.com/buildwithgrove/path/relayer"
	shannonRelayer "github.com/buildwithgrove/path/relayer/shannon"
)

// TODO: Update this when a new release is made
const pathImageName = "ghcr.io/buildwithgrove/path:local-dev"

var defaultServices = map[relayer.ServiceID]config.ServiceConfig{
	"0021": {Alias: "eth-mainnet"},
}

/* -------------------- Command Initialization -------------------- */

// quickstartCmd represents the quickstart command
var quickstartCmd = &cobra.Command{
	Use:   "quickstart",
	Short: "Start the PATH service with a quickstart guide",
	Long: `This command guides you through the steps to start the PATH service.
It collects necessary inputs, generates configuration, and starts the service using Docker.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		clearScreen()
		displayWelcomeMessage()

		if !confirmProceed(reader) {
			displaySetupAborted()
			return
		}

		checkPrerequisites()

		clearScreen()
		displaySetupRequirements()

		if !confirmProceedWithSetup(reader) {
			displaySetupAborted()
			return
		}

		configInputs := collectUserInputs(reader)
		configYAMLData := generateConfigYAML(configInputs)
		startDockerService(configYAMLData)
		healthCheckWithProgressBar()
	},
}

func init() {
	rootCmd.AddCommand(quickstartCmd)
}

/* -------------------- Helper Functions and Types -------------------- */

/* -------------------- Screen Utilities -------------------- */

// clearScreen clears the terminal screen
func clearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	if _, err := cmd.Output(); err != nil {
		cmd = exec.Command("clear")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
}

/* -------------------- Display Messages -------------------- */

// displayWelcomeMessage shows the initial welcome message
func displayWelcomeMessage() {
	color.Green("🌿 Welcome to PATH. This will guide you through the steps to start the service.")
	color.Cyan("🐳 In order to proceed, Docker must be installed and running on your machine.")
}

// displaySetupAborted shows a message when the setup is aborted
func displaySetupAborted() {
	color.Red("❌ Setup aborted.")
}

// displaySetupRequirements shows the requirements needed before proceeding
func displaySetupRequirements() {
	color.Cyan("🔧 In order to proceed with setup you will need a Shannon Full Node and the following values for actors staked on the Shannon protocol:")
	fmt.Println("- A Gateway address")
	fmt.Println("- A Gateway private key")
	fmt.Println("- An address of an Application delegated to the Gateway")
	fmt.Println()
	fmt.Println("📄 For instructions on how to set all of this up yourself, please see:")
	color.Blue("https://dev.poktroll.com/operate/quickstart/docker_compose_walkthrough")
	fmt.Println()
}

// displayServiceRunningMessage shows a message when the service is successfully running
func displayServiceRunningMessage() {
	color.Green("🌿 PATH Service is now running!")
	color.Cyan("You may now send service requests for service '0021' (eth-mainnet) using http://eth-mainnet.localhost:3000/v1")
	fmt.Println()
	color.Yellow("💡 Example service request using cURL:")
	fmt.Println(`curl http://eth-mainnet.localhost:3000/v1 -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'`)
	fmt.Println()
	color.Cyan("🌱 To enable additional services, edit the 'services' section of the .config.yaml file and restart the PATH service.")
	fmt.Println()
	color.Green("💚 Happy relaying!")
}

/* -------------------- User Confirmation Prompts -------------------- */

// confirmProceed asks the user if they want to proceed
func confirmProceed(reader *bufio.Reader) bool {
	for {
		fmt.Print("❔ Would you like to proceed? (y/n): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "y" || input == "yes" {
			return true
		} else if input == "n" || input == "no" {
			return false
		} else {
			fmt.Println("Please enter 'y' or 'n'.")
		}
	}
}

// confirmProceedWithSetup asks the user if they have all requirements and want to proceed
func confirmProceedWithSetup(reader *bufio.Reader) bool {
	for {
		fmt.Print("❓ Do you have all of the above and would like to proceed? (y/n): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "y" || input == "yes" {
			return true
		} else if input == "n" || input == "no" {
			return false
		} else {
			fmt.Println("Please enter 'y' or 'n'.")
		}
	}
}

/* -------------------- Prerequisite Checks -------------------- */

// checkPrerequisites ensures that Docker is installed and running
func checkPrerequisites() {
	if !commandExists("docker") {
		color.Red("❌ Docker is not installed. Please install Docker and try again.")
		os.Exit(1)
	}

	if !dockerDaemonRunning() {
		color.Red("❌ Docker daemon is not running. Please start Docker and try again.")
		os.Exit(1)
	}
}

// commandExists checks if a command exists in the system
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// dockerDaemonRunning checks if the Docker daemon is running
func dockerDaemonRunning() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	return err == nil
}

/* -------------------- User Input Collection -------------------- */

// ConfigInputs holds the user inputs
type ConfigInputs struct {
	rpcURL              string
	hostPort            string
	useTLS              bool
	gatewayAddress      string
	gatewayPrivateKey   string
	delegatedAppAddress string
}

// collectUserInputs collects inputs from the user with validation
func collectUserInputs(reader *bufio.Reader) ConfigInputs {
	var config ConfigInputs

	// Collecting inputs using standard prompts with validation
	config.rpcURL = promptInput(reader, "🔗 Please enter your Full Node URL (e.g., http://path-service:26657):", validateURL)
	config.hostPort = promptInput(reader, "🔗 Please enter your Full Node gRPC host & port (e.g., path-service:9090):", validateHostPort)
	useTLSInput := promptSelect(reader, "❓ Does your Full Node gRPC connection use TLS? (Yes/No):", []string{"Yes", "No"})
	config.useTLS = strings.EqualFold(useTLSInput, "Yes")
	config.gatewayAddress = promptInput(reader, "🔗 Please enter your Gateway address (43 characters starting with pokt1...):", validateAddress)
	config.gatewayPrivateKey = promptPassword(reader, "🔗 Please enter your Gateway private key (64-character hexadecimal string):", validateGatewayPrivateKey)
	config.delegatedAppAddress = promptInput(reader, "🔗 Please enter your delegated Application address (43 characters starting with pokt1...):", validateAddress)

	return config
}

/* -------------------- Input Prompts -------------------- */

// promptInput prompts the user for input and validates it
func promptInput(reader *bufio.Reader, message string, validateFunc func(string) error) string {
	for {
		fmt.Println(message)
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if err := validateFunc(input); err != nil {
			fmt.Println(err.Error())
		} else {
			return input
		}
	}
}

// promptPassword securely prompts the user for a password input
func promptPassword(reader *bufio.Reader, message string, validateFunc func(string) error) string {
	for {
		fmt.Println(message)
		fmt.Println("NOTE: Input will not be displayed on screen.")
		fmt.Print("> ")
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			fmt.Println("Error reading password:", err)
			continue
		}
		input := strings.TrimSpace(string(bytePassword))
		if err := validateFunc(input); err != nil {
			fmt.Println(err.Error())
		} else {
			return input
		}
	}
}

// promptSelect prompts the user to select from provided options
func promptSelect(reader *bufio.Reader, message string, options []string) string {
	for {
		fmt.Println(message)
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		for _, option := range options {
			if strings.EqualFold(input, option) {
				return option
			}
		}
		fmt.Printf("Please enter one of the following options: %s\n", strings.Join(options, ", "))
	}
}

/* -------------------- Input Validation -------------------- */

// validateURL checks if the provided URL is valid
func validateURL(url string) error {
	re := regexp.MustCompile(`^http[s]?://[a-zA-Z0-9.-]+(:[0-9]+)?`)
	if !re.MatchString(url) {
		return fmt.Errorf("Invalid URL. Must be a valid URL (e.g., https://example.com)")
	}
	return nil
}

// validateHostPort checks if the host and port are valid
func validateHostPort(hostport string) error {
	re := regexp.MustCompile(`^[a-zA-Z0-9.-]+:[0-9]+$`)
	if !re.MatchString(hostport) {
		return fmt.Errorf("Invalid host port. Must be in the format 'hostname:port' (e.g., localhost:9090)")
	}
	return nil
}

// validateAddress checks if the address is valid
func validateAddress(address string) error {
	re := regexp.MustCompile(`^pokt1[0-9a-zA-Z]{38}$`)
	if !re.MatchString(address) {
		return fmt.Errorf("Invalid address. Must be 43 characters long and start with 'pokt1'")
	}
	return nil
}

// validateGatewayPrivateKey checks if the private key is valid
func validateGatewayPrivateKey(key string) error {
	re := regexp.MustCompile(`^[0-9a-fA-F]{64}$`)
	if !re.MatchString(key) {
		return fmt.Errorf("Invalid gateway private key. Must be a 64-character hexadecimal string")
	}
	return nil
}

/* -------------------- Configuration Generation -------------------- */

// generateConfigYAML creates the configuration YAML data from user inputs
func generateConfigYAML(configInput ConfigInputs) []byte {

	cfg := config.GatewayConfig{
		ShannonConfig: &shannon.ShannonGatewayConfig{
			FullNodeConfig: shannonRelayer.FullNodeConfig{
				RpcURL: configInput.rpcURL,
				GRPCConfig: shannonRelayer.GRPCConfig{
					HostPort: configInput.hostPort,
				},
				GatewayAddress:    configInput.gatewayAddress,
				GatewayPrivateKey: configInput.gatewayPrivateKey,
				DelegatedApps:     []string{configInput.delegatedAppAddress},
			},
		},
		Services: defaultServices,
	}

	// Set 'insecure' based on useTLS
	if configInput.useTLS {
		cfg.ShannonConfig.FullNodeConfig.GRPCConfig.Insecure = false
	} else {
		cfg.ShannonConfig.FullNodeConfig.GRPCConfig.Insecure = true
	}

	// Marshal the updated config to YAML
	outData, err := yaml.Marshal(&cfg)
	if err != nil {
		color.Red("❌ Failed to marshal config data: %v", err)
		os.Exit(1)
	}

	return outData
}

/* -------------------- Docker Service Management -------------------- */

// startDockerService pulls the Docker image and starts the container with the configuration
func startDockerService(configYAMLData []byte) {
	color.Cyan("🌿 Starting PATH service...")

	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		color.Red("❌ Failed to create Docker client: %v", err)
		os.Exit(1)
	}
	cli.NegotiateAPIVersion(context.Background())

	ctx := context.Background()

	// Check if image exists locally
	_, _, err = cli.ImageInspectWithRaw(ctx, pathImageName)
	if err != nil {
		if client.IsErrNotFound(err) {
			color.Red("❌ Docker image not found locally. Please pull the image manually using:")
			fmt.Printf("   docker pull %s\n", pathImageName)
			os.Exit(1)
		} else {
			color.Red("❌ Failed to inspect Docker image: %v", err)
			os.Exit(1)
		}
	} else {
		color.Green("✅ Docker image found; starting PATH service...")
	}

	// Expose and map port 3000
	port, err := nat.NewPort("tcp", "3000")
	if err != nil {
		color.Red("❌ Failed to parse port: %v", err)
		os.Exit(1)
	}

	// Create a container with port bindings
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: pathImageName,
		ExposedPorts: nat.PortSet{
			port: struct{}{},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			port: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "3000",
				},
			},
		},
	}, nil, nil, "path_gateway")
	if err != nil {
		color.Red("❌ Failed to create container: %v", err)
		os.Exit(1)
	}

	// Convert the in-memory YAML data to a tar stream for Docker
	yamlFileTar, err := createTarFromBytes(".config.yaml", configYAMLData)
	if err != nil {
		color.Red("❌ Failed to create tar from bytes: %v", err)
		os.Exit(1)
	}

	// Copy the YAML file into the container at /app/.config.yaml
	err = cli.CopyToContainer(ctx, resp.ID, "/app", yamlFileTar, container.CopyToContainerOptions{})
	if err != nil {
		color.Red("❌ Failed to copy config file to container: %v", err)
		os.Exit(1)
	}

	// Start the container
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		color.Red("❌ Failed to start container: %v", err)
		os.Exit(1)
	}

	// Check if the container is running
	go func() {
		statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNextExit)
		select {
		case err := <-errCh:
			if err != nil {
				color.Red("❌ Container exited with error: %v", err)
				os.Exit(1)
			}
		case status := <-statusCh:
			if status.StatusCode != 0 {
				out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
				if err != nil {
					color.Red("❌ Failed to retrieve container logs: %v", err)
				} else {
					_, _ = io.Copy(os.Stdout, out)
				}
				color.Red("❌ Container exited with status code: %d", status.StatusCode)
				os.Exit(1)
			}
		}
	}()

	color.Green("🚀 PATH service started successfully.")
}

// createTarFromBytes creates a tar archive containing the in-memory YAML data
func createTarFromBytes(filename string, data []byte) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	hdr := &tar.Header{
		Name: filename,
		Mode: 0600,
		Size: int64(len(data)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}

	if _, err := tw.Write(data); err != nil {
		return nil, err
	}

	tw.Close()
	return buf, nil
}

/* -------------------- Health Check -------------------- */

// healthCheckWithProgressBar checks if the service is up and running
func healthCheckWithProgressBar() {
	// Wait for the service to start with a progress bar
	timeout := 20
	bar := progressbar.NewOptions(timeout,
		progressbar.OptionSetDescription("⏳ Waiting for PATH service to become healthy..."),
		progressbar.OptionSetWidth(30),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(false),
	)

	for i := 0; i < timeout; i++ {
		resp, err := http.Get("http://localhost:3000/healthz")
		if err == nil && resp.StatusCode == 200 {
			clearScreen()
			displayServiceRunningMessage()
			os.Exit(0)
		}
		_ = bar.Add(1)
		time.Sleep(1 * time.Second)
	}
	_ = bar.Finish()
	color.Red("❌ Service health check failed after %d seconds.", timeout)
	os.Exit(1)
}
