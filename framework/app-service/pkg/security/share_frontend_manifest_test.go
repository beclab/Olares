package security

import (
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

func TestShareFrontendUsesPlatformNginxConfig(t *testing.T) {
	c, err := loader.Load(filepath.Clean("../../../../apps/.olares/config/user/helm-charts/system-apps"))
	if err != nil {
		t.Fatal(err)
	}
	values, err := chartutil.ToRenderValues(c, map[string]any{}, chartutil.ReleaseOptions{
		Name: "system-apps", Namespace: "user-space-test", IsUpgrade: true,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := engine.Render(c, values)
	if err != nil {
		t.Fatal(err)
	}

	var file []byte
	for _, asset := range c.Files {
		if asset.Name == "files/share.conf" {
			file = asset.Data
		}
	}
	source := string(rendered[c.Name()+"/templates/olares-app.yaml"])
	if source == "" {
		t.Fatal("system frontend manifest was not rendered")
	}
	var deployed, configured bool
	decoder := yaml.NewDecoder(strings.NewReader(source))
	for {
		var item struct {
			Kind     string
			Metadata struct{ Name string }
			Data     map[string]string
			Spec     struct {
				Template struct {
					Metadata struct{ Annotations map[string]string }
					Spec     struct {
						Containers []struct {
							Name         string
							Command      []string
							VolumeMounts []struct {
								Name      string
								MountPath string `yaml:"mountPath"`
								SubPath   string `yaml:"subPath"`
								ReadOnly  bool   `yaml:"readOnly"`
							} `yaml:"volumeMounts"`
						}
						Volumes []struct {
							Name      string
							ConfigMap struct {
								Name  string
								Items []struct{ Key, Path string }
							} `yaml:"configMap"`
						}
					}
				}
			}
		}
		if err := decoder.Decode(&item); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		switch {
		case item.Kind == "Deployment" && item.Metadata.Name == "olares-app-deployment":
			var mounted, projected bool
			for _, c := range item.Spec.Template.Spec.Containers {
				if c.Name != "olares-app" {
					continue
				}
				for _, m := range c.VolumeMounts {
					mounted = mounted || (m.Name == "olares-app-nginx-config" && m.MountPath == "/etc/nginx/conf.d/share.conf" &&
						m.SubPath == "share.conf" && m.ReadOnly)
				}
				if !strings.Contains(strings.Join(c.Command, "\n"), "share.conf) continue ;;") {
					t.Fatal("startup still copies the image's Share configuration over the chart mount")
				}
			}
			for _, v := range item.Spec.Template.Spec.Volumes {
				if v.Name != "olares-app-nginx-config" || v.ConfigMap.Name != "olares-app-nginx-config" {
					continue
				}
				for _, entry := range v.ConfigMap.Items {
					projected = projected || (entry.Key == "share.conf" && entry.Path == "share.conf")
				}
			}
			checksum := item.Spec.Template.Metadata.Annotations["checksum/share-nginx"]
			if checksum != fmt.Sprintf("%x", sha256.Sum256(file)) {
				t.Fatal("Nginx ConfigMap changes must trigger a Pod rollout")
			}
			deployed = mounted && projected
		case item.Kind == "ConfigMap" && item.Metadata.Name == "olares-app-nginx-config":
			got := item.Data["share.conf"]
			if got == "" || got != strings.TrimSuffix(string(file), "\n") {
				t.Fatal("rendered ConfigMap does not carry the checked-in Share configuration")
			}
			for _, blocked := range []string{"location = /index.html { return 404; }", "location = /sharable-link/ { return 404; }"} {
				if !strings.Contains(got, blocked) {
					t.Fatalf("missing Share boundary: %s", blocked)
				}
			}
			if strings.Contains(got, "location /bfl/") || strings.Contains(got, "location /seafhttp/") {
				t.Fatal("Share configuration reopens a legacy proxy prefix")
			}
			configured = true
		}
	}
	if !deployed || !configured {
		t.Fatalf("missing platform Share configuration: deployment=%v configMap=%v", deployed, configured)
	}
}
