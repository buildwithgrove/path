package config

import (
	"fmt"

	"github.com/buildwithgrove/path/config/utils"
	"gopkg.in/yaml.v3"
)

/* --------------------------------- User Data Config Struct -------------------------------- */

// UserDataConfig contains user data configuration settings.
type UserDataConfig struct {
	DBConnectionString string `yaml:"db_connection_string"`
}

// UnmarshalYAML is a custom unmarshaller for UserDataConfig.
func (c *UserDataConfig) UnmarshalYAML(value *yaml.Node) error {
	// Temp alias to avoid recursion; this is the recommend pattern for Go YAML custom unmarshalers
	type temp UserDataConfig
	var val struct {
		temp `yaml:",inline"`
	}
	if err := value.Decode(&val); err != nil {
		return err
	}
	*c = UserDataConfig(val.temp)
	return c.validate()
}

/* --------------------------------- User Data Config Private Helpers -------------------------------- */

func (c *UserDataConfig) validate() error {
	if !utils.IsValidDBConnectionString(c.DBConnectionString) {
		return fmt.Errorf("invalid DB connection string: %s", c.DBConnectionString)
	}
	return nil
}
