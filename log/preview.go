package log

// TODO_TECHDEBT(@adshmh): Refactor as follows:
// 1. Make this value configurable.
// 2. Consider moving this functionality to "github.com/pokt-network/poktroll/pkg/polylog"
//
// defaultMaxLoggedStrLen limits preview string length to prevent log spam.
const defaultMaxLoggedStrLen = 100

// Preview returns a log-safe preview of str.
//
// maxLen is optional and defaults to defaultMaxLoggedStrLen.
// Returns:
//   - Original string if len <= effective max length
//   - Truncated string if len > effective max length
func Preview(str string, maxLen ...int) string {
	l := defaultMaxLoggedStrLen
	if len(maxLen) > 0 {
		l = maxLen[0]
	}
	return previewWithLengthAndEllipsis(str, l)
}

// previewWithLengthAndEllipsis truncates str to the length provided for logging.
// Returns:
//   - Original string if len <= maxLen
//   - Truncated string if len > maxLen
func previewWithLengthAndEllipsis(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	if maxLen <= 3 {
		return str[:maxLen]
	}
	return str[:maxLen-3] + "..."
}
