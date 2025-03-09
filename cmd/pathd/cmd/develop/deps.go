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

// Dependency represents an external dependency required to run PATH in development mode.
type Dependency struct {
	Name        string
	Cmd         string
	Description string
	InstallCmd  string
	InstallFunc func() error
}

// getDependencies returns the full list of dependencies along with metadata.
func getDependencies() []Dependency {
	var dockerInstallCmd string
	if runtime.GOOS == "darwin" {
		dockerInstallCmd = "brew install --cask docker"
	} else if runtime.GOOS == "linux" {
		dockerInstallCmd = "curl -fsSL https://get.docker.com -o get-docker.sh && sh get-docker.sh"
	} else {
		dockerInstallCmd = "N/A"
	}

	kindInstallCmd := "curl -Lo /tmp/kind <latest_release_url> && chmod +x /tmp/kind && mv /tmp/kind /usr/local/bin/kind"
	kubectlInstallCmd := "curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/<os>/<arch>/kubectl && chmod +x kubectl && mv kubectl /usr/local/bin/kubectl"

	return []Dependency{
		{
			Name:        "🐳 Docker",
			Cmd:         "docker",
			Description: "Docker is a container engine that lets you run applications in containers.",
			InstallCmd:  dockerInstallCmd,
			InstallFunc: checkAndInstallDocker,
		},
		{
			Name:        "🌀 Kind",
			Cmd:         "kind",
			Description: "Kind creates local Kubernetes clusters using Docker container nodes.",
			InstallCmd:  kindInstallCmd,
			InstallFunc: checkAndInstallKind,
		},
		{
			Name:        "🔧 kubectl",
			Cmd:         "kubectl",
			Description: "kubectl is the CLI tool for controlling Kubernetes clusters.",
			InstallCmd:  kubectlInstallCmd,
			InstallFunc: checkAndInstallKubectl,
		},
		{
			Name:        "⛵ Helm",
			Cmd:         "helm",
			Description: "Helm is a package manager for Kubernetes.",
			InstallCmd:  "curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash",
			InstallFunc: checkAndInstallHelm,
		},
		{
			Name:        "🚀 Tilt",
			Cmd:         "tilt",
			Description: "Tilt simplifies development on Kubernetes by automating build & deploy cycles.",
			InstallCmd:  "curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash",
			InstallFunc: checkAndInstallTilt,
		},
		{
			Name:        "🚚 Relay Util",
			Cmd:         "relay-util",
			Description: "Relay Util is a simple load-testing tool for PATH relays.",
			InstallCmd:  "go install github.com/commoddity/relay-util/v2@latest",
			InstallFunc: checkAndInstallRelayUtil,
		},
	}
}

// getMissingDependencies compiles a list of dependencies that are not currently installed.
func getMissingDependencies() []Dependency {
	var missing []Dependency
	for _, dep := range getDependencies() {
		if !commandExists(dep.Cmd) {
			missing = append(missing, dep)
		}
	}
	return missing
}

// promptUserToInstall displays the missing dependencies list and prompts the user
// to confirm installation of all missing items.
func promptUserToInstall(missing []Dependency, reader *bufio.Reader) (bool, error) {
	fmt.Println(log.Red + "\n🚨 The following required dependencies are missing:" + log.ResetColor)
	for _, dep := range missing {
		fmt.Printf("%s: %s\n", log.Yellow+dep.Name+log.ResetColor, dep.Description)
		fmt.Printf("   Install command: %s\n", log.Green+dep.InstallCmd+log.ResetColor)
	}

	answer, err := prompt(reader, log.Blue+"\n❔ Would you like to install these dependencies? (y/n): "+log.ResetColor)
	if err != nil {
		return false, err
	}

	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer == "y" || answer == "yes" {
		return true, nil
	} else {
		return false, nil
	}
}

// installMissingDependencies iterates over missing dependencies and runs their install function.
func installMissingDependencies(missing []Dependency) error {
	for _, dep := range missing {
		fmt.Printf(log.Blue+"Installing %s...\n"+log.ResetColor, dep.Name)
		if err := dep.InstallFunc(); err != nil {
			return fmt.Errorf("failed to install %s: %v", dep.Name, err)
		}
	}
	return nil
}

