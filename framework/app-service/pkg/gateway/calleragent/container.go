package calleragent

import (
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	callerAgentImageEnv = "CALLER_AGENT_IMAGE"
	// DefaultImage is the R1 nginx+njs agent image (digest pin in charts; no :latest).
	DefaultImage = "beclab/nginx:1.30.0-alpine-njs-olares-r1"
	listenPort   = 15443
	listenPortName = "caller-https"

	CertsVolumeName  = "olares-caller-certs"
	CertsMountPath   = "/var/run/olares/caller-certs"
	HostsVolumeName  = "olares-caller-shared-hosts"
	HostsMountPath   = "/var/run/olares/caller-shared-hosts"
	ConfVolumeName   = "olares-caller-agent-conf"
	ConfMountPath    = "/etc/nginx"
	InitContainerName = "olares-caller-agent-iptables"
)

// ContainerSpec returns the caller agent sidecar (WI-OC-CALLER-01 / IWO-OC-L1C-01).
func ContainerSpec() corev1.Container {
	return corev1.Container{
		Name:            ContainerName,
		Image:           callerAgentImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"nginx", "-g", "daemon off;"},
		Ports: []corev1.ContainerPort{
			{
				Name:          listenPortName,
				ContainerPort: listenPort,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{Name: FailClosedEnv, Value: "true"},
			{Name: "CALLER_AGENT_LISTEN_PORT", Value: "15443"},
			{Name: "CALLER_AGENT_GATEWAY_HOST", Value: "app-gateway-data.app-gateway.svc"},
			{Name: "CALLER_AGENT_GATEWAY_HTTP_PORT", Value: "80"},
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: JWTSecretVolumeName, MountPath: JWTSecretMountPath, ReadOnly: true},
			{Name: CertsVolumeName, MountPath: CertsMountPath, ReadOnly: true},
			{Name: HostsVolumeName, MountPath: HostsMountPath, ReadOnly: true},
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

// InitContainerSpec redirects outbound TCP/443 into the caller agent listen port.
func InitContainerSpec() corev1.Container {
	script := "iptables -t nat -A OUTPUT -p tcp --dport 443 -m owner ! --uid-owner 101 -j REDIRECT --to-ports 15443 || true"
	return corev1.Container{
		Name:            InitContainerName,
		Image:           callerAgentImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/bin/sh", "-c", script},
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{Add: []corev1.Capability{"NET_ADMIN", "NET_RAW"}},
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

// CertsVolume mounts WI-N4 caller cert replicas (optional until N4 Ready).
func CertsVolume() corev1.Volume {
	return corev1.Volume{
		Name: CertsVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: "olares-caller-certs",
				Optional:   boolPtr(true),
			},
		},
	}
}

// SharedHostsVolume mounts WI-N6 shared-hosts allowlist ConfigMap.
func SharedHostsVolume() corev1.Volume {
	return corev1.Volume{
		Name: HostsVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: "olares-caller-shared-hosts"},
				Optional:             boolPtr(true),
			},
		},
	}
}

// ConfVolume carries RenderNginxConf output as an emptyDir populated by an
// upstream controller or init; for scaffold the emptyDir holds the rendered file
// via annotation-driven admission in a follow-up.
func ConfVolume() corev1.Volume {
	return corev1.Volume{
		Name: ConfVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

func callerAgentImage() string {
	if img := strings.TrimSpace(os.Getenv(callerAgentImageEnv)); img != "" {
		return img
	}
	return DefaultImage
}

func boolPtr(v bool) *bool { return &v }

// IsStubImage reports whether image is the deprecated scaffold stub.
func IsStubImage(image string) bool {
	return strings.Contains(image, "0.0.0-stub")
}
