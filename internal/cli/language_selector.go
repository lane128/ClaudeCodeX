package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type selectorKey int

const (
	selectorUnknown selectorKey = iota
	selectorUp
	selectorDown
	selectorTab
	selectorEnter
	selectorInterrupt
)

type languageOption struct {
	Code  string
	Label string
}

var languageOptions = []languageOption{
	{Code: "zh", Label: "中文"},
	{Code: "en", Label: "English"},
}

func runLanguageSelector(input *os.File, output io.Writer, current string) (string, error) {
	restore, err := makeRawTerminal(input)
	if err != nil {
		return "", err
	}
	defer restore()
	hideCursor(output)
	defer showCursor(output)

	selected := defaultLanguageIndex(current)
	reader := bufio.NewReader(input)
	rendered := false

	for {
		renderLanguageSelector(output, selected, rendered)
		rendered = true

		key, err := readSelectorKey(reader)
		if err != nil {
			return "", err
		}

		switch key {
		case selectorUp:
			selected = wrapIndex(selected-1, len(languageOptions))
		case selectorDown, selectorTab:
			selected = wrapIndex(selected+1, len(languageOptions))
		case selectorEnter:
			clearLanguageSelector(output)
			return languageOptions[selected].Code, nil
		case selectorInterrupt:
			clearLanguageSelector(output)
			fmt.Fprintln(output)
			return "", fmt.Errorf("interrupted")
		}
	}
}

func renderLanguageSelector(output io.Writer, selected int, rendered bool) {
	lines := []string{
		renderSelectorLine(languageOptions[0].Label, selected == 0),
		renderSelectorLine(languageOptions[1].Label, selected == 1),
	}

	if rendered {
		fmt.Fprintf(output, "\033[%dA\r", len(lines)-1)
	}
	for i, line := range lines {
		fmt.Fprint(output, "\r\033[2K")
		fmt.Fprint(output, line)
		if i < len(lines)-1 {
			fmt.Fprint(output, "\n")
		}
	}
}

func clearLanguageSelector(output io.Writer) {
	lines := 2
	fmt.Fprintf(output, "\033[%dA\r", lines-1)
	for i := 0; i < lines; i++ {
		fmt.Fprint(output, "\r\033[2K")
		if i < lines-1 {
			fmt.Fprint(output, "\n")
		}
	}
	fmt.Fprintf(output, "\033[%dA\r", lines-1)
}

func renderSelectorLine(label string, selected bool) string {
	if selected {
		return "\033[7m" + label + "\033[0m"
	}
	return label
}

func readSelectorKey(reader *bufio.Reader) (selectorKey, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return selectorUnknown, err
	}

	switch b {
	case 3:
		return selectorInterrupt, nil
	case '\r', '\n':
		return selectorEnter, nil
	case '\t':
		return selectorTab, nil
	case 27:
		next, err := reader.ReadByte()
		if err != nil {
			return selectorUnknown, err
		}
		if next != '[' {
			return selectorUnknown, nil
		}
		arrow, err := reader.ReadByte()
		if err != nil {
			return selectorUnknown, err
		}
		switch arrow {
		case 'A':
			return selectorUp, nil
		case 'B':
			return selectorDown, nil
		default:
			return selectorUnknown, nil
		}
	default:
		return selectorUnknown, nil
	}
}

func defaultLanguageIndex(current string) int {
	for i, option := range languageOptions {
		if option.Code == strings.TrimSpace(current) {
			return i
		}
	}
	return 1
}

func wrapIndex(value, size int) int {
	if size == 0 {
		return 0
	}
	for value < 0 {
		value += size
	}
	return value % size
}

func makeRawTerminal(input *os.File) (func(), error) {
	stateCmd := exec.Command("stty", "-g")
	stateCmd.Stdin = input
	state, err := stateCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("read terminal state: %w", err)
	}

	raw := exec.Command("stty", "raw", "-echo")
	raw.Stdin = input
	if err := raw.Run(); err != nil {
		return nil, fmt.Errorf("enable raw terminal: %w", err)
	}

	original := strings.TrimSpace(string(state))
	return func() {
		restore := exec.Command("stty", original)
		restore.Stdin = input
		_ = restore.Run()
	}, nil
}

func hideCursor(output io.Writer) {
	fmt.Fprint(output, "\033[?25l")
}

func showCursor(output io.Writer) {
	fmt.Fprint(output, "\033[?25h")
}
