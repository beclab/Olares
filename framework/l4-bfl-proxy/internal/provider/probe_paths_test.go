package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUniqueHTTPProbePaths(t *testing.T) {
	got := uniqueHTTPProbePaths([]string{"/healthz", "ready", "/", "", "/healthz", " /live "})
	assert.Equal(t, []string{"/healthz", "/ready", "/live"}, got)
}

func TestHTTPProbePathsFromPod(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "p"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: "app",
				LivenessProbe: &corev1.Probe{ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{Path: "/healthz"},
				}},
				ReadinessProbe: &corev1.Probe{ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{Path: "/ready"},
				}},
			}},
		},
	}
	paths := uniqueHTTPProbePaths(httpProbePathsFromPod(pod))
	require.ElementsMatch(t, []string{"/healthz", "/ready"}, paths)
	assert.True(t, podHasHTTPProbe(pod))
	assert.False(t, podHasHTTPProbe(&corev1.Pod{}))
}
