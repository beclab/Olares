package oac_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/oac"
	"sigs.k8s.io/yaml"
)

const testdataFullManifest = "testdata/full_manifest"

// TestLoadAppConfiguration_PreservesAllYAMLFields loads an OlaresManifest.yaml
// that exercises many yaml-tagged fields and asserts that LoadAppConfiguration
// does not drop leaf values when compared after parse → struct → marshal.
//
// Raw manifests often use flat keys (e.g. olaresManifest.version) while
// yaml.Marshal may emit nested maps; structs that only have json tags (e.g.
// TailScale, metav1.LabelSelector) may marshal with different key casing than
// hand-written YAML. normalizeYAMLTreeForCompare removes those false positives
// before leaf-path comparison.
func TestLoadAppConfiguration_PreservesAllYAMLFields(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(testdataFullManifest, oac.ManifestFileName))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var inputDoc any
	if err := yaml.Unmarshal(raw, &inputDoc); err != nil {
		t.Fatalf("unmarshal input yaml: %v", err)
	}
	inputNorm := normalizeYAMLTreeForCompare(inputDoc)
	inputLeaves := make(map[string]any)
	collectYAMLLeaves("", inputNorm, inputLeaves)
	if len(inputLeaves) == 0 {
		t.Fatal("fixture must declare at least one leaf field")
	}

	cfg, err := oac.LoadAppConfiguration(testdataFullManifest)
	if err != nil {
		t.Fatalf("LoadAppConfiguration: %v", err)
	}
	fmt.Printf("-----------\n")
	fmt.Printf("%#v", cfg.Options)

	fmt.Printf("-----------\n")

	if cfg == nil {
		t.Fatal("expected non-nil *AppConfiguration")
	}
	if cfg.ConfigVersion == "" || cfg.ConfigType == "" {
		t.Fatalf("LoadAppConfiguration must populate olaresManifest meta: ConfigVersion=%q ConfigType=%q",
			cfg.ConfigVersion, cfg.ConfigType)
	}

	outBytes, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal loaded cfg: %v", err)
	}
	var outputDoc any
	if err := yaml.Unmarshal(outBytes, &outputDoc); err != nil {
		t.Fatalf("unmarshal output yaml: %v", err)
	}
	outputNorm := normalizeYAMLTreeForCompare(outputDoc)

	var missing, mismatched []string
	for path, want := range inputLeaves {
		got, ok := lookupTestYAMLPath(outputNorm, path)
		if !ok {
			missing = append(missing, path)
			continue
		}
		if !yamlLeafValuesEqual(want, got) {
			mismatched = append(mismatched, fmt.Sprintf("%s: input=%#v output=%#v", path, want, got))
		}
	}

	if len(missing) > 0 || len(mismatched) > 0 {
		sort.Strings(missing)
		sort.Strings(mismatched)
		var b strings.Builder
		b.WriteString("LoadAppConfiguration dropped or altered yaml-tagged fields:\n")
		for _, path := range missing {
			fmt.Fprintf(&b, "  missing after load: %s (input=%#v)\n", path, inputLeaves[path])
		}
		for _, line := range mismatched {
			fmt.Fprintf(&b, "  value mismatch: %s\n", line)
		}
		t.Fatal(b.String())
	}
}

// normalizeYAMLTreeForCompare returns a deep copy of v with:
//   - map[interface{}]interface{} coerced to map[string]any
//   - flat keys "olaresManifest.<subkey>" merged into nested olaresManifest
//   - common json-vs-yaml key spellings unified (subRoutes, matchLabels)
func normalizeYAMLTreeForCompare(v any) any {
	v = yamlToGenericMap(v)
	switch x := v.(type) {
	case map[string]any:
		promoteOlaresManifestFlatKeys(x)
		out := make(map[string]any, len(x))
		for k, vv := range x {
			out[canonicalYAMLMapKey(k)] = normalizeYAMLTreeForCompare(vv)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, el := range x {
			out[i] = normalizeYAMLTreeForCompare(el)
		}
		return out
	default:
		return v
	}
}

func yamlToGenericMap(v any) any {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]any, len(x))
		for k, vv := range x {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			m[ks] = yamlToGenericMap(vv)
		}
		return m
	case map[string]any:
		m := make(map[string]any, len(x))
		for k, vv := range x {
			m[k] = yamlToGenericMap(vv)
		}
		return m
	case []any:
		out := make([]any, len(x))
		for i, el := range x {
			out[i] = yamlToGenericMap(el)
		}
		return out
	default:
		return v
	}
}

