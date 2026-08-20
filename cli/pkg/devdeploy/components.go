package devdeploy

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ComponentsFile is the repo-relative path of the component build map.
const ComponentsFile = "build/dev-components.yaml"

// Component is one entry of the map: everything needed to turn a
// component name into a `docker build` and the image tag the released
// charts reference.
type Component struct {
	// Name is the map key, filled in by LoadComponents.
	Name string `yaml:"-"`
	// Image is the repository (no tag) the charts reference, e.g.
	// "beclab/app-service".
	Image string `yaml:"image"`
	// Context is the docker build context, relative to the repo root.
	Context string `yaml:"context"`
	// Dockerfile is relative to the repo root, not to Context — the two
	// differ for most components here (Dockerfile.image, Dockerfile.api,
	// docker/<name>/Dockerfile ...).
	Dockerfile string `yaml:"dockerfile"`
	// Workflow names the publish workflow this entry mirrors, so a
	// reviewer can check the two against each other.
	Workflow string `yaml:"workflow"`
}

type componentsDoc struct {
	Components map[string]Component `yaml:"components"`
}

// LoadComponents parses the component map rooted at repoRoot.
func LoadComponents(repoRoot string) (map[string]Component, error) {
	path := filepath.Join(repoRoot, ComponentsFile)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc componentsDoc
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(doc.Components) == 0 {
		return nil, fmt.Errorf("%s declares no components", path)
	}
	out := make(map[string]Component, len(doc.Components))
	for name, c := range doc.Components {
		c.Name = name
		out[name] = c
	}
	return out, nil
}

// ComponentNames returns the map's keys in sorted order.
func ComponentNames(components map[string]Component) []string {
	out := make([]string, 0, len(components))
	for name := range components {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Lookup resolves one component, with a "did you mean" style error that
// lists what is actually available — the map is long enough that a bare
// "unknown component" would send the reader back to the YAML.
func Lookup(components map[string]Component, name string) (Component, error) {
	c, ok := components[name]
	if ok {
		return c, nil
	}
	var near []string
	for _, candidate := range ComponentNames(components) {
		if strings.Contains(candidate, name) || strings.Contains(name, candidate) {
			near = append(near, candidate)
		}
	}
	if len(near) > 0 {
		return Component{}, fmt.Errorf("unknown component %q; did you mean: %s", name, strings.Join(near, ", "))
	}
	return Component{}, fmt.Errorf("unknown component %q; known components: %s",
		name, strings.Join(ComponentNames(components), ", "))
}

// Validate checks the map against the working tree: every context must
// be a directory and every dockerfile a file.
//
// This is the guard that keeps the map honest. The publish workflows it
// mirrors are only exercised on release, so a Dockerfile that moves in
// a refactor can leave both the workflow and this map pointing at
// nothing for months — as `framework/bfl/Dockerfile.ingress` already
// does in module_bfl_publish_ingress.yaml, which is why that component
// is deliberately absent here.
func Validate(repoRoot string, components map[string]Component) error {
	var problems []string
	for _, name := range ComponentNames(components) {
		c := components[name]
		if strings.TrimSpace(c.Image) == "" {
			problems = append(problems, fmt.Sprintf("%s: image is empty", name))
		}
		if strings.Contains(c.Image, ":") {
			problems = append(problems, fmt.Sprintf("%s: image %q must not carry a tag", name, c.Image))
		}
		if fi, err := os.Stat(filepath.Join(repoRoot, c.Context)); err != nil || !fi.IsDir() {
			problems = append(problems, fmt.Sprintf("%s: context %q is not a directory", name, c.Context))
		}
		if fi, err := os.Stat(filepath.Join(repoRoot, c.Dockerfile)); err != nil || fi.IsDir() {
			problems = append(problems, fmt.Sprintf("%s: dockerfile %q is not a file", name, c.Dockerfile))
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("%s is out of sync with the tree:\n  %s",
			ComponentsFile, strings.Join(problems, "\n  "))
	}
	return nil
}

// FindRepoRoot walks up from start looking for the component map, so
// the dev verbs work from anywhere inside a checkout.
func FindRepoRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ComponentsFile)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not inside an Olares checkout: no %s found in %s or any parent",
				ComponentsFile, start)
		}
		dir = parent
	}
}
