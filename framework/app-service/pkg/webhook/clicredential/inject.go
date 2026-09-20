// Package clicredential mounts the Olares credential that app-service
// provisions for apps declaring permission.loginOlaresCLI, so that olares-cli
// running inside the container starts out logged in as the app's owner.
//
// The contract with the container is two directories and the environment
// variables that point at them:
//
//	/olares/credentials  read-only: credential.json with refreshToken,
//	                     olaresId, and appName. Backed by the
//	                     olares-cli-credential Secret that pkg/olarescli
//	                     writes into the namespace.
//	/olares/cache        writable scratch for whatever the CLI derives from
//	                     the credential (access tokens, mostly). An emptyDir,
//	                     so it is per-pod and disappears with it.
//
// Only the refresh token is handed over. Access tokens last a day, so one
// baked into a mounted file would be stale far more often than not; the
// consumer exchanges the refresh token when it needs one.
package clicredential

import (
	"github.com/beclab/Olares/framework/app-service/pkg/olarescli"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

const (
	credentialVolumeName = "olares-cli-credential"
	cacheVolumeName      = "olares-cli-cache"

	// EnvCredentialsDir and EnvCacheDir let the consumer find the mounts
	// without compiling the paths in.
	EnvCredentialsDir = "OLARES_CLI_CREDENTIALS_DIR"
	EnvCacheDir       = "OLARES_CLI_CACHE_DIR"

	// credentialFileMode is world-readable on purpose. The files are owned by
	// root, apps run as uid 1000, and every process in the pod could read the
	// Secret through its own ServiceAccount anyway; a stricter mode would
	// only lock out the one process the credential exists for.
	credentialFileMode int32 = 0444
)

// Inject adds the credential mounts to every container in the pod. It is
// idempotent: a pod that already carries the volumes (a re-admitted pod, or
// one whose chart declared them) is left alone.
//
// Init containers are skipped. They run before the app is up and have no use
// for a session, and keeping them out limits how far the credential travels.
func Inject(pod *corev1.Pod) {
	if pod == nil {
		return
	}
	addVolumes(pod)
	for i := range pod.Spec.Containers {
		addMounts(&pod.Spec.Containers[i])
	}
}

func addVolumes(pod *corev1.Pod) {
	existing := make(map[string]struct{}, len(pod.Spec.Volumes))
	for _, v := range pod.Spec.Volumes {
		existing[v.Name] = struct{}{}
	}
	if _, ok := existing[credentialVolumeName]; !ok {
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
			Name: credentialVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  olarescli.CredentialSecretName,
					DefaultMode: ptr.To(credentialFileMode),
				},
			},
		})
	}
	if _, ok := existing[cacheVolumeName]; !ok {
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
			Name:         cacheVolumeName,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		})
	}
}

func addMounts(container *corev1.Container) {
	for _, m := range container.VolumeMounts {
		// A chart that already mounts something at these paths knows what it
		// is doing; overwriting it would break the container instead of
		// helping it.
		if m.MountPath == olarescli.CredentialMountPath || m.MountPath == olarescli.CacheMountPath {
			return
		}
	}
	container.VolumeMounts = append(container.VolumeMounts,
		corev1.VolumeMount{
			Name:      credentialVolumeName,
			MountPath: olarescli.CredentialMountPath,
			ReadOnly:  true,
		},
		corev1.VolumeMount{
			Name:      cacheVolumeName,
			MountPath: olarescli.CacheMountPath,
		},
	)
	setEnv(container, EnvCredentialsDir, olarescli.CredentialMountPath)
	setEnv(container, EnvCacheDir, olarescli.CacheMountPath)
}

// setEnv leaves an existing value in place: an app that points the CLI
// somewhere else has overridden the platform on purpose.
func setEnv(container *corev1.Container, name, value string) {
	for _, e := range container.Env {
		if e.Name == name {
			return
		}
	}
	container.Env = append(container.Env, corev1.EnvVar{Name: name, Value: value})
}
