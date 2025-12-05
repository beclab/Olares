package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var envNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

// EnvNameToResourceName validates and converts an env name to a k8s resource name.
// - must start with a letter
// - allowed chars: letters, digits, underscore
// - output: lowercase, underscores converted to hyphens
func EnvNameToResourceName(envName string) (string, error) {
	if !envNamePattern.MatchString(envName) {
		return "", fmt.Errorf("invalid env name: must start with a letter and contain only letters, digits, and underscores")
	}

	result := strings.ToLower(envName)
	result = strings.ReplaceAll(result, "_", "-")
	return result, nil
}
