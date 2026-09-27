package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Green  = "\033[92m"
	Cyan   = "\033[96m"
	Yellow = "\033[93m"
	Red    = "\033[91m"
	Dim    = "\033[2m"
)

func Header(title string) {
	fmt.Printf("\n%s%s═══════════════════════════════════════════════════════%s\n", Bold, Cyan, Reset)
	fmt.Printf("       %s%s%s\n", Bold, title, Reset)
	fmt.Printf("%s%s═══════════════════════════════════════════════════════%s\n\n", Bold, Cyan, Reset)
}

func Success(msg string) {
	fmt.Printf("\n%s%s✅ %s%s\n", Bold, Green, msg, Reset)
}

func Error(msg string) {
	fmt.Printf("%s%s❌ %s%s\n", Bold, Red, msg, Reset)
}

func Warn(msg string) {
	fmt.Printf("%s⚠️  %s%s\n", Yellow, msg, Reset)
}

func Prompt(label, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	if defaultValue != "" {
		fmt.Printf("%s [%s%s%s]: ", label, Bold, defaultValue, Reset)
	} else {
		fmt.Printf("%s: ", label)
	}

	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return defaultValue
	}
	return text
}
