package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	cfgEditor "github.com/buildwithgrove/gdi/config/editor"
	"github.com/buildwithgrove/gdi/log"
	"gopkg.in/yaml.v3"

	"github.com/buildwithgrove/path/cmd/pathd/config"
)

// ConfigureShannon performs an interactive configuration for Shannon settings.
// It loads the schema, then prompts for gateway_address, gateway_private_key_hex,
// and owned_apps_private_keys_hex using descriptions and regex patterns extracted
// directly from the schema.
func ConfigureShannon(conf *config.Config, schema *yaml.Node) error {
	reader := bufio.NewReader(os.Stdin)

	// Determine the path to the Shannon config file.
	configPath := conf.GetPATHConfigFilepath()

	// Load existing configuration, if any.
	cfgMap, err := loadShannonConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load shannon config: %v", err)
	}

	// Prompt the user for each field.
	gatewayAddress, err := promptGatewayAddress(reader, schema)
	if err != nil {
		return err
	}

	gatewayPrivateKey, err := promptGatewayPrivateKey(reader, schema)
	if err != nil {
		return err
	}

	ownedAppsKeys, err := promptOwnedAppsPrivateKeys(reader, schema)
	if err != nil {
		return err
	}

	// Update configuration map with new values.
	updateShannonConfig(cfgMap, gatewayAddress, gatewayPrivateKey, ownedAppsKeys)

	// Save the updated configuration back to the file.
	if err := saveShannonConfig(configPath, cfgMap); err != nil {
		return fmt.Errorf("failed to save shannon config: %v", err)
	}

	fmt.Println(log.Green + "✅ Shannon configuration updated successfully." + log.ResetColor)
	return nil
}

// getFieldDetailsFromSchema traverses the given schema node using the provided dot-delimited fieldPath
// and returns the description and regex pattern for that field, sourcing the regex from the schema's "pattern" field.
func getFieldDetailsFromSchema(fieldPath string, schema *yaml.Node) (description, pattern string) {
	parts := strings.Split(fieldPath, ".")
	props := getMappingValue(schema, "properties")
	if props == nil {
		return "", ""
	}

	current := props
	for i, part := range parts {
		node := getMappingValue(current, part)
		if node == nil {
			return "", ""
		}
		if i == len(parts)-1 {
			descNode := getMappingValue(node, "description")
			patNode := getMappingValue(node, "pattern")
			return getValueOrEmpty(descNode), getValueOrEmpty(patNode)
		}
		next := getMappingValue(node, "properties")
		if next == nil {
			return "", ""
		}
		current = next
	}
	return "", ""
}

// promptGatewayAddress prompts the user for gateway_address using schema-sourced details.
// It reprompts if the input does not pass regex validation.
func promptGatewayAddress(reader *bufio.Reader, schema *yaml.Node) (string, error) {
	fieldPath := "shannon_config.gateway_config.gateway_address"
	description, pattern := getFieldDetailsFromSchema(fieldPath, schema)
	for {
		fmt.Println(log.Blue + "🔑 " + description + log.ResetColor)
		fmt.Print(log.Blue + "📝 Gateway Address: " + log.ResetColor)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(log.Red + "❌ Error reading input. Please try again." + log.ResetColor)
			continue
		}
		input = strings.TrimSpace(input)
		matched, err := regexp.MatchString(pattern, input)
		if err != nil || !matched {
			fmt.Println(log.Red + "❌ Input does not match required format. Expected pattern: " + pattern + log.ResetColor)
			continue
		}
		cfgEditor.ClearTerminal()
		return input, nil
	}
}

// promptGatewayPrivateKey prompts for gateway_private_key_hex using schema-sourced details.
// It masks the input and reprompts the user until the input satisfies the regex pattern.
func promptGatewayPrivateKey(reader *bufio.Reader, schema *yaml.Node) (string, error) {
	fieldPath := "shannon_config.gateway_config.gateway_private_key_hex"
	description, pattern := getFieldDetailsFromSchema(fieldPath, schema)
	for {
		fmt.Println(log.Blue + "🔒 " + description + log.ResetColor)
		// Mask the input using ReadHiddenInput
		input := cfgEditor.ReadHiddenInput(log.Blue + "📝 Gateway Private Key (hex): " + log.ResetColor)
		input = strings.TrimSpace(input)
		matched, err := regexp.MatchString(pattern, input)
		if err != nil || !matched {
			fmt.Println(log.Red + "❌ Input does not match required format. Expected pattern: " + pattern + log.ResetColor)
			continue
		}
		cfgEditor.ClearTerminal()
		return input, nil
	}
}

// promptOwnedAppsPrivateKeys prompts for owned_apps_private_keys_hex using schema-sourced details.
// It masks the input and reprompts until all provided keys match the regex pattern.
func promptOwnedAppsPrivateKeys(reader *bufio.Reader, schema *yaml.Node) ([]string, error) {
	fieldPath := "shannon_config.gateway_config.owned_apps_private_keys_hex"
	description, pattern := getFieldDetailsFromSchema(fieldPath, schema)
	for {
		fmt.Println(log.Blue + "📋 " + description + log.ResetColor)
		// Mask the input for sensitive keys.
		input := cfgEditor.ReadHiddenInput(log.Blue + "📝 Owned Apps Private Keys (hex, comma-separated): " + log.ResetColor)
		input = strings.TrimSpace(input)
		cfgEditor.ClearTerminal()
		if input == "" {
			return nil, nil
		}
		keys := strings.Split(input, ",")
		valid := true
		for i := range keys {
			keys[i] = strings.TrimSpace(keys[i])
			matched, err := regexp.MatchString(pattern, keys[i])
			if err != nil || !matched {
				fmt.Println(log.Red + "❌ One or more keys do not match required format. Expected pattern: " + pattern + log.ResetColor)
				valid = false
				break
			}
		}
		if valid {
			return keys, nil
		}
	}
}

// updateShannonConfig updates the shannon configuration in the provided map with the given values.
func updateShannonConfig(cfgMap map[string]interface{}, gatewayAddress, gatewayPrivateKey string, ownedAppsKeys []string) {
	shannonConfig, ok := cfgMap["shannon_config"].(map[string]interface{})
	if !ok {
		shannonConfig = make(map[string]interface{})
		cfgMap["shannon_config"] = shannonConfig
	}

	gatewayConfig, ok := shannonConfig["gateway_config"].(map[string]interface{})
	if !ok {
		gatewayConfig = make(map[string]interface{})
		shannonConfig["gateway_config"] = gatewayConfig
	}

	gatewayConfig["gateway_address"] = gatewayAddress
	gatewayConfig["gateway_private_key_hex"] = gatewayPrivateKey
	gatewayConfig["owned_apps_private_keys_hex"] = ownedAppsKeys
}

// --- Helper functions for schema traversal ---

// getMappingValue returns the value node corresponding to a key in a mapping node.
func getMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

// getValueOrEmpty returns the Value of the node or an empty string if the node is nil.
func getValueOrEmpty(node *yaml.Node) string {
	if node == nil {
		return ""
	}
	return node.Value
}

// loadShannonConfig loads the YAML configuration from the provided file path.
// If the file does not exist, an empty map is returned.
func loadShannonConfig(filePath string) (map[string]interface{}, error) {
	cfgMap := make(map[string]interface{})
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfgMap, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, &cfgMap); err != nil {
		return nil, err
	}
	return cfgMap, nil
}

// saveShannonConfig saves the configuration map to the specified YAML file.
func saveShannonConfig(filePath string, cfgMap map[string]interface{}) error {
	data, err := yaml.Marshal(cfgMap)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}
