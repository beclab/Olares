package upgrade

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsAllowedPlatformEnvoy(t *testing.T) {
	cases := []struct {
		name       string
		pod        corev1.Pod
		c          corev1.Container
		wantAllow  bool
		wantBizOES bool
	}{
		{
			name: "l4",
			pod: corev1.Pod{ObjectMeta: metav1.ObjectMeta{
				Name: "l4-bfl-proxy-0", Labels: map[string]string{"app": "l4-bfl-proxy"},
			}},
			c:          corev1.Container{Name: "proxy", Image: "beclab/l4-bfl-proxy:v0.3.39"},
			wantAllow:  true,
			wantBizOES: false,
		},
		{
			name: "eg data",
			pod: corev1.Pod{ObjectMeta: metav1.ObjectMeta{
				Namespace: "os-gateway", Name: "app-gateway-data-x",
				Labels: map[string]string{"app": "app-gateway-data"},
			}},
			c:          corev1.Container{Name: "envoy", Image: "beclab/envoy:v1.25.11.1"},
			wantAllow:  true,
			wantBizOES: false,
		},
		{
			name: "system-backplane",
			pod: corev1.Pod{ObjectMeta: metav1.ObjectMeta{
				Name: "system-backplane-proxy-abc",
				Labels: map[string]string{
					"app":                    "system-backplane-proxy",
					"envoy.olares.io/role":   "platform-backplane",
					"envoy.olares.io/allowlist": "true",
				},
			}},
			c:          corev1.Container{Name: "proxy", Image: "beclab/envoy:v1.25.11.1"},
			wantAllow:  true,
			wantBizOES: false,
		},
		{
			name: "legacy system-server proxy",
			pod: corev1.Pod{ObjectMeta: metav1.ObjectMeta{
				Name: "system-server-xyz", Labels: map[string]string{"app": "systemserver"},
			}},
			c:          corev1.Container{Name: "proxy", Image: "beclab/envoy:v1.25.11.1"},
			wantAllow:  true,
			wantBizOES: false,
		},
		{
			name: "business oes",
			pod: corev1.Pod{ObjectMeta: metav1.ObjectMeta{
				Namespace: "user-space-alice", Name: "files-0",
			}},
			c:          corev1.Container{Name: "olares-envoy-sidecar", Image: "beclab/envoy:v1.25.11.1"},
			wantAllow:  false,
			wantBizOES: true,
		},
		{
			name: "business envoy image other name",
			pod: corev1.Pod{ObjectMeta: metav1.ObjectMeta{
				Namespace: "user-space-alice", Name: "desk-0",
			}},
			c:          corev1.Container{Name: "sidecar", Image: "beclab/envoy:v1.25.11.1"},
			wantAllow:  false,
			wantBizOES: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAllowedPlatformEnvoy(tc.pod, tc.c); got != tc.wantAllow {
				t.Fatalf("isAllowedPlatformEnvoy=%v want %v", got, tc.wantAllow)
			}
			if got := isBusinessOESContainer(tc.pod, tc.c); got != tc.wantBizOES {
				t.Fatalf("isBusinessOESContainer=%v want %v", got, tc.wantBizOES)
			}
		})
	}
}

func TestAllowlistLabelShortcut(t *testing.T) {
	pod := corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name: "custom-platform-envoy",
		Labels: map[string]string{
			"envoy.olares.io/allowlist": "true",
		},
	}}
	c := corev1.Container{Name: "proxy", Image: "beclab/envoy:v1.25.11.1"}
	if !isAllowedPlatformEnvoy(pod, c) {
		t.Fatal("envoy.olares.io/allowlist=true must allow platform envoy")
	}
	if isBusinessOESContainer(pod, c) {
		t.Fatal("allowlisted envoy must not count as business oes")
	}
}
