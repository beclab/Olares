package workload

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func templateWith(containers, initContainers []WorkloadContainer, annotations map[string]string) Workload {
	return Workload{
		Metadata: WorkloadMetadata{Name: "app", Namespace: "os-framework", Annotations: annotations},
		Spec: WorkloadSpec{
			Template: &WorkloadTemplate{
				Spec: WorkloadPodSpec{Containers: containers, InitContainers: initContainers},
			},
		},
	}
}

func TestResolveTargetContainer(t *testing.T) {
	one := []WorkloadContainer{{Name: "app", Image: "beclab/app-service:0.6.22"}}
	two := append([]WorkloadContainer{}, one...)
	two = append(two, WorkloadContainer{Name: "sidecar", Image: "envoy:v1"})
	inits := []WorkloadContainer{{Name: "wait-db", Image: "busybox:1"}}

	tests := []struct {
		name       string
		w          Workload
		want       string
		wantType   string
		wantErr    bool
		errSubstr  string
		lookupName string
	}{
		{
			name: "single container defaults without -c",
			w:    templateWith(one, nil, nil),
			want: "app", wantType: "container",
		},
		{
			name:    "multiple containers refuse to guess",
			w:       templateWith(two, nil, nil),
			wantErr: true, errSubstr: "pass -c/--container",
		},
		{
			name: "explicit name picks a sibling",
			w:    templateWith(two, nil, nil), lookupName: "sidecar",
			want: "sidecar", wantType: "container",
		},
		{
			name: "initContainers are addressable by the same flag",
			w:    templateWith(one, inits, nil), lookupName: "wait-db",
			want: "wait-db", wantType: "initContainer",
		},
		{
			name: "unknown name lists what exists",
			w:    templateWith(one, inits, nil), lookupName: "nope",
			wantErr: true, errSubstr: "have: app, wait-db",
		},
		{
			name:    "no pod template is a clear error",
			w:       Workload{Metadata: WorkloadMetadata{Name: "app"}},
			wantErr: true, errSubstr: "no spec.template",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, gotType, err := resolveTargetContainer(tc.w, tc.lookupName)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got container %q", got.Name)
				}
				if tc.errSubstr != "" && !strings.Contains(err.Error(), tc.errSubstr) {
					t.Fatalf("error %q does not mention %q", err, tc.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Name != tc.want {
				t.Errorf("container = %q, want %q", got.Name, tc.want)
			}
			if gotType != tc.wantType {
				t.Errorf("containerType = %q, want %q", gotType, tc.wantType)
			}
		})
	}
}

