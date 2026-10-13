package apiserver

import (
	"context"
	"encoding/json"
	"testing"

	wfv1alpha1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/webhook"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/google/uuid"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Admission must retain the image registry identity. Rewriting it to a mirror
// would bypass namespace-specific credentials and fallback at the scheduled node.
func TestCronWorkflowRetainsRegistryNamespace(t *testing.T) {
	refs := []string{
		"docker.io/library/busybox:1.37",
		"ghcr.io/acme/app:v1",
		"gitlab.example.com:5000/team/project/app:v1",
		"gitea.example.com/alice/app@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"busybox:1.37",
	}
	wf := wfv1alpha1.CronWorkflow{}
	for _, ref := range refs {
		wf.Spec.WorkflowSpec.Templates = append(wf.Spec.WorkflowSpec.Templates, wfv1alpha1.Template{
			Container: &corev1.Container{Image: ref},
		})
	}
	original, err := json.Marshal(wf)
	if err != nil {
		t.Fatal(err)
	}
	h := &Handler{sidecarWebhook: &webhook.Webhook{}}
	response := h.cronWorkflowMutate(context.Background(), &admissionv1.AdmissionRequest{
		UID: "test", Object: runtime.RawExtension{Raw: original},
	}, uuid.New())
	if !response.Allowed {
		t.Fatal("workflow rejected")
	}
	patch, err := jsonpatch.DecodePatch(response.Patch)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := patch.Apply(original)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(updated, &wf); err != nil {
		t.Fatal(err)
	}
	for i, ref := range refs {
		want := ref
		if i == len(refs)-1 {
			want = "docker.io/library/busybox:1.37"
		}
		if got := wf.Spec.WorkflowSpec.Templates[i].Container.Image; got != want {
			t.Errorf("image %q became %q, want %q", ref, got, want)
		}
	}
}
