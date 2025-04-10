package develop

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/buildwithgrove/gdi/log"
)

// DependencyName represents a string enum of dependency names.
type DependencyName string

const (
	Docker    DependencyName = "🐳 Docker"
	Kind      DependencyName = "🌀 Kind"
	Kubectl   DependencyName = "🔧 kubectl"
	Helm      DependencyName = "⛵ Helm"
	Tilt      DependencyName = "🚀 Tilt"
	RelayUtil DependencyName = "🚚 Relay Util"
)

// SystemArch represents the architecture and OS of the system.
type SystemArch int

const (
	ArchUnknown SystemArch = iota
	MacX86
	MacARM
	LinuxX86
	LinuxARM
)

// getSystemArch determines the architecture and OS of the system.
func getSystemArch() SystemArch {
	arch := runtime.GOARCH
	osType := runtime.GOOS

	switch {
	case osType == "darwin" && arch == "amd64":
		return MacX86
	case osType == "darwin" && arch == "arm64":
		return MacARM
	case osType == "linux" && arch == "amd64":
		return LinuxX86
	case osType == "linux" && arch == "arm64":
		return LinuxARM
	default:
		return ArchUnknown
	}
}

// Dependency represents an external dependency required to run PATH in development mode.
type Dependency struct {
	Name        DependencyName
	Cmd         string
	Description string
	InstallCmd  string
	InstallFunc func() error
}

// -------------------- Main Func --------------------

// checkAndInstallDependencies checks for missing dependencies and installs them if the user agrees.
func checkAndInstallDependencies(reader *bufio.Reader) error {
	// Get system arch
	arch := getSystemArch()

	// Compute install commands
	cmds := computeInstallCommands(arch)

	// Get missing dependencies
	missing := getMissingDependencies(arch, cmds)

	// If no missing dependencies, return
	if len(missing) == 0 {
		fmt.Println(log.Green + "✅ All dependencies are installed." + log.ResetColor)
		return nil
	}

	// Prompt user to install missing dependencies
	shouldInstall, err := promptUserToInstall(missing, reader)
	if err != nil {
		return err
	}

	// If user doesn't want to install, return
	if !shouldInstall {
		return fmt.Errorf("installation aborted by user")
	}

	// Install missing dependencies
	return installMissingDependencies(missing)
}

// -------------------- Install Commands --------------------