// The patch has to address the containers list by merge key, not
// replace it: a merge-patch-style whole-array write would delete every
// sibling container in the pod template.
func TestBuildSetImagePatchTargetsOneContainerByName(t *testing.T) {
	w := templateWith([]WorkloadContainer{
		{Name: "app", Image: "beclab/app-service:0.6.22", ImagePullPolicy: "Always"},
		{Name: "sidecar", Image: "envoy:v1"},
	}, nil, nil)

	body, err := buildSetImagePatch(w, w.Spec.Template.Spec.Containers[0], "container", "beclab/app-service:dev", "IfNotPresent")
	if err != nil {
		t.Fatalf("buildSetImagePatch: %v", err)
	}

	spec := body["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})
	list, ok := spec["containers"].([]interface{})
	if !ok {
		t.Fatalf("patch has no containers list: %#v", spec)
	}
	if len(list) != 1 {
		t.Fatalf("patch touches %d containers, want exactly 1 (the merge key does the rest)", len(list))
	}
	entry := list[0].(map[string]interface{})
	if entry["name"] != "app" {
		t.Errorf("patch entry name = %v, want app", entry["name"])
	}
	if entry["image"] != "beclab/app-service:dev" {
		t.Errorf("patch entry image = %v", entry["image"])
	}
	if entry["imagePullPolicy"] != "IfNotPresent" {
		t.Errorf("patch entry imagePullPolicy = %v", entry["imagePullPolicy"])
	}
	if _, present := spec["initContainers"]; present {
		t.Error("patch touches initContainers when the target is a container")
	}
}

func TestBuildSetImagePatchTargetsInitContainerList(t *testing.T) {
	inits := []WorkloadContainer{{Name: "wait-db", Image: "busybox:1"}}
	w := templateWith([]WorkloadContainer{{Name: "app", Image: "x:1"}}, inits, nil)

	body, err := buildSetImagePatch(w, inits[0], "initContainer", "busybox:2", "")
	if err != nil {
		t.Fatalf("buildSetImagePatch: %v", err)
	}
	spec := body["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})
	if _, present := spec["containers"]; present {
		t.Error("patch touches containers when the target is an initContainer")
	}
	list := spec["initContainers"].([]interface{})
	entry := list[0].(map[string]interface{})
	if _, present := entry["imagePullPolicy"]; present {
		t.Error(`empty --pull-policy must leave the policy untouched, but the patch sets it`)
	}
}

// The recorded original must survive a second override. Overwriting it
// would make revert restore a dev image instead of the released one.
func TestBuildSetImagePatchKeepsFirstRecordedOriginal(t *testing.T) {
	existing, err := json.Marshal(map[string]PreviousContainerState{
		"app": {Image: "beclab/app-service:0.6.22", ImagePullPolicy: "Always"},
	})
	if err != nil {
		t.Fatal(err)
	}
	w := templateWith(
		[]WorkloadContainer{{Name: "app", Image: "beclab/app-service:dev", ImagePullPolicy: "IfNotPresent"}},
		nil,
		map[string]string{PreviousImagesAnnotation: string(existing)},
	)

	body, err := buildSetImagePatch(w, w.Spec.Template.Spec.Containers[0], "container", "beclab/app-service:dev2", "IfNotPresent")
	if err != nil {
		t.Fatalf("buildSetImagePatch: %v", err)
	}
	raw := body["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})[PreviousImagesAnnotation].(string)
	got, err := DecodePreviousImages(map[string]string{PreviousImagesAnnotation: raw})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := map[string]PreviousContainerState{
		"app": {Image: "beclab/app-service:0.6.22", ImagePullPolicy: "Always"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("recorded original = %#v, want %#v (the FIRST override is the true pre-dev state)", got, want)
	}
}

// Overriding a second container must not erase the first one's record.
func TestBuildSetImagePatchMergesAnnotationAcrossContainers(t *testing.T) {
	existing, _ := json.Marshal(map[string]PreviousContainerState{
		"app": {Image: "beclab/app-service:0.6.22"},
	})
	w := templateWith([]WorkloadContainer{
		{Name: "app", Image: "beclab/app-service:dev"},
		{Name: "sidecar", Image: "envoy:v1"},
	}, nil, map[string]string{PreviousImagesAnnotation: string(existing)})

	body, err := buildSetImagePatch(w, w.Spec.Template.Spec.Containers[1], "container", "envoy:dev", "IfNotPresent")
	if err != nil {
		t.Fatalf("buildSetImagePatch: %v", err)
	}
	raw := body["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})[PreviousImagesAnnotation].(string)
	got, err := DecodePreviousImages(map[string]string{PreviousImagesAnnotation: raw})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 2 || got["app"].Image != "beclab/app-service:0.6.22" || got["sidecar"].Image != "envoy:v1" {
		t.Errorf("annotation = %#v, want both containers recorded", got)
	}
}

func TestDecodePreviousImages(t *testing.T) {
	t.Run("missing annotation yields an empty non-nil map", func(t *testing.T) {
		got, err := DecodePreviousImages(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Errorf("got %#v, want an empty non-nil map", got)
		}
	})

	// A malformed annotation must not be silently discarded: dropping it
	// would strand the workload on a dev image with no recorded way back.
	t.Run("malformed annotation is an error", func(t *testing.T) {
		_, err := DecodePreviousImages(map[string]string{PreviousImagesAnnotation: "{not json"})
		if err == nil {
			t.Fatal("expected an error for a malformed annotation")
		}
		if !strings.Contains(err.Error(), "malformed") {
			t.Errorf("error %q should say the annotation is malformed", err)
		}
	})
}

func TestImageRefKindPlural(t *testing.T) {
	tests := []struct {
		kind    string
		want    string
		wantErr bool
	}{
		{kind: "StatefulSet", want: "statefulsets"},
		{kind: "Deployment", want: "deployments"},
		{kind: "DaemonSet", want: "daemonsets"},
		// Jobs reference images but their pod templates are immutable,
		// so a fan-out must skip them with an explanation rather than
		// PATCH and get a 422.
		{kind: "Job", wantErr: true},
		{kind: "CronJob", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.kind, func(t *testing.T) {
			got, err := ImageRef{Kind: tc.kind, Namespace: "ns", Workload: "w"}.KindPlural()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error for %s", tc.kind)
				}
				if !strings.Contains(err.Error(), "immutable") {
					t.Errorf("error %q should explain why", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("KindPlural() = %q, want %q", got, tc.want)
			}
		})
	}
}
