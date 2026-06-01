package workload

import (
	"strings"
	"testing"
)

func TestCollectWorkloadImageRefsIncludesInitAndAppContainers(t *testing.T) {
	results := []workloadKindResult{{
		Kind: "deployments",
		Items: []Workload{{
			Kind: "Deployment",
			Metadata: WorkloadMetadata{
				Namespace: "default",
				Name:      "api",
			},
			Spec: WorkloadSpec{
				Template: &WorkloadTemplate{
					Spec: WorkloadPodSpec{
						InitContainers: []WorkloadContainer{{Name: "wait", Image: "busybox:1.28"}},
						Containers:     []WorkloadContainer{{Name: "api", Image: "docker.io/example/api:v1"}},
					},
				},
			},
		}},
	}}

	refs := collectWorkloadImageRefs(results)

	if len(refs) != 2 {
		t.Fatalf("refs len = %d, want 2: %#v", len(refs), refs)
	}
	if got, want := refs[0].ContainerType, "initContainer"; got != want {
		t.Fatalf("first container type = %q, want %q", got, want)
	}
	if got, want := refs[0].Image, "busybox:1.28"; got != want {
		t.Fatalf("first image = %q, want %q", got, want)
	}
	if got, want := refs[1].ContainerType, "container"; got != want {
		t.Fatalf("second container type = %q, want %q", got, want)
	}
	if got, want := refs[1].Image, "docker.io/example/api:v1"; got != want {
		t.Fatalf("second image = %q, want %q", got, want)
	}
}

func TestCollectWorkloadImageRefsSkipsMissingImages(t *testing.T) {
	results := []workloadKindResult{{
		Kind: "daemonsets",
		Items: []Workload{{
			Kind: "DaemonSet",
			Metadata: WorkloadMetadata{
				Namespace: "kube-system",
				Name:      "agent",
			},
			Spec: WorkloadSpec{
				Template: &WorkloadTemplate{
					Spec: WorkloadPodSpec{
						Containers: []WorkloadContainer{{Name: "agent"}},
					},
				},
			},
		}},
	}}

	refs := collectWorkloadImageRefs(results)

	if len(refs) != 0 {
		t.Fatalf("refs len = %d, want 0: %#v", len(refs), refs)
	}
}

func TestResolveImageScanKinds(t *testing.T) {
	cases := []struct {
		in          string
		wantPrimary string
		wantBatch   []string
		wantErr     bool
	}{
		{"all", KindAll, []string{"jobs", "cronjobs"}, false},
		{"deploy", "deployments", nil, false},
		{"statefulset", "statefulsets", nil, false},
		{"job", "", []string{"jobs"}, false},
		{"cronjobs", "", []string{"cronjobs"}, false},
		{"bogus", "", nil, true},
	}
	for _, c := range cases {
		primary, batch, err := resolveImageScanKinds(c.in)
		if c.wantErr {
			if err == nil {
				t.Fatalf("resolveImageScanKinds(%q) err = nil, want error", c.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("resolveImageScanKinds(%q) err = %v", c.in, err)
		}
		if primary != c.wantPrimary {
			t.Fatalf("resolveImageScanKinds(%q) primary = %q, want %q", c.in, primary, c.wantPrimary)
		}
		if strings.Join(batch, ",") != strings.Join(c.wantBatch, ",") {
			t.Fatalf("resolveImageScanKinds(%q) batch = %v, want %v", c.in, batch, c.wantBatch)
		}
	}
}

func TestFilterRefsByImageNormalizesAndMatches(t *testing.T) {
	refs := []workloadImageRef{
		{Workload: "web", Image: "docker.io/library/nginx:latest"},
		{Workload: "cache", Image: "docker.io/library/redis:7"},
		{Workload: "edge", Image: "nginx"},
	}

	got := filterRefsByImage(refs, "nginx")

	if len(got) != 2 {
		t.Fatalf("filtered len = %d, want 2: %#v", len(got), got)
	}
	for _, ref := range got {
		if ref.Workload == "cache" {
			t.Fatalf("redis ref leaked into nginx filter: %#v", got)
		}
	}
}