// checkAndInstallDependencies first checks for missing dependencies,
// prompts the user to install them, and if agreed, installs them.
func checkAndInstallDependencies(reader *bufio.Reader) error {
	missing := getMissingDependencies()
	if len(missing) == 0 {
		return nil
	}

	install, err := promptUserToInstall(missing, reader)
	if err != nil {
		return err
	}

	if install {
		if err := installMissingDependencies(missing); err != nil {
			return err
		}
	} else {
		fmt.Println(log.Yellow + "👋 Exiting without installing required dependencies." + log.ResetColor)
		os.Exit(0)
	}

	return nil
}

// commandExists checks if a given command exists in the system's PATH.
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func checkAndInstallDocker() error {
	if commandExists("docker") {
		return nil
	}

	osType := runtime.GOOS
	if osType == "darwin" {
		if !commandExists("brew") {
			fmt.Println(log.Yellow + "⚠️ Docker not found and Homebrew is missing. Please install Docker Desktop manually from https://www.docker.com/products/docker-desktop" + log.ResetColor)
			return nil
		}
		fmt.Println(log.Blue + "🐳 Installing Docker Desktop via Homebrew..." + log.ResetColor)
		cmd := exec.Command("brew", "install", "--cask", "docker")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("❌ failed to install Docker via Homebrew: %v, output: %s", err, string(output))
		}
		fmt.Println(log.Green + "✅ Docker installed successfully." + log.ResetColor)
	} else if osType == "linux" {
		fmt.Println(log.Blue + "🐳 Installing Docker using the official install script..." + log.ResetColor)
		cmd := exec.Command("wget", "-qO", "get-docker.sh", "https://get.docker.com")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to download Docker install script using wget: %v, output: %s", err, string(output))
		}
		cmd = exec.Command("sh", "get-docker.sh")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to run Docker install script: %v, output: %s", err, string(output))
		}
		os.Remove("get-docker.sh")
		// Ensure the docker socket has appropriate permissions
		if _, err := os.Stat("/var/run/docker.sock"); err == nil {
			cmd = exec.Command("chmod", "666", "/var/run/docker.sock")
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to set permissions on docker socket: %v, output: %s", err, string(output))
			}
		}
		// Check if the Docker daemon is running; if not, attempt to start it
		if _, err := exec.Command("pgrep", "dockerd").CombinedOutput(); err != nil {
			fmt.Println(log.Yellow + "Docker daemon not running. Attempting to start dockerd..." + log.ResetColor)
			dcmd := exec.Command("dockerd")
			if err := dcmd.Start(); err != nil {
				return fmt.Errorf("failed to start Docker daemon: %v", err)
			}
			// Wait a few seconds for the daemon to initialize
			time.Sleep(3 * time.Second)
			if _, err := exec.Command("pgrep", "dockerd").CombinedOutput(); err != nil {
				return fmt.Errorf("docker daemon did not start correctly")
			}
			fmt.Println(log.Green + "Docker daemon started successfully." + log.ResetColor)
		}
		fmt.Println(log.Green + "✅ Docker installed successfully." + log.ResetColor)
	} else {
		return fmt.Errorf("unsupported OS for Docker installation: %s", osType)
	}
	return nil
}

func checkAndInstallKind() error {
	if commandExists("kind") {
		return nil
	}

	osType := runtime.GOOS
	var binaryName string
	if osType == "darwin" {
		binaryName = "kind-darwin-amd64"
	} else if osType == "linux" {
		binaryName = "kind-linux-amd64"
	} else {
		return fmt.Errorf("unsupported OS for Kind installation: %s", osType)
	}

	resp, err := http.Get("https://api.github.com/repos/kubernetes-sigs/kind/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to fetch Kind release info: %v", err)
	}
	defer resp.Body.Close()
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to decode Kind release info: %v", err)
	}
	version := release.TagName
	downloadURL := fmt.Sprintf("https://kind.sigs.k8s.io/dl/%s/%s", version, binaryName)

	fmt.Println(log.Blue + "🌀 Installing Kind..." + log.ResetColor)
	tmpFile := "/tmp/kind"
	cmd := exec.Command("wget", "-qO", tmpFile, downloadURL)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to download Kind binary: %v, output: %s", err, string(output))
	}
	if err := os.Chmod(tmpFile, 0755); err != nil {
		return fmt.Errorf("failed to chmod Kind binary: %v", err)
	}
	destPath := "/usr/local/bin/kind"
	if err := os.Rename(tmpFile, destPath); err != nil {
		in, err := os.Open(tmpFile)
		if err != nil {
			return fmt.Errorf("failed to open temporary Kind binary: %v", err)
		}
		defer in.Close()
		out, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create destination for Kind binary: %v", err)
		}
		defer out.Close()
		if _, err = io.Copy(out, in); err != nil {
			return fmt.Errorf("failed to copy Kind binary: %v", err)
		}
		os.Remove(tmpFile)
	}
	fmt.Println(log.Green + "✅ Kind installed successfully." + log.ResetColor)
	return nil
}

