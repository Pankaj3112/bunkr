// internal/recipe/prompt.go
package recipe

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func PromptUser(prompts []Prompt) (map[string]string, error) {
	reader := bufio.NewReader(os.Stdin)
	values := make(map[string]string)

	for _, p := range prompts {
		var value string
		var err error

		if p.Secret {
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

func promptSecret(p Prompt) (string, error) {
	fmt.Printf("  ? %s: ", p.Label)
	data, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // newline after hidden input
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
