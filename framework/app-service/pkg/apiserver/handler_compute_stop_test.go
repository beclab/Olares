package apiserver

import (
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
)

func TestClassifyBoundStop(t *testing.T) {
	running := appv1alpha1.ApplicationManagerStatus{
		State:  appv1alpha1.Running,
		OpType: appv1alpha1.InstallOp,
		OpID:   "old-op",
	}
	cases := []struct {
		name   string
		am     *appv1alpha1.ApplicationManager
		class  boundStopClass
		reason string
	}{
		{
			name: "running submits stop",
			am: &appv1alpha1.ApplicationManager{
				Spec:   appv1alpha1.ApplicationManagerSpec{OpType: appv1alpha1.InstallOp},
				Status: running,
			},
			class: boundStopSubmit,
		},
		{
			name: "pending stop is already in flight",
			am: &appv1alpha1.ApplicationManager{
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
					domain.AnnotationOperationRequestID: "pending-stop",
				}},
				Spec:   appv1alpha1.ApplicationManagerSpec{OpType: appv1alpha1.StopOp},
				Status: running,
			},
			class: boundStopAlready,
		},
		{
			name: "unconsumed upgrade blocks the switch",
			am: &appv1alpha1.ApplicationManager{
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
					domain.AnnotationOperationRequestID: "pending-upgrade",
				}},
				Spec:   appv1alpha1.ApplicationManagerSpec{OpType: appv1alpha1.UpgradeOp},
				Status: running,
			},
			class:  boundStopBlocked,
			reason: "unconsumed-operation",
		},
		{
			name: "unconsumed uninstall blocks the switch",
			am: &appv1alpha1.ApplicationManager{
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
					domain.AnnotationOperationRequestID: "pending-uninstall",
				}},
				Spec:   appv1alpha1.ApplicationManagerSpec{OpType: appv1alpha1.UninstallOp},
				Status: running,
			},
			class:  boundStopBlocked,
			reason: "unconsumed-operation",
		},
		{
			name: "unconsumed applyenv blocks the switch",
			am: &appv1alpha1.ApplicationManager{
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
					domain.AnnotationOperationRequestID: "pending-apply",
				}},
				Spec:   appv1alpha1.ApplicationManagerSpec{OpType: appv1alpha1.ApplyEnvOp},
				Status: running,
			},
			class:  boundStopBlocked,
			reason: "unconsumed-operation",
		},
		{
			name: "upgrading is not stoppable",
			am: &appv1alpha1.ApplicationManager{
				Status: appv1alpha1.ApplicationManagerStatus{State: appv1alpha1.Upgrading},
			},
			class:  boundStopBlocked,
			reason: "stop-operation-not-allowed",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			class, reason := classifyBoundStop(tc.am)
			if class != tc.class || reason != tc.reason {
				t.Fatalf("classifyBoundStop = %d, %q, want %d, %q", class, reason, tc.class, tc.reason)
			}
		})
	}
}