func checkAndInstallKubectl() error {
	if commandExists("kubectl") {
		return nil
	}

	osType := runtime.GOOS
	var osPath string
	if osType == "darwin" {
		osPath = "darwin"
	} else if osType == "linux" {
		osPath = "linux"
	} else {
		return fmt.Errorf("unsupported OS for kubectl installation: %s", osType)
	}
	arch := "amd64"
	resp, err := http.Get("https://storage.googleapis.com/kubernetes-release/release/stable.txt")
	if err != nil {
		return fmt.Errorf("failed to get kubectl version: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read kubectl version: %v", err)
	}
	version := strings.TrimSpace(string(body))
	downloadURL := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/%s/%s/kubectl", version, osPath, arch)

	fmt.Println(log.Blue + "🔧 Installing kubectl..." + log.ResetColor)
	tmpFile := "/tmp/kubectl"
	cmd := exec.Command("wget", "-qO", tmpFile, downloadURL)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to download kubectl: %v, output: %s", err, string(output))
	}
	if err := os.Chmod(tmpFile, 0755); err != nil {
		return fmt.Errorf("failed to chmod kubectl binary: %v", err)
	}
	destPath := "/usr/local/bin/kubectl"
	if err := os.Rename(tmpFile, destPath); err != nil {
		in, err := os.Open(tmpFile)
		if err != nil {
			return fmt.Errorf("failed to open temporary kubectl binary: %v", err)
		}
		defer in.Close()
		out, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create destination for kubectl binary: %v", err)
		}
		defer out.Close()
		if _, err = io.Copy(out, in); err != nil {
			return fmt.Errorf("failed to copy kubectl binary: %v", err)
		}
		os.Remove(tmpFile)
	}
	fmt.Println(log.Green + "✅ kubectl installed successfully." + log.ResetColor)
	return nil
}

func checkAndInstallHelm() error {
	if commandExists("helm") {
		return nil
	}
	fmt.Println(log.Blue + "⛵ Installing Helm..." + log.ResetColor)
	cmd := exec.Command("sh", "-c", "wget -qO- https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("❌ failed to install Helm: %v, output: %s", err, string(output))
	}
	fmt.Println(log.Green + "✅ Helm installed successfully." + log.ResetColor)
	return nil
}

func checkAndInstallTilt() error {
	if commandExists("tilt") {
		return nil
	}
	fmt.Println(log.Blue + "🚀 Installing Tilt..." + log.ResetColor)

	// Ensure the local bin directory exists
	localBin := os.ExpandEnv("$HOME/.local/bin")
	if _, err := os.Stat(localBin); os.IsNotExist(err) {
		if err := os.MkdirAll(localBin, 0755); err != nil {
			return fmt.Errorf("failed to create local bin directory: %v", err)
		}
	}

	// Update PATH to include the local bin and run the Tilt installer with NO_SUDO=1
	cmd := exec.Command("sh", "-c", "export PATH="+localBin+":$PATH && wget -qO- https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("❌ failed to install Tilt: %v, output: %s", err, string(output))
	}
	fmt.Println(log.Green + "✅ Tilt installed successfully." + log.ResetColor)
	return nil
}

func checkAndInstallRelayUtil() error {
	if commandExists("relay-util") {
		return nil
	}
	fmt.Println(log.Blue + "🚚 Installing Relay Util..." + log.ResetColor)
	cmd := exec.Command("go", "install", "github.com/commoddity/relay-util/v2@latest")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("❌ failed to install Relay Util: %v, output: %s", err, string(output))
	}
	fmt.Println(log.Green + "✅ Relay Util installed successfully." + log.ResetColor)
	return nil
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
