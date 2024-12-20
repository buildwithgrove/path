package config

import (
	"fmt"
	"strings"
)

/* --------------------------------- Logger Config Defaults -------------------------------- */

const (
	defaultLogLevel = "debug"
)

/* --------------------------------- Logger Config Struct -------------------------------- */

// LoggerConfig contains logger configuration settings
type LoggerConfig struct {
	// Level sets the minimum log level. Valid values are:
	// "debug", "info", "warn", "error"
	Level string `yaml:"level"`
}

/* --------------------------------- Logger Config Private Helpers -------------------------------- */

// hydrateLoggerDefaults assigns default values to LoggerConfig fields if they are not set
func (c *LoggerConfig) hydrateLoggerDefaults() {
	if c.Level == "" {
		c.Level = defaultLogLevel
	}
}

// Validate ensures the logger configuration is valid
func (c LoggerConfig) Validate() error {
	// polyzero.ParseLevel already handles case-insensitive validation
	// and returns a default level (debug) for invalid inputs
	level := strings.ToLower(c.Level)
	switch level {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("invalid log level: %s", c.Level)
	}
}
