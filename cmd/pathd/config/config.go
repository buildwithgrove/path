package config

import (
	"fmt"
	"io"
	"os"

	_ "embed"

	"gopkg.in/yaml.v3"
)

var (
	// eg. /Users/greg/.path
	PathConfigDir = os.Getenv("HOME") + "/.path"
	// eg. /Users/greg/.path/.pathd.yaml
	ConfigFilePath = PathConfigDir + "/.pathd.yaml"

	pathLocalConfigFilepath      = "/local/path/.config.yaml"
	morseExampleConfigFilepath   = "/config/examples/config.morse_example.yaml"
	shannonExampleConfigFilepath = "/config/examples/config.shannon_example.yaml"
)

// Config represents the configuration for LLM providers and Git.
type Config struct {
	PATHRepo string `yaml:"path_repo"`
}

// LoadConfig loads the configuration from a YAML file.
func LoadConfig() (*Config, error) {
	file, err := os.Open(ConfigFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func ConfigExists() bool {
	_, err := os.Stat(ConfigFilePath)
	return err == nil
}

func SaveConfigToFile(config *Config) error {
	// Create the ~/.pathd directory if it doesn't exist
	err := os.MkdirAll(PathConfigDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFilePath, data, 0644)
}

func (c *Config) GetPATHRepoFilepath() string {
	return c.PATHRepo
}

func (c *Config) GetPATHConfigFilepath() string {
	return c.PATHRepo + pathLocalConfigFilepath
}

func (c *Config) GetExamplePATHConfigFilepath(exampleConfigName string) string {
	switch exampleConfigName {
	case "morse":
		return c.PATHRepo + morseExampleConfigFilepath
	case "shannon":
		return c.PATHRepo + shannonExampleConfigFilepath
	}
	return ""
}

func InitEmptyConfig() error {
	var config Config
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFilePath, data, 0644)
}
