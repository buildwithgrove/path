package filter

import "errors"

type Config struct {
	ParentOnly     string            `json:"parent_only,omitempty"`
	RequestHeaders map[string]string `json:"request_headers,omitempty" envoy:"mergeable,preserve_root"`
	Invalid        bool              `json:"invalid,omitempty" envoy:"mergeable"`
}

func (c *Config) Validate() error {
	if c.Invalid {
		return errors.New("invalid is enabled, hence this error returned")
	}

	return nil
}
