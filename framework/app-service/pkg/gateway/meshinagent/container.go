package meshinagent

import (
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	meshInAgentImageEnv = "MESH_IN_AGENT_IMAGE"
	// DefaultImage is the R1 mesh-in-agent product image (engine: nginx+njs) (digest pin in charts; no :latest).
	DefaultImage = "beclab/mesh-in-agent:1.30.0-r1"
	listenPort   = 15443
	listenPortName = "mesh-in-https"

	CertsVolumeName  = "olares-mesh-in-certs"
	CertsMountPath   = "/var/run/olares/mesh-in-certs"
	HostsVolumeName  = "olares-mesh-in-shared-hosts"
	HostsMountPath   = "/var/run/olares/mesh-in-shared-hosts"
	ConfVolumeName   = "olares-mesh-in-agent-conf"
	ConfMountPath    = "/etc/nginx"
	InitContainerName = "olares-mesh-in-agent-iptables"
)

// ContainerSpec returns the mesh-in agent sidecar (WI-OC-CALLER-01 / IWO-OC-L1C-01).
func ContainerSpec() corev1.Container {
	return corev1.Container{
		Name:            ContainerName,
		Image:           meshInAgentImage(),
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
			{Name: "MESH_IN_AGENT_LISTEN_PORT", Value: "15443"},
			{Name: "MESH_IN_AGENT_GATEWAY_HOST", Value: "app-gateway-data.app-gateway.svc"},
			{Name: "MESH_IN_AGENT_GATEWAY_HTTP_PORT", Value: "80"},
		},
		// Do not mount ConfVolume over /etc/nginx until an init seeds it
		// (RenderNginxConf or image copy). An emptyDir here shadows the
		// product image's nginx.conf and crash-loops the sidecar.
		VolumeMounts: []corev1.VolumeMount{
			{Name: JWTSecretVolumeName, MountPath: JWTSecretMountPath, ReadOnly: true},
			{Name: CertsVolumeName, MountPath: CertsMountPath, ReadOnly: true},
			{Name: HostsVolumeName, MountPath: HostsMountPath, ReadOnly: true},
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

// InitContainerSpec redirects outbound TCP/443 into the mesh-in agent listen port.
func InitContainerSpec() corev1.Container {
	script := "iptables -t nat -A OUTPUT -p tcp --dport 443 -m owner ! --uid-owner 101 -j REDIRECT --to-ports 15443 || true"
	return corev1.Container{
		Name:            InitContainerName,
		Image:           meshInAgentImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/bin/sh", "-c", script},
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{Add: []corev1.Capability{"NET_ADMIN", "NET_RAW"}},
		},
	}
}

// JWTSecretVolume returns the projected JWT secret volume for the mesh-in agent.
func JWTSecretVolume() corev1.Volume {
	return corev1.Volume{
		Name: JWTSecretVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: "mesh-in-jwt",
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
				SecretName: "olares-mesh-in-certs",
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
				LocalObjectReference: corev1.LocalObjectReference{Name: "olares-mesh-in-shared-hosts"},
				Optional:             boolPtr(true),
			},
		},
	}
}

// ConfVolume is reserved for a follow-up that seeds RenderNginxConf (or copies
// image defaults) into an emptyDir before mounting at ConfMountPath. Callers
// must not mount it empty over /etc/nginx.
func ConfVolume() corev1.Volume {
	return corev1.Volume{
		Name: ConfVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

// ConfSeedInitContainerSpec copies the product image /etc/nginx into ConfVolume
// so a later mount at ConfMountPath does not start empty. Prefer this (or
// writing RenderNginxConf) before remounting ConfVolume on the sidecar.
func ConfSeedInitContainerSpec() corev1.Container {
	return corev1.Container{
		Name:            "olares-mesh-in-agent-conf-seed",
		Image:           meshInAgentImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command: []string{
			"/bin/sh", "-c",
			"cp -a /etc/nginx/. /conf/ && chmod -R a+rX /conf",
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: ConfVolumeName, MountPath: "/conf"},
		},
	}
}

func meshInAgentImage() string {
	if img := strings.TrimSpace(os.Getenv(meshInAgentImageEnv)); img != "" {
		return img
	}
	return DefaultImage
}

func boolPtr(v bool) *bool { return &v }

// IsStubImage reports whether image is the deprecated scaffold stub.
func IsStubImage(image string) bool {
	return strings.Contains(image, "0.0.0-stub")
}
