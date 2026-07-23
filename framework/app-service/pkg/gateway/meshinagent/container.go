package meshinagent

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway/callerjwt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	meshInAgentImageEnv = "MESH_IN_AGENT_IMAGE"
	// DefaultImage is the mesh-in-agent product image (engine: nginx+njs) (digest pin in charts; no :latest).
	DefaultImage = "beclab/mesh-in-agent:1.30.0-r3"
	listenPort     = HTTPListenPort
	listenPortName = "mesh-in-http"

	CertsVolumeName   = "olares-mesh-in-certs"
	CertsMountPath    = "/var/run/olares/mesh-in-certs"
	HostsVolumeName   = "olares-mesh-in-shared-hosts"
	HostsMountPath    = "/var/run/olares/mesh-in-shared-hosts"
	ConfVolumeName    = "olares-mesh-in-agent-conf"
	ConfMountPath     = "/etc/nginx"
	InitContainerName = "olares-mesh-in-agent-iptables"

	// NginxWorkerUID is the alpine nginx image worker uid; OUTPUT redirects skip this owner to avoid loops.
	NginxWorkerUID = "101"
)

// ContainerSpec returns the mesh-in agent sidecar (WI-OC-JWT-INJECT-01).
func ContainerSpec() corev1.Container {
	return corev1.Container{
		Name:            ContainerName,
		Image:           meshInAgentImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/bin/sh", "-c", agentStartScript()},
		Ports: []corev1.ContainerPort{
			{
				Name:          listenPortName,
				ContainerPort: listenPort,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{Name: FailClosedEnv, Value: "true"},
			{Name: "MESH_IN_AGENT_LISTEN_PORT", Value: fmt.Sprintf("%d", HTTPListenPort)},
			{Name: "MESH_IN_AGENT_GATEWAY_HOST", Value: DefaultGatewayHost},
			{Name: "MESH_IN_AGENT_GATEWAY_HTTP_PORT", Value: "80"},
		},
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

func agentStartScript() string {
	conf := RenderNginxConf(NginxConfInput{FailClosed: true})
	js := BearerJS()
	confB64 := base64.StdEncoding.EncodeToString([]byte(conf))
	jsB64 := base64.StdEncoding.EncodeToString([]byte(js))
	return fmt.Sprintf(`set -eu
mkdir -p /tmp/mesh-in /var/log/nginx
echo '%s' | base64 -d > /tmp/mesh-in/nginx.conf
echo '%s' | base64 -d > /tmp/mesh-in/bearer.js
exec nginx -c /tmp/mesh-in/nginx.conf -g 'daemon off;'
`, confB64, jsB64)
}

// InitContainerSpec redirects outbound TCP/80 toward the shared gateway into the mesh-in HTTP listener.
// Rules are inserted at the head of OUTPUT so they take precedence over olares-envoy PROXY_OUTBOUND.
func InitContainerSpec() corev1.Container {
	script := fmt.Sprintf(`set -eu
GW_HOST="${MESH_IN_AGENT_GATEWAY_HOST:-%s}"
GW_IP=""
# Prefer configured host; fall back to legacy app-gateway NS for older installs.
for h in "$GW_HOST" "app-gateway-data.os-gateway.svc" "app-gateway-data.os-gateway.svc.cluster.local" \
  "app-gateway-data.app-gateway.svc" "app-gateway-data.app-gateway.svc.cluster.local"; do
  if command -v getent >/dev/null 2>&1; then
    GW_IP=$(getent ahosts "$h" 2>/dev/null | awk '{print $1; exit}')
  fi
  if [ -z "$GW_IP" ] && command -v nslookup >/dev/null 2>&1; then
    GW_IP=$(nslookup "$h" 2>/dev/null | awk '/^Address: / { a=$2 } END { print a }')
  fi
  if [ -n "$GW_IP" ]; then
    echo "mesh-in-agent: resolved $h -> $GW_IP"
    break
  fi
done
if [ -z "$GW_IP" ]; then
  echo "mesh-in-agent: cannot resolve gateway host $GW_HOST" >&2
  exit 1
fi
echo "mesh-in-agent: redirect $GW_IP:80 -> %d (skip uid %s)"
iptables -t nat -C OUTPUT -p tcp -d "$GW_IP" --dport 80 -m owner ! --uid-owner %s -j REDIRECT --to-ports %d 2>/dev/null \
  || iptables -t nat -I OUTPUT 1 -p tcp -d "$GW_IP" --dport 80 -m owner ! --uid-owner %s -j REDIRECT --to-ports %d
`, DefaultGatewayHost, HTTPListenPort, NginxWorkerUID, NginxWorkerUID, HTTPListenPort, NginxWorkerUID, HTTPListenPort)

	return corev1.Container{
		Name:            InitContainerName,
		Image:           meshInAgentImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/bin/sh", "-c", script},
		Env: []corev1.EnvVar{
			{Name: "MESH_IN_AGENT_GATEWAY_HOST", Value: DefaultGatewayHost},
		},
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{Add: []corev1.Capability{"NET_ADMIN", "NET_RAW"}},
		},
	}
}

// JWTSecretVolume mounts the caller JWT-SVID Secret issued by callerjwt (name must match AppJWTSecretName).
func JWTSecretVolume() corev1.Volume {
	return corev1.Volume{
		Name: JWTSecretVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: callerjwt.AppJWTSecretName,
				Optional:   boolPtr(false),
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
