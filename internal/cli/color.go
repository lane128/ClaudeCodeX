package cli

import "os"

const (
	ansiGreen = "\033[32m"
	ansiRed   = "\033[31m"
	ansiReset = "\033[0m"
)

var colorEnabled = isInteractiveTerminal(os.Stdin, os.Stdout)

// colorStatus colors a status string (e.g. "成功", "FAILED") by its value.
func colorStatus(status string) string {
	return colorize(status, status)
}

// colorize colors text using the given status to pick the color and prepends an icon.
func colorize(status, text string) string {
	icon := "❌ "
	color := ansiRed
	if status == "success" {
		icon = "✅ "
		color = ansiGreen
	}
	if !colorEnabled {
		return icon + text
	}
	return icon + color + text + ansiReset
}
