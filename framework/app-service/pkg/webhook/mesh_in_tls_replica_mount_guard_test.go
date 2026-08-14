package webhook

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
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

func TestIsLegitMeshInProtectedVolumeCustom(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: constants.MeshInAgentContainerName,
				VolumeMounts: []corev1.VolumeMount{
					{Name: constants.MeshInCustomCertsVolumeName, MountPath: "/custom"},
				},
			}},
		},
	}
	if !isLegitMeshInProtectedVolume(pod, constants.MeshInCustomCertsVolumeName, constants.MeshInCustomCertsVolumeName) {
		t.Fatal("mesh-in custom vol must allow")
	}
	if isLegitMeshInProtectedVolume(pod, constants.MeshInCertsVolumeName, constants.MeshInCustomCertsVolumeName) {
		t.Fatal("viewer vol name must not satisfy custom expected name")
	}
}

func TestValidateTLSReplicaMount_customSecretDenyAppMount(t *testing.T) {
	ns := "caller-alice"
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.MeshInCustomTLSSecretName,
			Namespace: ns,
			Labels:    map[string]string{constants.LabelTLSCustomReplica: "true"},
		},
	}
	wh := &Webhook{kubeClient: fake.NewSimpleClientset(sec)}
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{{
				Name: "steal",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{SecretName: constants.MeshInCustomTLSSecretName},
				},
			}},
			Containers: []corev1.Container{{
				Name: "app",
				VolumeMounts: []corev1.VolumeMount{
					{Name: "steal", MountPath: "/keys"},
				},
			}},
		},
	}
	ok, code := wh.ValidateTLSReplicaMount(context.Background(), pod, ns)
	if ok || code != codeTLSReplicaMountDenied {
		t.Fatalf("want deny %s, got ok=%v code=%s", codeTLSReplicaMountDenied, ok, code)
	}
}

func TestValidateTLSReplicaMount_customSecretAllowMeshIn(t *testing.T) {
	ns := "caller-alice"
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.MeshInCustomTLSSecretName,
			Namespace: ns,
			Labels:    map[string]string{constants.LabelTLSCustomReplica: "true"},
		},
	}
	wh := &Webhook{kubeClient: fake.NewSimpleClientset(sec)}
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{{
				Name: constants.MeshInCustomCertsVolumeName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{SecretName: constants.MeshInCustomTLSSecretName},
				},
			}},
			Containers: []corev1.Container{{
				Name: constants.MeshInAgentContainerName,
				VolumeMounts: []corev1.VolumeMount{
					{Name: constants.MeshInCustomCertsVolumeName, MountPath: "/custom"},
				},
			}},
		},
	}
	ok, code := wh.ValidateTLSReplicaMount(context.Background(), pod, ns)
	if !ok || code != "" {
		t.Fatalf("want allow, got ok=%v code=%s", ok, code)
	}
}

func TestValidateTLSReplicaMount_customSecretWrongVolumeName(t *testing.T) {
	ns := "caller-alice"
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.MeshInCustomTLSSecretName,
			Namespace: ns,
			Labels:    map[string]string{constants.LabelTLSCustomReplica: "true"},
		},
	}
	wh := &Webhook{kubeClient: fake.NewSimpleClientset(sec)}
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{{
				Name: constants.MeshInCertsVolumeName, // wrong: viewer vol name
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{SecretName: constants.MeshInCustomTLSSecretName},
				},
			}},
			Containers: []corev1.Container{{
				Name: constants.MeshInAgentContainerName,
				VolumeMounts: []corev1.VolumeMount{
					{Name: constants.MeshInCertsVolumeName, MountPath: "/custom"},
				},
			}},
		},
	}
	ok, code := wh.ValidateTLSReplicaMount(context.Background(), pod, ns)
	if ok || code != codeTLSReplicaMountDenied {
		t.Fatalf("want deny for wrong volume name, got ok=%v code=%s", ok, code)
	}
}

func TestValidateTLSReplicaMount_missingCustomSecretDenyAppMount(t *testing.T) {
	ns := "caller-alice"
	wh := &Webhook{kubeClient: fake.NewSimpleClientset()}
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{{
				Name: "steal",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: constants.MeshInCustomTLSSecretName,
						Optional:   boolPtr(true),
					},
				},
			}},
			Containers: []corev1.Container{{
				Name: "app",
				VolumeMounts: []corev1.VolumeMount{
					{Name: "steal", MountPath: "/keys"},
				},
			}},
		},
	}
	ok, code := wh.ValidateTLSReplicaMount(context.Background(), pod, ns)
	if ok || code != codeTLSReplicaMountDenied {
		t.Fatalf("missing well-known custom secret must still deny app mount, got ok=%v code=%s", ok, code)
	}
}

func TestValidateTLSReplicaMount_missingCustomSecretAllowMeshIn(t *testing.T) {
	ns := "caller-alice"
	wh := &Webhook{kubeClient: fake.NewSimpleClientset()}
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{{
				Name: constants.MeshInCustomCertsVolumeName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: constants.MeshInCustomTLSSecretName,
						Optional:   boolPtr(true),
					},
				},
			}},
			Containers: []corev1.Container{{
				Name: constants.MeshInAgentContainerName,
				VolumeMounts: []corev1.VolumeMount{
					{Name: constants.MeshInCustomCertsVolumeName, MountPath: "/custom"},
				},
			}},
		},
	}
	ok, code := wh.ValidateTLSReplicaMount(context.Background(), pod, ns)
	if !ok || code != "" {
		t.Fatalf("mesh-in optional mount of missing custom secret must allow, got ok=%v code=%s", ok, code)
	}
}

func boolPtr(v bool) *bool { return &v }
