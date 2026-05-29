package workload

import (
	"encoding/json"
	"testing"
)

func TestWorkloadTemplateDecodesContainerImages(t *testing.T) {
	raw := []byte(`{
		"kind": "Deployment",
		"metadata": {"name": "api", "namespace": "default"},
		"spec": {
			"template": {
				"spec": {
					"initContainers": [
						{"name": "wait", "image": "busybox:1.28"}
					],
					"containers": [
						{"name": "api", "image": "docker.io/example/api:v1"}
					]
				}
			}
		}
	}`)

	var w Workload
	if err := json.Unmarshal(raw, &w); err != nil {
		t.Fatalf("decode workload: %v", err)
	}
	if w.Spec.Template == nil {
		t.Fatalf("expected spec.template to be decoded")
	}
	if got, want := w.Spec.Template.Spec.InitContainers[0].Image, "busybox:1.28"; got != want {
		t.Fatalf("init container image = %q, want %q", got, want)
	}
	if got, want := w.Spec.Template.Spec.Containers[0].Image, "docker.io/example/api:v1"; got != want {
		t.Fatalf("container image = %q, want %q", got, want)
	}
}

func TestMaybeStripWorkloadTemplatesKeepsDefaultListOutputLight(t *testing.T) {
	items := []Workload{{
		Kind: "Deployment",
		Spec: WorkloadSpec{
			Template: &WorkloadTemplate{
				Spec: WorkloadPodSpec{
					Containers: []WorkloadContainer{{Name: "api", Image: "docker.io/example/api:v1"}},
				},
			},
		},
	}}

	maybeStripWorkloadTemplates(items, false)

	if items[0].Spec.Template != nil {
		t.Fatalf("default list output kept spec.template; want it stripped (full=false)")
	}
}

func TestMaybeStripWorkloadTemplatesKeepsTemplateWhenFull(t *testing.T) {
	items := []Workload{{
		Kind: "Deployment",
		Spec: WorkloadSpec{
			Template: &WorkloadTemplate{
				Spec: WorkloadPodSpec{
					Containers: []WorkloadContainer{{Name: "api", Image: "docker.io/example/api:v1"}},
				},
			},
		},
	}}

	// full=true is the internal path the `workload images` / `doctor images`
	// verbs use to keep pod-template image fields.
	maybeStripWorkloadTemplates(items, true)

	if items[0].Spec.Template == nil {
		t.Fatalf("full=true stripped spec.template; want container images preserved")
	}
}
