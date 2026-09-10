package preinstall

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type olaresYAMLFile struct {
	Output olaresYAMLOutput `yaml:"output"`
}

type olaresYAMLOutput struct {
	Containers []olaresYAMLContainer `yaml:"containers"`
}

type olaresYAMLContainer struct {
	Name string `yaml:"name"`
}

// ReadOlaresYAMLContainers returns output.containers[].name from a
// source-tree Olares.yaml. The returned slice is always non-nil so callers
// can distinguish "the file was read" from an unset CheckOptions.OlaresImages.
func ReadOlaresYAMLContainers(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var parsed olaresYAMLFile
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	names := make([]string, 0, len(parsed.Output.Containers))
	for _, c := range parsed.Output.Containers {
		name := strings.TrimSpace(c.Name)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	return names, nil
}
