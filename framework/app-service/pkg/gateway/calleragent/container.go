package calleragent

import (
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	callerAgentImageEnv = "CALLER_AGENT_IMAGE"
	defaultStubImage    = "beclab/caller-agent:0.0.0-stub"
	listenPort          = 15044
	listenPortName      = "caller-http"
)

// ContainerSpec returns the minimal caller agent sidecar scaffold (WI-OC-L1C).
// The stub terminates HTTPS locally and injects Authorization: Bearer <JWT-SVID>
// before forwarding to app-gateway-data.
func ContainerSpec() corev1.Container {
	return corev1.Container{
		Name:            ContainerName,
		Image:           callerAgentImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports: []corev1.ContainerPort{
			{
				Name:          listenPortName,
				ContainerPort: listenPort,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{Name: FailClosedEnv, Value: "true"},
			{Name: "CALLER_AGENT_LISTEN_PORT", Value: "15044"},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      JWTSecretVolumeName,
				MountPath: JWTSecretMountPath,
				ReadOnly:  true,
			},
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("32Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("64Mi"),
			},
		},
	}
}

// JWTSecretVolume returns the projected JWT secret volume for the caller agent.
func JWTSecretVolume() corev1.Volume {
	return corev1.Volume{
		Name: JWTSecretVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: "caller-jwt",
				Optional:   boolPtr(true),
			},
		},
	}
}

func callerAgentImage() string {
	if img := strings.TrimSpace(os.Getenv(callerAgentImageEnv)); img != "" {
		return img
	}
	return defaultStubImage
}

func boolPtr(v bool) *bool {
	return &v
}
