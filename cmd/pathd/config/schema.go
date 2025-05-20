package config

import (
	"io"
	"net/http"

	"gopkg.in/yaml.v3"
)

const schemaUrl = "https://raw.githubusercontent.com/buildwithgrove/path/refs/heads/main/config/config.schema.yaml"

// LoadSchema loads the schema from the specified URL.
func LoadSchema() (*yaml.Node, error) {
	resp, err := http.Get(schemaUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the YAML content
	var schemaNode yaml.Node
	if err := yaml.Unmarshal(body, &schemaNode); err != nil {
		return nil, err
	}

	// Ensure that the schemaNode is the mapping node
	if schemaNode.Kind != yaml.MappingNode && len(schemaNode.Content) > 0 {
		schemaNode = *schemaNode.Content[0]
	}

	return &schemaNode, nil
}
