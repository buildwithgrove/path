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

// displayShannonPreamble prints the color-coded Shannon preamble and waits for user confirmation.
func displayShannonPreamble(reader *bufio.Reader) error {
	preamble := fmt.Sprintf(
		`%s🌿 Configuring PATH gateway for the Pocket Shannon Protocol.🌿%s

%s🚨 IMPORTANT: READ THIS CAREFULLY 🚨%s

Configuring a gateway for the Pocket Shannon Protocol requires the following fields:
 - %s'gateway_address'%s - The gateway_address is the address of the gateway you want to configure.
 - %s'gateway_private_key_hex'%s - The gateway_private_key_hex is the private key of the gateway you want to configure.
 - One or more %s'owned_apps_private_keys_hex'%s - An owned app means an Application delegated to the onchain Gateway.

💡 These fields may be obtained by following the Gateway Quickstart Guide:
%s https://dev.poktroll.com/operate/cheat_sheets/gateway_cheatsheet%s
(⏰ approximate time to complete: 10-15 minutes)

👉 Once you have these fields, proceed to configure PATH on Shannon.`,
		log.Green, log.ResetColor,
		log.Red, log.ResetColor,
		log.Purple, log.ResetColor,
		log.Purple, log.ResetColor,
		log.Purple, log.ResetColor,
		log.Cyan, log.ResetColor,
	)
	fmt.Println(preamble)

	// Use the prompt function to ensure consistent input prompt style
	input, err := prompt(reader, log.Blue+"\nPress 'y' to continue: "+log.ResetColor)
	if err != nil {
		return err
	}
	if strings.ToLower(strings.TrimSpace(input)) != "y" {
		return fmt.Errorf("shannon configuration aborted by user")
	}
	cfgEditor.ClearTerminal()
	return nil
}

// ConfigureShannon performs an interactive configuration for Shannon settings.
// It first displays the preamble then proceeds with prompting for configuration fields.
func ConfigureShannon(conf *config.Config, schema *yaml.Node) error {
	reader := bufio.NewReader(os.Stdin)

	// Display preamble and wait for confirmation.
	if err := displayShannonPreamble(reader); err != nil {
		return err
	}

	configPath := conf.GetPATHConfigFilepath()
	examplePath := conf.GetExamplePATHConfigFilepath("shannon")
	if examplePath == "" {
		return fmt.Errorf("no example config found for shannon")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := copyAndStripComments(examplePath, configPath); err != nil {
			return fmt.Errorf("failed to create shannon config file: %v", err)
		}
		fmt.Printf(log.Green+"✅ Created config file for shannon at '%s'\n"+log.ResetColor, configPath)
	}

	cfgMap, err := loadShannonConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load shannon config: %v", err)
	}

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

	// Clear the terminal only once after all inputs have been collected.
	cfgEditor.ClearTerminal()

	updateShannonConfig(cfgMap, gatewayAddress, gatewayPrivateKey, ownedAppsKeys)

	if err := saveShannonConfig(configPath, cfgMap); err != nil {
		return fmt.Errorf("failed to save shannon config: %v", err)
	}

	fmt.Println("✅ Shannon configuration updated successfully.")
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
			pat := getValueOrEmpty(patNode)
			// If pattern not defined at the current level, check under "items".
			if pat == "" {
				itemsNode := getMappingValue(node, "items")
				if itemsNode != nil {
					itemPatNode := getMappingValue(itemsNode, "pattern")
					pat = getValueOrEmpty(itemPatNode)
				}
			}
			return getValueOrEmpty(descNode), pat
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
		fmt.Println(log.Blue + "🏠 " + description + log.ResetColor)
		input, err := prompt(reader, "Enter the staked Gateway actor's address: ")
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
		input, err := promptHidden(reader, "Enter the staked Gateway actor's private key hex [input hidden]: ")
		if err != nil {
			fmt.Println(log.Red + "❌ Error reading hidden input. Please try again." + log.ResetColor)
			continue
		}
		input = strings.TrimSpace(input)
		matched, err := regexp.MatchString(pattern, input)
		if err != nil || !matched {
			fmt.Println(log.Red + "❌ Input does not match required format. Expected pattern: " + pattern + log.ResetColor)
			continue
		}
		return input, nil
	}
}

// promptOwnedAppsPrivateKeys prompts the user for one key at a time.
// The user is instructed to press Enter without input to finish entering keys.
func promptOwnedAppsPrivateKeys(reader *bufio.Reader, schema *yaml.Node) ([]string, error) {
	fieldPath := "shannon_config.gateway_config.owned_apps_private_keys_hex"
	description, pattern := getFieldDetailsFromSchema(fieldPath, schema)
	fmt.Println(log.Blue + "🔐 " + description + log.ResetColor)

	var keys []string
	for {
		promptMsg := "Enter the private key hex of an Application delegated to the Gateway (or press Enter to finish): "
		if len(keys) > 0 {
			promptMsg = "Enter another delegated Application's private key (or press Enter to finish): "
		}
		key, err := promptHidden(reader, promptMsg)
		if err != nil {
			fmt.Println(log.Red + "❌ Error reading hidden input. Please try again." + log.ResetColor)
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			break
		}
		matched, err := regexp.MatchString(pattern, key)
		if err != nil || !matched {
			fmt.Println(log.Red + "❌ Key does not match required format. Expected pattern: " + pattern + log.ResetColor)
			continue
		}
		keys = append(keys, key)
	}
	return keys, nil
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