const olaresManifestKeyPrefix = "olaresManifest."

// promoteOlaresManifestFlatKeys turns top-level keys like olaresManifest.version
// into olaresManifest: { version: ... } so they match yaml.Marshal output.
func promoteOlaresManifestFlatKeys(m map[string]any) {
	extras := make(map[string]any)
	for k, v := range m {
		if strings.HasPrefix(k, olaresManifestKeyPrefix) && len(k) > len(olaresManifestKeyPrefix) {
			sub := k[len(olaresManifestKeyPrefix):]
			extras[sub] = v
			delete(m, k)
		}
	}
	if len(extras) == 0 {
		return
	}
	if om, ok := m["olaresManifest"].(map[string]any); ok {
		for sk, sv := range extras {
			if _, has := om[sk]; !has {
				om[sk] = sv
			}
		}
	} else {
		m["olaresManifest"] = extras
	}
}

// canonicalYAMLMapKey aligns hand-written YAML keys with what encoding/json-only
// structs typically emit under gopkg.in/yaml.v3.
func canonicalYAMLMapKey(k string) string {
	switch {
	case strings.EqualFold(k, "subroutes"):
		return "subRoutes"
	case strings.EqualFold(k, "matchlabels"):
		return "matchLabels"
	default:
		return k
	}
}

// collectYAMLLeaves records every scalar / leaf collection in a YAML tree using
// dotted paths (e.g. "spec.resources.0.mode", "metadata.name").
func collectYAMLLeaves(prefix string, v any, out map[string]any) {
	switch node := v.(type) {
	case map[string]any:
		for k, child := range node {
			p := k
			if prefix != "" {
				p = prefix + "." + k
			}
			collectYAMLLeaves(p, child, out)
		}
	case map[interface{}]interface{}:
		for k, child := range node {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			p := ks
			if prefix != "" {
				p = prefix + "." + ks
			}
			collectYAMLLeaves(p, child, out)
		}
	case []any:
		for i, child := range node {
			p := fmt.Sprintf("%s.%d", prefix, i)
			collectYAMLLeaves(p, child, out)
		}
	default:
		if prefix != "" {
			out[prefix] = v
		}
	}
}

// lookupTestYAMLPath walks a dotted path. If a map has no single-segment key
// for the next step, it tries the longest composite key formed from the
// remaining path segments (handles flat keys like "olaresManifest.version").
func lookupTestYAMLPath(root any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	cur := root
	for i := 0; i < len(parts); i++ {
		if cur == nil {
			return nil, false
		}
		part := parts[i]
		switch v := cur.(type) {
		case map[string]any:
			if n, ok := v[part]; ok {
				cur = n
				continue
			}
			found := false
			for j := len(parts) - 1; j > i; j-- {
				composite := strings.Join(parts[i:j+1], ".")
				if n, ok := v[composite]; ok {
					cur = n
					i = j
					found = true
					break
				}
			}
			if !found {
				return nil, false
			}
		case map[interface{}]interface{}:
			if n, ok := v[part]; ok {
				cur = n
				continue
			}
			found := false
			for j := len(parts) - 1; j > i; j-- {
				composite := strings.Join(parts[i:j+1], ".")
				if n, ok := v[composite]; ok {
					cur = n
					i = j
					found = true
					break
				}
			}
			if !found {
				return nil, false
			}
		case []any:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(v) {
				return nil, false
			}
			cur = v[idx]
		default:
			return nil, false
		}
	}
	return cur, true
}

func yamlLeafValuesEqual(a, b any) bool {
	return reflect.DeepEqual(normalizeYAMLLeaf(a), normalizeYAMLLeaf(b))
}

func normalizeYAMLLeaf(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case bool:
		return x
	case string:
		return strings.TrimRight(x, "\n")
	case int:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case float32:
		return normalizeFloat(float64(x))
	case float64:
		return normalizeFloat(x)
	case []any:
		out := make([]any, len(x))
		for i, item := range x {
			out[i] = normalizeYAMLLeaf(item)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, item := range x {
			out[k] = normalizeYAMLLeaf(item)
		}
		return out
	case map[interface{}]interface{}:
		out := make(map[string]any, len(x))
		for k, item := range x {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			out[ks] = normalizeYAMLLeaf(item)
		}
		return out
	default:
		return v
	}
}

func normalizeFloat(x float64) any {
	if x == float64(int64(x)) {
		return int64(x)
	}
	return x
}
