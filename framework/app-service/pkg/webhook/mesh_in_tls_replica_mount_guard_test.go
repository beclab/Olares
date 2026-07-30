package webhook

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsLegitMeshInReplicaVolume(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "p"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: constants.MeshInAgentContainerName,
					VolumeMounts: []corev1.VolumeMount{
						{Name: constants.MeshInCertsVolumeName, MountPath: "/certs"},
					},
				},
			},
		},
	}
	if !isLegitMeshInReplicaVolume(pod, constants.MeshInCertsVolumeName) {
		t.Fatal("expected allow for mesh-in-agent")
	}
	pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{
		Name: "app",
		VolumeMounts: []corev1.VolumeMount{
			{Name: constants.MeshInCertsVolumeName, MountPath: "/steal"},
		},
	})
	if isLegitMeshInReplicaVolume(pod, constants.MeshInCertsVolumeName) {
		t.Fatal("expected deny when app also mounts replica volume")
	}
	if isLegitMeshInReplicaVolume(pod, "other-vol") {
		t.Fatal("wrong volume name must deny")
	}
}
