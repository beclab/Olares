package webhook

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestResolveSharedFlags(t *testing.T) {
	mgr := &v1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "engine"},
		Spec:       v1alpha1.ApplicationManagerSpec{AppName: "engine"},
	}
	mgrLabeled := mgr.DeepCopy()
	mgrLabeled.Labels = map[string]string{constants.AppSharedLabel: constants.AppSharedTrue}

	cases := []struct {
		name            string
		ns              *corev1.Namespace
		mgr             *v1alpha1.ApplicationManager
		wantShared      bool
		wantSharedApp   bool
	}{
		{
			name: "v3 shared ns: app-shared without applications name label",
			ns: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name: "engine-shared",
				Labels: map[string]string{
					constants.AppSharedLabel: constants.AppSharedTrue,
				},
			}},
			mgr:           mgr,
			wantShared:    false,
			wantSharedApp: true,
		},
		{
			name: "v2 shared ns: ns-shared + name label",
			ns: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name: "engine-shared",
				Labels: map[string]string{
					"bytetrade.io/ns-shared":       "true",
					constants.ApplicationNameLabel: "engine",
				},
			}},
			mgr:           mgr,
			wantShared:    true,
			wantSharedApp: true,
		},
		{
			name: "manager app-shared label alone",
			ns: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name:   "engine-shared",
				Labels: map[string]string{},
			}},
			mgr:           mgrLabeled,
			wantShared:    false,
			wantSharedApp: true,
		},
		{
			name: "ordinary ns",
			ns: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name: "user-app",
				Labels: map[string]string{
					constants.ApplicationNameLabel: "chat",
				},
			}},
			mgr: &v1alpha1.ApplicationManager{
				Spec: v1alpha1.ApplicationManagerSpec{AppName: "chat"},
			},
			wantShared:    false,
			wantSharedApp: false,
		},
		{
			name:          "nil inputs",
			ns:            nil,
			mgr:           nil,
			wantShared:    false,
			wantSharedApp: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotShared, gotApp := resolveSharedFlags(tc.ns, tc.mgr)
			if gotShared != tc.wantShared || gotApp != tc.wantSharedApp {
				t.Fatalf("resolveSharedFlags() = (%v,%v), want (%v,%v)",
					gotShared, gotApp, tc.wantShared, tc.wantSharedApp)
			}
		})
	}
}

// TestLegacyIsSharedIndependentOfSharedApp documents AC-8: mesh-in uses isSharedApp
// while ws/upload/oes/mesh-out/appkey keep reading legacy isShared. For a v3 shared
// NS the two flags diverge; callers must not swap them.
func TestLegacyIsSharedIndependentOfSharedApp(t *testing.T) {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
		Name: "vllm-shared",
		Labels: map[string]string{
			constants.AppSharedLabel: constants.AppSharedTrue,
			// no applications.app.bytetrade.io/name → legacy isShared stays false
		},
	}}
	mgr := &v1alpha1.ApplicationManager{
		Spec: v1alpha1.ApplicationManagerSpec{AppName: "vllm"},
	}
	isShared, isSharedApp := resolveSharedFlags(ns, mgr)
	if isShared {
		t.Fatal("legacy isShared must stay false for v3 shared without name label")
	}
	if !isSharedApp {
		t.Fatal("isSharedApp must be true from app-shared label")
	}
}
