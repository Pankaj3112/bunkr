// internal/recipe/prompt.go
package recipe

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

func PromptUser(prompts []Prompt) (map[string]string, error) {
	reader := bufio.NewReader(os.Stdin)
	values := make(map[string]string)

	for _, p := range prompts {
		var value string
		var err error

		if len(p.Options) > 0 {
			value, err = promptSelect(reader, p)
		} else if p.Secret {
			value, err = promptSecret(p)
		} else {
			value, err = promptText(reader, p)
		}
		if err != nil {
			return nil, err
		}

		if value == "" && p.Default != "" {
			value = p.Default
		}

		if value == "" && p.Required {
			return nil, fmt.Errorf("%s is required", p.Key)
		}

		values[p.Key] = value
	}

	return values, nil
}

func promptText(reader *bufio.Reader, p Prompt) (string, error) {
	label := p.Label
	if p.Default != "" {
		label = fmt.Sprintf("%s [%s]", label, p.Default)
	}
	fmt.Printf("  ? %s: ", label)

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func promptSelect(reader *bufio.Reader, p Prompt) (string, error) {
	defaultIdx := 0
	for i, opt := range p.Options {
		if opt == p.Default {
			defaultIdx = i + 1
			break
		}
	}

	fmt.Printf("  ? %s:\n", p.Label)
	for i, opt := range p.Options {
		suffix := ""
		if i+1 == defaultIdx {
			suffix = " (default)"
		}
		fmt.Printf("    %d) %s%s\n", i+1, opt, suffix)
	}

	for {
		if defaultIdx > 0 {
			fmt.Printf("  Enter number [%d]: ", defaultIdx)
		} else {
			fmt.Print("  Enter number: ")
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		input = strings.TrimSpace(input)

		if input == "" && defaultIdx > 0 {
			return p.Options[defaultIdx-1], nil
		}

		num, err := strconv.Atoi(input)
		if err != nil || num < 1 || num > len(p.Options) {
			fmt.Printf("  ✗ Please enter a number between 1 and %d.\n", len(p.Options))
			continue
		}

		return p.Options[num-1], nil
	}
}

func promptSecret(p Prompt) (string, error) {
	fmt.Printf("  ? %s: ", p.Label)
	data, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // newline after hidden input
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