// computeInstallCommands computes all the install commands for the dependencies given the system arch.
func computeInstallCommands(arch SystemArch) map[DependencyName]string {
	// Get Docker install command
	cmds := make(map[DependencyName]string)
	switch arch {
	case MacX86, MacARM:
		cmds[Docker] = "brew install --cask docker"
	case LinuxX86, LinuxARM:
		cmds[Docker] = "curl -fsSL https://get.docker.com -o get-docker.sh && sudo sh get-docker.sh && rm get-docker.sh"
	default:
		cmds[Docker] = "N/A"
	}

	// Get Kind install command
	var kindURL string
	kindResp, err := http.Get("https://api.github.com/repos/kubernetes-sigs/kind/releases/latest")
	if err == nil {
		defer kindResp.Body.Close()
		var release struct {
			TagName string `json:"tag_name"`
		}
		if err = json.NewDecoder(kindResp.Body).Decode(&release); err == nil {
			var osName, archName string
			switch arch {
			case MacX86, MacARM:
				osName = "darwin"
			case LinuxX86, LinuxARM:
				osName = "linux"
			default:
				osName = "unknown"
			}
			switch arch {
			case MacX86, LinuxX86:
				archName = "amd64"
			case MacARM, LinuxARM:
				archName = "arm64"
			default:
				archName = "unknown"
			}
			binaryName := fmt.Sprintf("kind-%s-%s", osName, archName)
			kindURL = fmt.Sprintf("https://kind.sigs.k8s.io/dl/%s/%s", release.TagName, binaryName)
		}
	}
	if kindURL == "" {
		var osName, archName string
		switch arch {
		case MacX86, MacARM:
			osName = "darwin"
		case LinuxX86, LinuxARM:
			osName = "linux"
		default:
			osName = "unknown"
		}
		switch arch {
		case MacX86, LinuxX86:
			archName = "amd64"
		case MacARM, LinuxARM:
			archName = "arm64"
		default:
			archName = "unknown"
		}
		binaryName := fmt.Sprintf("kind-%s-%s", osName, archName)
		kindURL = fmt.Sprintf("https://kind.sigs.k8s.io/dl/%s/%s", KindVersion, binaryName)
	}
	cmds[Kind] = fmt.Sprintf("curl -Lo /tmp/kind '%s' && chmod +x /tmp/kind && sudo mv /tmp/kind /usr/local/bin/kind", kindURL)

	// Get Kubectl install command
	var osStr, archStr string
	switch arch {
	case MacX86, MacARM:
		osStr = "darwin"
	case LinuxX86, LinuxARM:
		osStr = "linux"
	default:
		osStr = ""
	}
	switch arch {
	case MacX86, LinuxX86:
		archStr = "amd64"
	case MacARM, LinuxARM:
		archStr = "arm64"
	default:
		archStr = ""
	}
	var kubectlVersion string
	kubectlResp, err := http.Get("https://storage.googleapis.com/kubernetes-release/release/stable.txt")
	if err == nil {
		defer kubectlResp.Body.Close()
		bytes, err := io.ReadAll(kubectlResp.Body)
		if err == nil {
			kubectlVersion = strings.TrimSpace(string(bytes))
		}
	}
	if kubectlVersion == "" {
		kubectlVersion = "latest"
	}
	cmds[Kubectl] = fmt.Sprintf("curl -LO https://storage.googleapis.com/kubernetes-release/release/%s/bin/%s/%s/kubectl && chmod +x kubectl && sudo mv kubectl /usr/local/bin/kubectl", kubectlVersion, osStr, archStr)

	// Get Helm install command
	cmds[Helm] = "curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash"

	// Get Tilt install command
	cmds[Tilt] = "curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash"

	// Get Relay Util install command
	cmds[RelayUtil] = "go install github.com/commoddity/relay-util/v2@latest"

	return cmds
}

// getMissingDependencies compiles a list of dependencies that are not currently installed,
// using the install commands provided.
func getMissingDependencies(arch SystemArch, cmds map[DependencyName]string) []Dependency {
	var missing []Dependency
	for _, dep := range getDependencies(arch, cmds) {
		if !commandExists(dep.Cmd) {
			missing = append(missing, dep)
		}
	}
	return missing
}

// getDependencies returns the full list of dependencies along with metadata.
func getDependencies(arch SystemArch, cmds map[DependencyName]string) []Dependency {
	return []Dependency{
		{
			Name:        Docker,
			Cmd:         "docker",
			Description: "Docker is a container engine that lets you run applications in containers.",
			InstallCmd:  cmds[Docker],
			InstallFunc: func() error { return checkAndInstallDocker(arch, cmds[Docker]) },
		},
		{
			Name:        Kind,
			Cmd:         "kind",
			Description: "Kind creates local Kubernetes clusters using Docker container nodes.",
			InstallCmd:  cmds[Kind],
			InstallFunc: func() error { return checkAndInstallKind(arch, cmds[Kind]) },
		},
		{
			Name:        Kubectl,
			Cmd:         "kubectl",
			Description: "kubectl is the CLI tool for controlling Kubernetes clusters.",
			InstallCmd:  cmds[Kubectl],
			InstallFunc: func() error { return checkAndInstallKubectl(arch, cmds[Kubectl]) },
		},
		{
			Name:        Helm,
			Cmd:         "helm",
			Description: "Helm is a package manager for Kubernetes.",
			InstallCmd:  cmds[Helm],
			InstallFunc: func() error { return checkAndInstallHelm(cmds[Helm]) },
		},
		{
			Name:        Tilt,
			Cmd:         "tilt",
			Description: "Tilt simplifies development on Kubernetes by automating build & deploy cycles.",
			InstallCmd:  cmds[Tilt],
			InstallFunc: func() error { return checkAndInstallTilt(cmds[Tilt]) },
		},
		{
			Name:        RelayUtil,
			Cmd:         "relay-util",
			Description: "Relay Util is a simple load-testing tool for PATH relays.",
			InstallCmd:  cmds[RelayUtil],
			InstallFunc: func() error { return checkAndInstallRelayUtil(cmds[RelayUtil]) },
		},
	}
}

