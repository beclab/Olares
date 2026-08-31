package webhook

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateMacvlanAnnotationRejectsUnlabeledDirectSelection(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				multusNetworksAnnotation: `[{"name":"underlay-macvlan","namespace":"kube-system"}]`,
			},
		},
	}
	if err := ValidateMacvlanAnnotation(pod); err == nil {
		t.Fatal("expected unlabeled direct macvlan selection to be rejected")
	} else if !strings.Contains(err.Error(), constants.ApplicationMacvlanInitLabel) {
		t.Fatalf("unexpected rejection: %v", err)
	}
}

func TestValidateMacvlanAnnotationAllowsPlatformSelection(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				constants.ApplicationMacvlanInitLabel: "true",
			},
			Annotations: map[string]string{
				multusNetworksAnnotation: `[{"name":"underlay-macvlan","namespace":"kube-system"}]`,
			},
		},
	}
	if err := ValidateMacvlanAnnotation(pod); err != nil {
		t.Fatalf("expected platform selection to be allowed: %v", err)
	}
}

func TestValidateMacvlanAnnotationRejectsMalformedSelection(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				multusNetworksAnnotation: `[{"name":"underlay-macvlan"}`,
			},
		},
	}
	if err := ValidateMacvlanAnnotation(pod); err == nil {
		t.Fatal("expected malformed macvlan selection to be rejected")
	}
}
