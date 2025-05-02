package config

// HTTPDataReporterConfig defines settings for HTTP-based data reporting.
// Only JSON-accepting data pipelines are supported as of PR #215
// e.g. Fluentd (HTTP input plugin → BigQuery Output plugin) → BigQuery
type HTTPDataReporterConfig struct {
	// HTTP endpoint for data delivery.
	// Example: Fluentd HTTP input plugin address.
	TargetURL string `yaml:"target_url"`
}