// promptUserToInstall displays the missing dependencies list and prompts the user
// to confirm installation of all missing items.
func promptUserToInstall(missing []Dependency, reader *bufio.Reader) (bool, error) {
	fmt.Println(log.Red + "\n🚨 The following required dependencies are missing:\n" + log.ResetColor)
	for _, dep := range missing {
		fmt.Printf("%s: %s\n", log.Yellow+string(dep.Name)+log.ResetColor, dep.Description)
		cmdToLog := dep.InstallCmd
		if strings.Contains(cmdToLog, "&&") {
			parts := strings.Split(cmdToLog, "&&")
			cmdToLog = strings.TrimSpace(parts[0])
		}
		fmt.Printf(log.Purple+"   Install command: \n"+log.ResetColor+"    %s\n", cmdToLog)
	}
	answer, err := prompt(reader, log.Blue+"\n❔ Would you like to install these dependencies? (y/n): "+log.ResetColor)
	if err != nil {
		return false, err
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer == "y" || answer == "yes" {
		return true, nil
	}
	return false, nil
}

// installMissingDependencies iterates over missing dependencies and runs their install function.
func installMissingDependencies(missing []Dependency) error {
	for _, dep := range missing {
		if err := dep.InstallFunc(); err != nil {
			return fmt.Errorf("failed to install %s: %v", dep.Name, err)
		}
	}
	return nil
}

// commandExists checks if a given command exists in the system's PATH.
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// prompt ensures that the prompt shows `>` on a new line.
func prompt(reader *bufio.Reader, message string) (string, error) {
	fmt.Println(message)
	fmt.Print("> ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// -------------------- Docker --------------------
func checkAndInstallDocker(arch SystemArch, installCmd string) error {
	if commandExists("docker") {
		return nil
	}
	// If using brew on mac, ensure brew exists
	if (arch == MacX86 || arch == MacARM) && strings.Contains(installCmd, "brew") {
		if !commandExists("brew") {
			fmt.Println(log.Yellow + "🚨 Docker not found and Homebrew is missing. Please install Docker Desktop manually from https://www.docker.com/products/docker-desktop" + log.ResetColor)
			return fmt.Errorf("Homebrew is missing. Please install Docker Desktop manually from https://www.docker.com/products/docker-desktop")
		}
	}
	fmt.Println(log.Blue + "🐳 Installing Docker..." + log.ResetColor)
	cmd := exec.Command("sh", "-c", installCmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("❌ failed to install Docker: %v, output: %s", err, string(output))
	}
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		cmd = exec.Command("sudo", "chmod", "666", "/var/run/docker.sock")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to set permissions on docker socket: %v, output: %s", err, string(output))
		}
	}
	if arch == LinuxX86 || arch == LinuxARM {
		if _, err := exec.Command("pgrep", "dockerd").CombinedOutput(); err != nil {
			fmt.Println(log.Yellow + "Docker daemon not running. Attempting to start dockerd..." + log.ResetColor)
			dcmd := exec.Command("dockerd")
			if err := dcmd.Start(); err != nil {
				return fmt.Errorf("failed to start Docker daemon: %v", err)
			}
			time.Sleep(3 * time.Second)
			if _, err := exec.Command("pgrep", "dockerd").CombinedOutput(); err != nil {
				return fmt.Errorf("docker daemon did not start correctly")
			}
			fmt.Println(log.Green + "✅ Docker daemon started successfully." + log.ResetColor)
		}
	}
	fmt.Println(log.Green + "✅ Docker installed successfully." + log.ResetColor)
	versionCmd := exec.Command("docker", "--version")
	versionOutput, err := versionCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s❌ failed to get Docker version: %v%s", log.Red, err, log.ResetColor)
	}
	fmt.Println("   " + log.White + strings.TrimSpace(string(versionOutput)) + log.ResetColor)
	return nil
}

const KindVersion = "v0.27.0"

