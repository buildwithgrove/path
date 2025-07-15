package log

// TODO_TECHDEBT(@adshmh): Refactor as follows:
// 1. Make this value configurable.
// 2. Consider moving this functionality to "github.com/pokt-network/poktroll/pkg/polylog"
//
// maxLoggedStrLen limits preview string length to prevent log spam.
const maxLoggedStrLen = 100

// Preview truncates str to maxLoggedStrLen for logging.
// Returns:
//   - Original string if len <= maxLoggedStrLen
//   - Truncated string if len > maxLoggedStrLen
func Preview(str string) string {
	return str[:min(maxLoggedStrLen, len(str))]
}
