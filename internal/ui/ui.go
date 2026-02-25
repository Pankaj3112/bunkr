// internal/ui/ui.go
package ui

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
)

func Success(msg string) {
	fmt.Printf("  %s %s\n", green("✓"), msg)
}

func Skip(msg string) {
	fmt.Printf("  %s %s\n", yellow("⏭"), msg)
}

func Warn(msg string) {
	fmt.Printf("  %s %s\n", yellow("⚠"), msg)
}

func Error(msg string) {
	fmt.Printf("  %s %s\n", red("✗"), msg)
}

func Info(msg string) {
	fmt.Printf("  %s\n", cyan(msg))
}

func Header(msg string) {
	fmt.Printf("\n  %s\n", bold(msg))
}

func Result(msg string) {
	fmt.Printf("\n  %s\n\n", green(msg))
}