// -------------------- Kind --------------------
func checkAndInstallKind(arch SystemArch, installCmd string) error {
	if commandExists("kind") {
		return nil
	}
	fmt.Println(log.Blue + "🌀 Installing Kind..." + log.ResetColor)
	cmd := exec.Command("sh", "-c", installCmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install Kind: %v, output: %s", err, string(output))
	}
	fmt.Println(log.Green + "✅ Kind installed successfully." + log.ResetColor)
	versionCmd := exec.Command("kind", "--version")
	versionOutput, err := versionCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s❌ failed to get Kind version: %v%s", log.Red, err, log.ResetColor)
	}
	fmt.Println("   " + log.White + strings.TrimSpace(string(versionOutput)) + log.ResetColor)
	return nil
}

// -------------------- Kubectl --------------------
func checkAndInstallKubectl(arch SystemArch, installCmd string) error {
	if commandExists("kubectl") {
		return nil
	}
	fmt.Println(log.Blue + "🔧 Installing kubectl..." + log.ResetColor)
	cmd := exec.Command("sh", "-c", installCmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install kubectl: %v, output: %s", err, string(output))
	}
	fmt.Println(log.Green + "✅ kubectl installed successfully." + log.ResetColor)
	versionCmd := exec.Command("kubectl", "version", "--client")
	versionOutput, err := versionCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s❌ failed to get kubectl version: %v%s", log.Red, err, log.ResetColor)
	}
	fmt.Println("   " + log.White + strings.TrimSpace(string(versionOutput)) + log.ResetColor)
	return nil
}

// -------------------- Helm --------------------
func checkAndInstallHelm(installCmd string) error {
	if commandExists("helm") {
		return nil
	}
	fmt.Println(log.Blue + "⛵ Installing Helm..." + log.ResetColor)
	cmd := exec.Command("sh", "-c", installCmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("❌ failed to install Helm: %v, output: %s", err, string(output))
	}
	fmt.Println(log.Green + "✅ Helm installed successfully." + log.ResetColor)
	versionCmd := exec.Command("helm", "version", "--short")
	versionOutput, err := versionCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s❌ failed to get Helm version: %v%s", log.Red, err, log.ResetColor)
	}
	fmt.Println("   " + log.White + strings.TrimSpace(string(versionOutput)) + log.ResetColor)
	return nil
}

// -------------------- Tilt --------------------
func checkAndInstallTilt(installCmd string) error {
	if commandExists("tilt") {
		return nil
	}
	fmt.Println(log.Blue + "🚀 Installing Tilt..." + log.ResetColor)
	localBin := os.ExpandEnv("$HOME/.local/bin")
	if _, err := os.Stat(localBin); os.IsNotExist(err) {
		if err := os.MkdirAll(localBin, 0755); err != nil {
			return fmt.Errorf("failed to create local bin directory: %v", err)
		}
	}
	cmd := exec.Command("sh", "-c", installCmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("❌ failed to install Tilt: %v, output: %s", err, string(output))
	}
	fmt.Println(log.Green + "✅ Tilt installed successfully." + log.ResetColor)
	versionCmd := exec.Command("tilt", "version")
	versionOutput, err := versionCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s❌ failed to get Tilt version: %v%s", log.Red, err, log.ResetColor)
	}
	fmt.Println("   " + log.White + strings.TrimSpace(string(versionOutput)) + log.ResetColor)
	return nil
}

// -------------------- Relay Util --------------------
func checkAndInstallRelayUtil(installCmd string) error {
	if commandExists("relay-util") {
		return nil
	}
	if !commandExists("go") {
		fmt.Println(log.Yellow + "🚨 Go is not installed. In order to install Relay Util, please install Go from https://go.dev/doc/install" + log.ResetColor)
		return nil
	}
	fmt.Println(log.Blue + "🚚 Installing Relay Util..." + log.ResetColor)
	cmd := exec.Command("sh", "-c", installCmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("❌ failed to install Relay Util: %v, output: %s", err, string(output))
	}
	fmt.Println(log.Green + "✅ Relay Util installed successfully." + log.ResetColor)
	return nil
}
