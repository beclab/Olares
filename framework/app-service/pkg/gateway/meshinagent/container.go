package meshinagent

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/callerjwt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	meshInAgentImageEnv = "MESH_IN_AGENT_IMAGE"
	// DefaultImage is the mesh-in-agent product image (engine: nginx+njs+stream) (digest pin in charts; no :latest).
	DefaultImage       = "beclab/mesh-in-agent:1.30.0-r4"
	listenPortName     = "mesh-in-http"
	httpsListenPortName = "mesh-in-https"

	CertsVolumeName   = "olares-mesh-in-certs"
	CertsMountPath    = "/var/run/olares/mesh-in-certs"
	HostsVolumeName   = "olares-mesh-in-shared-hosts"
	HostsMountPath    = "/var/run/olares/mesh-in-shared-hosts"
	ConfVolumeName    = "olares-mesh-in-agent-conf"
	ConfMountPath     = "/etc/nginx"
	InitContainerName = "olares-mesh-in-agent-iptables"
)

// NginxWorkerUID is the mesh-in process uid (constants.MeshInAgentUID).
func NginxWorkerUID() string {
	return strconv.FormatInt(constants.MeshInAgentUID, 10)
}

// ContainerSpec returns the mesh-in agent sidecar (JWT inject + CT-1 TLS offload).
func ContainerSpec() corev1.Container {
	uid := constants.MeshInAgentUID
	nonRoot := true
	noEsc := false
	return corev1.Container{
		Name:            ContainerName,
		Image:           meshInAgentImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/bin/sh", "-c", agentStartScript()},
		Ports: []corev1.ContainerPort{
			{
				Name:          listenPortName,
				ContainerPort: HTTPListenPort,
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          httpsListenPortName,
				ContainerPort: HTTPSListenPort,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{Name: FailClosedEnv, Value: "true"},
			{Name: "MESH_IN_AGENT_LISTEN_PORT", Value: fmt.Sprintf("%d", HTTPListenPort)},
			{Name: "MESH_IN_AGENT_HTTPS_LISTEN_PORT", Value: fmt.Sprintf("%d", HTTPSListenPort)},
			{Name: "MESH_IN_AGENT_GATEWAY_HOST", Value: DefaultGatewayHost},
			{Name: "MESH_IN_AGENT_GATEWAY_HTTP_PORT", Value: "80"},
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: JWTSecretVolumeName, MountPath: JWTSecretMountPath, ReadOnly: true},
			{Name: CertsVolumeName, MountPath: CertsMountPath, ReadOnly: true},
			{Name: HostsVolumeName, MountPath: HostsMountPath, ReadOnly: true},
		},
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:                &uid,
			RunAsGroup:               &uid,
			RunAsNonRoot:             &nonRoot,
			AllowPrivilegeEscalation: &noEsc,
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
	hostsFile := HostsMountPath + "/" + SharedHostsFileName
	return fmt.Sprintf(`set -eu
mkdir -p /tmp/mesh-in /var/log/nginx
echo '%s' | base64 -d > /tmp/mesh-in/nginx.conf
echo '%s' | base64 -d > /tmp/mesh-in/bearer.js
# SNI allowlist map (N6 shared-hosts.txt); empty → all HTTPS SNI rejected.
: > /tmp/mesh-in/sni_map.conf
HOSTS_FILE="%s"
if [ -f "$HOSTS_FILE" ]; then
  while IFS= read -r line || [ -n "$line" ]; do
    line=$(printf '%%s' "$line" | sed 's/#.*//;s/^[[:space:]]*//;s/[[:space:]]*$//' | tr 'A-Z' 'a-z')
    [ -z "$line" ] && continue
    printf '    %%s 127.0.0.1:%d;\n' "$line" >> /tmp/mesh-in/sni_map.conf
  done < "$HOSTS_FILE"
fi
# TLS material: prefer projected Secret; else ephemeral so nginx can start (HTTPS fail-closed for real clients).
CERT_DIR="%s"
if [ ! -s "$CERT_DIR/tls.crt" ] || [ ! -s "$CERT_DIR/tls.key" ]; then
  mkdir -p /tmp/mesh-in/certs
  if command -v openssl >/dev/null 2>&1; then
    openssl req -x509 -newkey rsa:2048 -nodes -days 1 \
      -subj '/CN=mesh-in-ephemeral' \
      -keyout /tmp/mesh-in/certs/tls.key \
      -out /tmp/mesh-in/certs/tls.crt >/dev/null 2>&1
    sed -i 's|%s/tls.crt|/tmp/mesh-in/certs/tls.crt|;s|%s/tls.key|/tmp/mesh-in/certs/tls.key|' /tmp/mesh-in/nginx.conf
  fi
fi
exec nginx -c /tmp/mesh-in/nginx.conf -g 'daemon off;'
`, confB64, jsB64, hostsFile, HTTPSTerminatePort, CertsMountPath, CertsMountPath, CertsMountPath)
}

// InitContainerSpec installs OUTPUT REDIRECT rules for Shared gateway HTTP and
// all outbound HTTPS. mesh-in upstream to gateway:80 must NOT use a blanket
// uid RETURN so Linkerd (uid 2102) can intercept plaintext after JWT inject.
//
// Exclusions on REDIRECT owners: mesh-in 1651 (no self-loop), linkerd 2102,
// envoy 1555 (legacy oes coexistence).
func InitContainerSpec() corev1.Container {
	envoyUID := strconv.FormatInt(constants.EnvoyUID, 10)
	linkerdUID := strconv.FormatInt(constants.LinkerdProxyUID, 10)
	root := int64(0)
	script := fmt.Sprintf(`set -eu
GW_HOST="${MESH_IN_AGENT_GATEWAY_HOST:-%s}"
GW_IP=""
NGINX_UID="%s"
ENVOY_UID="%s"
LINKERD_UID="%s"
HTTP_PORT="%d"
HTTPS_PORT="%d"
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
OWNER_SKIP="-m owner ! --uid-owner $NGINX_UID -m owner ! --uid-owner $ENVOY_UID -m owner ! --uid-owner $LINKERD_UID"
echo "mesh-in-agent: redirect $GW_IP:80 -> $HTTP_PORT; *:443 -> $HTTPS_PORT (skip uid $NGINX_UID/$ENVOY_UID/$LINKERD_UID)"
iptables -t nat -C OUTPUT -p tcp -d "$GW_IP" --dport 80 $OWNER_SKIP -j REDIRECT --to-ports "$HTTP_PORT" 2>/dev/null \
  || iptables -t nat -I OUTPUT 1 -p tcp -d "$GW_IP" --dport 80 $OWNER_SKIP -j REDIRECT --to-ports "$HTTP_PORT"
iptables -t nat -C OUTPUT -p tcp --dport 443 $OWNER_SKIP -j REDIRECT --to-ports "$HTTPS_PORT" 2>/dev/null \
  || iptables -t nat -I OUTPUT 2 -p tcp --dport 443 $OWNER_SKIP -j REDIRECT --to-ports "$HTTPS_PORT"
`, DefaultGatewayHost, NginxWorkerUID(), envoyUID, linkerdUID, HTTPListenPort, HTTPSListenPort)

	return corev1.Container{
		Name:            InitContainerName,
		Image:           meshInAgentImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/bin/sh", "-c", script},
		Env: []corev1.EnvVar{
			{Name: "MESH_IN_AGENT_GATEWAY_HOST", Value: DefaultGatewayHost},
		},
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &root,
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

// CertsVolume mounts caller TLS material for CT-1 (optional until replica Ready).
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

// SharedHostsVolume mounts SNI allowlist ConfigMap (key shared-hosts.txt).
func SharedHostsVolume() corev1.Volume {
	return corev1.Volume{
		Name: HostsVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: "olares-mesh-in-shared-hosts"},
				Optional:             boolPtr(true),
				Items: []corev1.KeyToPath{
					{Key: SharedHostsFileName, Path: SharedHostsFileName},
				},
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
