package egressagent

import (
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	egressAgentImageEnv = "EGRESS_AGENT_IMAGE"
	defaultStubImage    = "beclab/egress-agent:0.0.0-stub"
	systemServerHost    = "system-server.user-system"
	systemServerPort    = "28080"
)

// ContainerSpec returns the minimal egress agent scaffold (WI-OC-EGRESS-01).
func ContainerSpec() corev1.Container {
	return corev1.Container{
		Name:            ContainerName,
		Image:           egressAgentImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports: []corev1.ContainerPort{
			{
				Name:          ListenPortName,
				ContainerPort: ListenPort,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{Name: FailClosedEnv, Value: "true"},
			{Name: "EGRESS_SYSTEM_SERVER_HOST", Value: systemServerHost},
			{Name: "EGRESS_SYSTEM_SERVER_PORT", Value: systemServerPort},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      SATokenVolumeName,
				MountPath: SATokenMountPath,
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

// SATokenVolume projects the pod service account token for Bearer injection.
func SATokenVolume() corev1.Volume {
	expiration := int64(3600)
	return corev1.Volume{
		Name: SATokenVolumeName,
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{
					{
						ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
							Audience:          "system-server",
							ExpirationSeconds: &expiration,
							Path:              "token",
						},
					},
				},
			},
		},
	}
}

func egressAgentImage() string {
	if img := strings.TrimSpace(os.Getenv(egressAgentImageEnv)); img != "" {
		return img
	}
	return defaultStubImage
}
