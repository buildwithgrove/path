//go:build e2e

package e2e

import "os"

// isCIEnv returns true if running in a CI environment (CI or GitHub Actions)
func isCIEnv() bool {
	return os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != ""
}
