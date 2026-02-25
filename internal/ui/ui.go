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

func HardeningSummary(host string, sshPort int) {
	fmt.Println()
	fmt.Printf("  %s\n", bold("SSH access changed:"))
	fmt.Printf("    User:  %s (root login disabled)\n", cyan("bunkr"))
	fmt.Printf("    Port:  %s\n", cyan(fmt.Sprintf("%d", sshPort)))
	fmt.Printf("    Auth:  %s\n", cyan("SSH key only"))
	fmt.Println()
	fmt.Printf("  %s\n", bold("Connect with:"))
	fmt.Printf("    %s\n", green(fmt.Sprintf("ssh bunkr@%s -p %d", host, sshPort)))
	fmt.Println()
	fmt.Printf("  %s\n", bold("For bunkr commands, use:"))
	fmt.Printf("    %s\n", green(fmt.Sprintf("bunkr <command> --on bunkr@%s:%d", host, sshPort)))
	fmt.Println()
}
