package recipe

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

func GenerateEnv(values map[string]string) []byte {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var lines []string
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("%s=%s", k, values[k]))
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

func ExpandAutoGenerate(env map[string]string) map[string]string {
	result := make(map[string]string, len(env))
	for k, v := range env {
		switch v {
		case "auto_generate_32":
			result[k] = randomHex(16)
		case "auto_generate_64":
			result[k] = randomHex(32)
		default:
			result[k] = v
		}
	}
	return result
}

func MergeValues(promptValues, envValues map[string]string) map[string]string {
	merged := make(map[string]string, len(promptValues)+len(envValues))
	for k, v := range envValues {
		merged[k] = v
	}
	for k, v := range promptValues {
		merged[k] = v
	}
	return merged
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
