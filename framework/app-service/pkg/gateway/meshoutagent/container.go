package meshoutagent

import (
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	meshOutAgentImageEnv = "MESH_OUT_AGENT_IMAGE"
	// DefaultImage is the R2 mesh-out-agent product image (engine: nginx alpine-slim) (digest pin in charts).
	DefaultImage       = "beclab/mesh-out-agent:1.30.0-r2"
	systemServerHost   = "system-server.user-system"
	systemServerPort   = "28080"
	InitContainerName  = "olares-mesh-out-agent-iptables"
	ConfVolumeName     = "olares-mesh-out-agent-conf"
	ConfMountPath      = "/etc/nginx"
)

// ContainerSpec returns the mesh-out agent sidecar (WI-OC-EGRESS-01 / IWO-OC-L2B-01).
func ContainerSpec() corev1.Container {
	return corev1.Container{
		Name:            ContainerName,
		Image:           meshOutAgentImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"nginx", "-g", "daemon off;"},
		Ports: []corev1.ContainerPort{
			{
				Name:          ListenPortName,
				ContainerPort: ListenPort,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{Name: FailClosedEnv, Value: "true"},
			{Name: "MESH_OUT_SYSTEM_SERVER_HOST", Value: systemServerHost},
			{Name: "MESH_OUT_SYSTEM_SERVER_PORT", Value: systemServerPort},
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: SATokenVolumeName, MountPath: SATokenMountPath, ReadOnly: true},
			{Name: ConfVolumeName, MountPath: ConfMountPath, ReadOnly: true},
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

// InitContainerSpec redirects outbound TCP/80 and TCP/8080 into the mesh-out agent.
func InitContainerSpec() corev1.Container {
	script := "iptables -t nat -A OUTPUT -p tcp --dport 80 -m owner ! --uid-owner 101 -j REDIRECT --to-ports 15001 || true; " +
		"iptables -t nat -A OUTPUT -p tcp --dport 8080 -m owner ! --uid-owner 101 -j REDIRECT --to-ports 15001 || true"
	return corev1.Container{
		Name:            InitContainerName,
		Image:           meshOutAgentImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/bin/sh", "-c", script},
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{Add: []corev1.Capability{"NET_ADMIN", "NET_RAW"}},
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

// ConfVolume holds RenderMeshOutNginxConf output.
func ConfVolume() corev1.Volume {
	return corev1.Volume{
		Name: ConfVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

func meshOutAgentImage() string {
	if img := strings.TrimSpace(os.Getenv(meshOutAgentImageEnv)); img != "" {
		return img
	}
	return DefaultImage
}

// IsStubImage reports whether image is the deprecated scaffold stub.
func IsStubImage(image string) bool {
	return strings.Contains(image, "0.0.0-stub")
}
