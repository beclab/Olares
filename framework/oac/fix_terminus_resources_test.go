package oac

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

// TestFixTerminusResourceLimits bumps spec.requiredCpu / spec.requiredMemory when
// summed container requests exceed them. Run explicitly:
//
//	FIX_TERMINUS=1 go test -run TestFixTerminusResourceLimits -timeout 30m -count=1
func TestFixTerminusResourceLimits(t *testing.T) {
	if os.Getenv("FIX_TERMINUS") == "" {
		t.Skip("set FIX_TERMINUS=1 to update manifests")
	}

	root := filepath.Join("testdata", "terminus-apps")
	c := New(WithOwner("alice"), WithAdmin("admin"))

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		dir := filepath.Join(root, e.Name())
		if hasSkipMarker(dir) || isV2ManifestDir(dir) {
			continue
		}
		manifestPath := filepath.Join(dir, "OlaresManifest.yaml")
		raw, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		m, err := c.LoadManifestFile(dir)
		if err != nil {
			t.Logf("%s: load: %v", e.Name(), err)
			continue
		}
		obs := collectContainerObservations(c, m, dir)
		if len(obs) == 0 {
			continue
		}

		var sumCPU, sumMem resource.Quantity
		for _, o := range obs {
			if o.ReqCPU != "" {
				q, err := resource.ParseQuantity(o.ReqCPU)
				if err == nil {
					sumCPU.Add(q)
				}
			}
			if o.ReqMem != "" {
				q, err := resource.ParseQuantity(o.ReqMem)
				if err == nil {
					sumMem.Add(q)
				}
			}
		}

		curCPU, curMem := parseManifestLimitQuantities(string(raw))
		newCPU, newMem := curCPU, curMem
		changed := false
		if !sumCPU.IsZero() && sumCPU.Cmp(curCPU) > 0 {
			newCPU = sumCPU
			changed = true
		}
		if !sumMem.IsZero() && sumMem.Cmp(curMem) > 0 {
			newMem = sumMem
			changed = true
		}
		if !changed {
			continue
		}

		out := patchManifestLimitFields(string(raw), newCPU, newMem)
		t.Logf("%s: requiredCpu %s -> %s, requiredMemory %s -> %s",
			e.Name(), curCPU.String(), newCPU.String(), curMem.String(), newMem.String())
		if err := os.WriteFile(manifestPath, []byte(out), 0o644); err != nil {
			t.Errorf("%s: write: %v", e.Name(), err)
		}
	}
}

func parseManifestLimitQuantities(content string) (cpu, mem resource.Quantity) {
	cpu = resource.MustParse("0")
	mem = resource.MustParse("0")
	for _, line := range strings.Split(content, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "requiredCpu:") {
			v := strings.TrimSpace(strings.TrimPrefix(trim, "requiredCpu:"))
			v = strings.Trim(v, "'\"")
			if q, err := resource.ParseQuantity(v); err == nil {
				cpu = q
			}
		}
		if strings.HasPrefix(trim, "requiredMemory:") {
			v := strings.TrimSpace(strings.TrimPrefix(trim, "requiredMemory:"))
			if q, err := resource.ParseQuantity(v); err == nil {
				mem = q
			}
		}
	}
	return cpu, mem
}

var (
	reRequiredCPU    = regexp.MustCompile(`(?m)^(\s*)requiredCpu:\s*.*$`)
	reRequiredMemory = regexp.MustCompile(`(?m)^(\s*)requiredMemory:\s*.*$`)
)

func patchManifestLimitFields(content string, cpu, mem resource.Quantity) string {
	cpuVal := formatManifestQuantity(cpu)
	memVal := formatManifestQuantity(mem)
	if reRequiredCPU.MatchString(content) {
		content = reRequiredCPU.ReplaceAllString(content, "${1}requiredCpu: "+cpuVal)
	}
	if reRequiredMemory.MatchString(content) {
		content = reRequiredMemory.ReplaceAllString(content, "${1}requiredMemory: "+memVal)
	}
	return content
}

func formatManifestQuantity(q resource.Quantity) string {
	s := q.String()
	// Preserve quoted decimal cores when unambiguous (e.g. "2.575" -> use milli for clarity).
	if strings.HasSuffix(s, "m") || strings.ContainsAny(s, "eE") {
		return s
	}
	if strings.Contains(s, ".") {
		return "'" + s + "'"
	}
	return s
}
