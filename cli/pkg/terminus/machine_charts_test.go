package terminus

import (
	"path/filepath"
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

func renderChart(t *testing.T, path string) map[string]string {
	t.Helper()
	ch, err := loader.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	values, err := chartutil.ToRenderValues(ch, map[string]interface{}{}, chartutil.ReleaseOptions{Name: "test", Namespace: "os-frontend", IsInstall: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := engine.Render(ch, values)
	if err != nil {
		t.Fatal(err)
	}
	return rendered
}
func TestMachineChartsDoNotCreateFirstUserOrPersonalBindings(t *testing.T) {
	root := "../../../build/base-package/wizard/config"
	for _, name := range []string{"account", "settings"} {
		output := renderChart(t, filepath.Join(root, name))
		for file, manifest := range output {
			for _, forbidden := range []string{"kind: User\n", "kind: GlobalRoleBinding\n", "kind: RoleBinding\n", "name: user-space-test", "namespace: user-system-test", "name: user-space\n", "name: user-system\n", "namespace: user-space\n", "namespace: user-system\n"} {
				if strings.Contains(manifest, forbidden) {
					t.Errorf("%s still contains %q", file, forbidden)
				}
			}
		}
		if name == "settings" {
			cr := output["settings/templates/terminus_cr.yaml"]
			if !strings.Contains(cr, "domainName: ''") {
				t.Fatal("missing empty default domain", cr)
			}
			shared := output["settings/templates/os-frontend-namespace.yaml"]
			if !strings.Contains(shared, "name: os-frontend\n") {
				t.Fatal("shared namespaces missing", shared)
			}
		}
	}

}
