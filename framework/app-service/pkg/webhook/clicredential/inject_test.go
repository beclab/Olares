package clicredential

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/olarescli"

	corev1 "k8s.io/api/core/v1"
)

func podWith(containers ...corev1.Container) *corev1.Pod {
	return &corev1.Pod{Spec: corev1.PodSpec{Containers: containers}}
}

func mountPaths(c corev1.Container) []string {
	var paths []string
	for _, m := range c.VolumeMounts {
		paths = append(paths, m.MountPath)
	}
	return paths
}

func TestInjectMountsEveryContainer(t *testing.T) {
	pod := podWith(corev1.Container{Name: "app"}, corev1.Container{Name: "worker"})

	Inject(pod)

	if len(pod.Spec.Volumes) != 2 {
		t.Fatalf("expected credential and cache volumes, got %d: %v", len(pod.Spec.Volumes), pod.Spec.Volumes)
	}
	if got := pod.Spec.Volumes[0].Secret.SecretName; got != olarescli.CredentialSecretName {
		t.Errorf("credential volume points at %q, want %q", got, olarescli.CredentialSecretName)
	}
	for _, c := range pod.Spec.Containers {
		if len(c.VolumeMounts) != 2 {
			t.Fatalf("container %s: got mounts %v", c.Name, mountPaths(c))
		}
		if !c.VolumeMounts[0].ReadOnly {
			t.Errorf("container %s: the credential mount must be read-only", c.Name)
		}
		if c.VolumeMounts[1].ReadOnly {
			t.Errorf("container %s: the cache mount has to be writable, it is where refreshed tokens land", c.Name)
		}
		if len(c.Env) != 2 {
			t.Errorf("container %s: expected both directory env vars, got %v", c.Name, c.Env)
		}
	}
}

// Admission runs again on every pod update, and the platform must not stack up
// a second copy of the volumes each time.
func TestInjectIsIdempotent(t *testing.T) {
	pod := podWith(corev1.Container{Name: "app"})

	Inject(pod)
	Inject(pod)

	if len(pod.Spec.Volumes) != 2 {
		t.Errorf("volumes duplicated: %v", pod.Spec.Volumes)
	}
	if got := mountPaths(pod.Spec.Containers[0]); len(got) != 2 {
		t.Errorf("mounts duplicated: %v", got)
	}
	if got := pod.Spec.Containers[0].Env; len(got) != 2 {
		t.Errorf("env duplicated: %v", got)
	}
}

// A chart that already puts something at the credential path owns that path;
// replacing its mount would break the container rather than log it in.
func TestInjectLeavesConflictingMountAlone(t *testing.T) {
	pod := podWith(corev1.Container{
		Name: "app",
		VolumeMounts: []corev1.VolumeMount{
			{Name: "chart-own", MountPath: olarescli.CredentialMountPath},
		},
	})

	Inject(pod)

	if got := mountPaths(pod.Spec.Containers[0]); len(got) != 1 {
		t.Errorf("expected the chart's own mount to survive untouched, got %v", got)
	}
	if len(pod.Spec.Containers[0].Env) != 0 {
		t.Errorf("a container that was not mounted must not be told where the mounts are: %v",
			pod.Spec.Containers[0].Env)
	}
}

func TestInjectKeepsExistingEnvOverride(t *testing.T) {
	pod := podWith(corev1.Container{
		Name: "app",
		Env:  []corev1.EnvVar{{Name: EnvCredentialsDir, Value: "/somewhere/else"}},
	})

	Inject(pod)

	for _, e := range pod.Spec.Containers[0].Env {
		if e.Name == EnvCredentialsDir && e.Value != "/somewhere/else" {
			t.Errorf("%s was overwritten with %q", EnvCredentialsDir, e.Value)
		}
	}
}

// Init containers run before the app and have no session to use; keeping them
// out bounds how far the credential travels inside the pod.
func TestInjectSkipsInitContainers(t *testing.T) {
	pod := podWith(corev1.Container{Name: "app"})
	pod.Spec.InitContainers = []corev1.Container{{Name: "chown"}}

	Inject(pod)

	if len(pod.Spec.InitContainers[0].VolumeMounts) != 0 {
		t.Errorf("init container got mounts: %v", mountPaths(pod.Spec.InitContainers[0]))
	}
}
