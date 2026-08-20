package devdeploy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot locates the checkout from the package directory. Tests run
// with cwd set to the package dir, so this walks up the same way the
// dev verbs do from wherever the user happens to be.
func repoRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root, err := FindRepoRoot(cwd)
	if err != nil {
		t.Fatalf("FindRepoRoot: %v", err)
	}
	return root
}

// The map mirrors .github/workflows/module_*_publish_*.yaml, which only
// run on release — so a Dockerfile that moves during a refactor can
// leave both the workflow and this map pointing at nothing without
// anyone noticing. This test is what turns that into a build failure.
func TestComponentMapMatchesTheTree(t *testing.T) {
	root := repoRoot(t)
	components, err := LoadComponents(root)
	if err != nil {
		t.Fatalf("LoadComponents: %v", err)
	}
	if len(components) == 0 {
		t.Fatal("component map is empty")
	}
	if err := Validate(root, components); err != nil {
		t.Error(err)
	}
}

// Every entry claims to mirror a publish workflow; if that workflow is
// gone the entry has no upstream to be checked against.
func TestComponentWorkflowsExist(t *testing.T) {
	root := repoRoot(t)
	components, err := LoadComponents(root)
	if err != nil {
		t.Fatalf("LoadComponents: %v", err)
	}
	for _, name := range ComponentNames(components) {
		c := components[name]
		if c.Workflow == "" {
			t.Errorf("%s: no workflow recorded", name)
			continue
		}
		path := filepath.Join(root, ".github", "workflows", c.Workflow)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("%s: workflow %s does not exist", name, c.Workflow)
		}
	}
}

func TestValidateRejectsBadEntries(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "framework/thing"), 0o755); err != nil {
		t.Fatal(err)
	}

	tests := map[string]struct {
		component Component
		wantMsg   string
	}{
		"missing dockerfile": {
			component: Component{Image: "beclab/thing", Context: "framework/thing", Dockerfile: "framework/thing/Dockerfile"},
			wantMsg:   "is not a file",
		},
		"missing context": {
			component: Component{Image: "beclab/thing", Context: "framework/gone", Dockerfile: "framework/thing/Dockerfile"},
			wantMsg:   "is not a directory",
		},
		// A tagged image would make `--replaces <image>` ambiguous: the
		// Makefile appends its own tag, and "repo:tag:dev" is nonsense.
		"image carrying a tag": {
			component: Component{Image: "beclab/thing:1.2", Context: "framework/thing", Dockerfile: "framework/thing/Dockerfile"},
			wantMsg:   "must not carry a tag",
		},
		"empty image": {
			component: Component{Context: "framework/thing", Dockerfile: "framework/thing/Dockerfile"},
			wantMsg:   "image is empty",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := Validate(root, map[string]Component{"thing": tc.component})
			if err == nil {
				t.Fatal("expected a validation error")
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("error %q does not mention %q", err, tc.wantMsg)
			}
		})
	}
}

func TestLookupSuggestsNearMisses(t *testing.T) {
	components := map[string]Component{
		"app-service":   {Name: "app-service"},
		"image-service": {Name: "image-service"},
		"bfl":           {Name: "bfl"},
	}

	if _, err := Lookup(components, "app-service"); err != nil {
		t.Fatalf("exact lookup failed: %v", err)
	}

	// A substring match is the common typo (dropping a prefix or
	// suffix), so point at the candidates rather than dumping all 23.
	_, err := Lookup(components, "service")
	if err == nil {
		t.Fatal("expected an error for an unknown component")
	}
	if !strings.Contains(err.Error(), "did you mean") {
		t.Errorf("error %q should suggest near misses", err)
	}
	if !strings.Contains(err.Error(), "app-service") || !strings.Contains(err.Error(), "image-service") {
		t.Errorf("error %q should list both near misses", err)
	}

	_, err = Lookup(components, "zzz")
	if err == nil {
		t.Fatal("expected an error for an unknown component")
	}
	if !strings.Contains(err.Error(), "known components") {
		t.Errorf("error %q should list the known components when nothing is close", err)
	}
}

func TestFindRepoRootFailsOutsideACheckout(t *testing.T) {
	_, err := FindRepoRoot(t.TempDir())
	if err == nil {
		t.Fatal("expected an error outside a checkout")
	}
	if !strings.Contains(err.Error(), ComponentsFile) {
		t.Errorf("error %q should name the file it looked for", err)
	}
}
