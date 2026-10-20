package chart

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// escapeEnvironmentInterpolation writes a copy of each compose file in which
// every `$` inside a service's `environment` is doubled, so a normal
// (interpolating) load keeps those values verbatim while ports, volumes and
// the other structural fields still resolve `${VAR:-default}`. Loading with
// interpolation off instead leaves `"${PORT:-8088}:8088"` unparsed and the
// compose schema rejects it.
//
// The copies sit next to the originals because kompose resolves relative
// paths (build contexts, env_file, bind mounts) against the file's directory.
func escapeEnvironmentInterpolation(files []string) ([]string, func(), error) {
	var written []string
	cleanup := func() {
		for _, f := range written {
			_ = os.Remove(f)
		}
	}
	out := make([]string, 0, len(files))
	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			cleanup()
			return nil, func() {}, err
		}
		escaped, err := escapeComposeEnvironment(raw)
		if err != nil {
			cleanup()
			return nil, func() {}, fmt.Errorf("%s: %w", path, err)
		}
		tmp, err := os.CreateTemp(filepath.Dir(path), ".olares-compose-*"+filepath.Ext(path))
		if err != nil {
			cleanup()
			return nil, func() {}, err
		}
		written = append(written, tmp.Name())
		if _, err := tmp.Write(escaped); err != nil {
			_ = tmp.Close()
			cleanup()
			return nil, func() {}, err
		}
		if err := tmp.Close(); err != nil {
			cleanup()
			return nil, func() {}, err
		}
		out = append(out, tmp.Name())
	}
	return out, cleanup, nil
}

func escapeComposeEnvironment(raw []byte) ([]byte, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	if len(doc.Content) == 0 {
		return raw, nil
	}
	services := mappingValue(doc.Content[0], "services")
	if services == nil || services.Kind != yaml.MappingNode {
		return raw, nil
	}
	seen := map[*yaml.Node]bool{}
	for i := 1; i < len(services.Content); i += 2 {
		if env := mappingValue(resolveAlias(services.Content[i]), "environment"); env != nil {
			escapeScalars(env, seen)
		}
	}
	return yaml.Marshal(&doc)
}

func mappingValue(n *yaml.Node, key string) *yaml.Node {
	n = resolveAlias(n)
	if n == nil || n.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}
	return nil
}

func resolveAlias(n *yaml.Node) *yaml.Node {
	for n != nil && n.Kind == yaml.AliasNode {
		n = n.Alias
	}
	return n
}

// escapeScalars doubles `$` in every value under n, following aliases and
// `<<` merges so an environment shared through an anchor is escaped once.
func escapeScalars(n *yaml.Node, seen map[*yaml.Node]bool) {
	n = resolveAlias(n)
	if n == nil || seen[n] {
		return
	}
	seen[n] = true
	switch n.Kind {
	case yaml.ScalarNode:
		n.Value = strings.ReplaceAll(n.Value, "$", "$$")
	case yaml.MappingNode:
		for i := 1; i < len(n.Content); i += 2 {
			escapeScalars(n.Content[i], seen)
		}
	case yaml.SequenceNode:
		for _, c := range n.Content {
			escapeScalars(c, seen)
		}
	}
}
