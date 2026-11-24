package meshinagent

import (
	"encoding/base64"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/callerjwt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	meshInAgentImageEnv = "MESH_IN_AGENT_IMAGE"
	// DefaultImage is the mesh-in-agent product image (engine: nginx+njs+stream) (versioned tag; release overlays pin the digest).
	DefaultImage        = "beclab/mesh-in-agent:1.30.8-r1"
	listenPortName      = "mesh-in-http"
	httpsListenPortName = "mesh-in-https"

	CertsVolumeName   = constants.MeshInCertsVolumeName
	CertsMountPath    = "/var/run/olares/mesh-in-certs"
	HostsVolumeName   = constants.MeshInSharedHostsCMName
	HostsMountPath    = "/var/run/olares/mesh-in-shared-hosts"
	ConfVolumeName    = "olares-mesh-in-agent-conf"
	ConfMountPath     = "/etc/nginx"
	InitContainerName = "olares-mesh-in-agent-iptables"
)

// NginxWorkerUID is the mesh-in process uid (constants.MeshInAgentUID).
func NginxWorkerUID() string {
	return strconv.FormatInt(constants.MeshInAgentUID, 10)
}

// ContainerSpec returns the mesh-in agent sidecar (JWT inject + TLS offload).
func ContainerSpec() corev1.Container {
	uid := constants.MeshInAgentUID
	nonRoot := true
	noEsc := false
	always := corev1.ContainerRestartPolicyAlways
	return corev1.Container{
		Name:            ContainerName,
		Image:           meshInAgentImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/bin/sh", "-c", agentStartScript()},
		RestartPolicy:   &always,
		StartupProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{
				"/bin/sh", "-ec", fmt.Sprintf("nc -z -w 1 127.0.0.1 %d && nc -z -w 1 127.0.0.1 %d",
					HTTPListenPort, HTTPSListenPort),
			}}},
			TimeoutSeconds: 3, PeriodSeconds: 2, FailureThreshold: 60,
		},
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
	// Always enable HTTPS. Generate a pod-local placeholder only when real TLS
	// material is absent at startup; njs hot-switches new connections to real
	// files when they appear.
	confHTTPS := RenderNginxConf(NginxConfInput{FailClosed: true, EnableHTTPS: true})
	js := BearerJS()
	httpsB64 := base64.StdEncoding.EncodeToString([]byte(confHTTPS))
	jsB64 := base64.StdEncoding.EncodeToString([]byte(js))
	return fmt.Sprintf(`set -eu
mkdir -p /tmp/mesh-in /var/log/nginx %s
echo '%s' | base64 -d > /tmp/mesh-in/bearer.js
echo '%s' | base64 -d > /tmp/mesh-in/nginx.conf
CERT_DIR="%s"
if [ -s "$CERT_DIR/tls.crt" ] && [ -s "$CERT_DIR/tls.key" ]; then
  echo "mesh-in-agent: HTTPS listeners up; TLS mode=real (certs present under $CERT_DIR)"
else
  if command -v ensure-placeholder-cert >/dev/null 2>&1; then
    ensure-placeholder-cert -dir %s
  else
    echo "mesh-in-agent: ensure-placeholder-cert missing; HTTPS cannot start without real certs" >&2
    exit 1
  fi
  echo "mesh-in-agent: HTTPS listeners up; TLS mode=placeholder"
fi
exec nginx -c /tmp/mesh-in/nginx.conf -g 'daemon off;'
`, PlaceholderCertDir, jsB64, httpsB64, CertsMountPath, PlaceholderCertDir)
}

// InitContainerSpec installs OUTPUT REDIRECT rules for Shared gateway HTTP/HTTPS.
// Every REDIRECT must pin -d to the gateway data-plane IP (same constraint for
// 80 and 443): mesh-in always forwards to the gateway and cannot recover the
// original destination under REDIRECT when SNI is absent.
//
// Exclusions on REDIRECT owners: mesh-in 1651 (no self-loop), linkerd 2102,
// envoy 1555 (legacy oes coexistence).
//
// Resolve failure must not block the workload (exit 0 after retries).
func InitContainerSpec() corev1.Container {
	return InitContainerSpecWithGatewayIPs("")
}

// InitContainerSpecWithGatewayIPs is InitContainerSpec with an optional
// comma/space-separated ClusterIP list (admission snapshot). Empty falls back
// to DNS inside the init script.
func InitContainerSpecWithGatewayIPs(gatewayIPs string) corev1.Container {
	root := int64(0)
	env := []corev1.EnvVar{
		{Name: "MESH_IN_AGENT_GATEWAY_HOST", Value: DefaultGatewayHost},
	}
	if ips := strings.TrimSpace(gatewayIPs); ips != "" {
		env = append(env, corev1.EnvVar{Name: GatewayIPsEnv, Value: ips})
	}

	return corev1.Container{
		Name:            InitContainerName,
		Image:           meshInAgentImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/bin/sh", "-c", IptablesInitScript()},
		Env:             env,
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:    &root,
			Capabilities: &corev1.Capabilities{Add: []corev1.Capability{"NET_ADMIN", "NET_RAW"}},
		},
	}
}

// IptablesInitScript renders the mesh-in iptables bootstrap shell (unit-tested).
func IptablesInitScript() string {
	envoyUID := strconv.FormatInt(constants.EnvoyUID, 10)
	linkerdUID := strconv.FormatInt(constants.LinkerdProxyUID, 10)
	return fmt.Sprintf(`set -eu
GW_HOST="${MESH_IN_AGENT_GATEWAY_HOST:-%s}"
NGINX_UID="%s"
ENVOY_UID="%s"
LINKERD_UID="%s"
HTTP_PORT="%d"
HTTPS_PORT="%d"

normalize_ips() {
  printf '%%s\n' "$1" | tr ',;[:space:]' '\n' | awk '
    /^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$/ {
      split($0, octet, ".");
      if (octet[1] <= 255 && octet[2] <= 255 && octet[3] <= 255 && octet[4] <= 255 && !seen[$0]++) print;
      next;
    }
    /^[0-9a-fA-F:.]+$/ && /:/ && !seen[$0]++ {print}
  '
}

resolve_gateway_ips() {
  for host in "$GW_HOST" app-gateway-data.os-gateway.svc.cluster.local; do
    if command -v dig >/dev/null 2>&1; then
      dig +short +time=2 +tries=1 "$host" A 2>/dev/null || true
      dig +short +time=2 +tries=1 "$host" AAAA 2>/dev/null || true
    else
      getent ahosts "$host" 2>/dev/null | awk '{print $1}'
    fi
  done
}

redirect_port() {
  tool="$1"; ip="$2"; port="$3"; listen="$4"
  if ! "$tool" -w 5 -t nat -C OUTPUT -p tcp -d "$ip" --dport "$port" $OWNER_SKIP -j REDIRECT --to-ports "$listen" 2>/dev/null; then
    "$tool" -w 5 -t nat -I OUTPUT 1 -p tcp -d "$ip" --dport "$port" $OWNER_SKIP -j REDIRECT --to-ports "$listen" || return 1
    "$tool" -w 5 -t nat -C OUTPUT -p tcp -d "$ip" --dport "$port" $OWNER_SKIP -j REDIRECT --to-ports "$listen" || return 1
  fi
}

install_redirects() {
  for ip in $1; do
    case "$ip" in
      *:*) tool=ip6tables ;;
      *) tool=iptables ;;
    esac
    redirect_port "$tool" "$ip" 80 "$HTTP_PORT" || return 1
    redirect_port "$tool" "$ip" 443 "$HTTPS_PORT" || return 1
  done
}

OWNER_SKIP="-m owner ! --uid-owner $NGINX_UID -m owner ! --uid-owner $ENVOY_UID -m owner ! --uid-owner $LINKERD_UID"
GW_IPS=$(normalize_ips "${MESH_IN_AGENT_GATEWAY_IPS:-}")
for attempt in 1 2 3; do
  if [ -z "$GW_IPS" ]; then
    GW_IPS=$(normalize_ips "$(resolve_gateway_ips)")
  fi
  if [ -n "$GW_IPS" ] && install_redirects "$GW_IPS"; then
    echo "mesh-in-agent: installed and verified OUTPUT redirects for [$GW_IPS]"
    exit 0
  fi
  echo "mesh-in-agent: gateway resolution or redirect installation failed (attempt $attempt/3)" >&2
  if [ "$attempt" -lt 3 ]; then
    sleep "$attempt"
  fi
done

echo "mesh-in-agent: required gateway redirects are unavailable; blocking Pod startup" >&2
exit 1
`, DefaultGatewayHost, NginxWorkerUID(), envoyUID, linkerdUID, HTTPListenPort, HTTPSListenPort)
}

// JWTSecretVolume mounts the caller JWT-SVID Secret issued by callerjwt (name must match AppJWTSecretName).
// Optional is true so a missing Secret does not block Pod startup; fail-closed for
// allowlisted hosts is enforced by njs authDeny (401) when the token file is absent.
func JWTSecretVolume() corev1.Volume {
	return corev1.Volume{
		Name: JWTSecretVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: callerjwt.AppJWTSecretName,
				Optional:   boolPtr(true),
			},
		},
	}
}

// CertsVolume mounts caller TLS material (optional until replica Ready).
// When viewer is non-empty, secretName is olares-mesh-in-tls-<viewer>.
func CertsVolume() corev1.Volume {
	return CertsVolumeForViewer("")
}

// CertsVolumeForViewer mounts olares-mesh-in-tls-<viewer> (or the legacy
// olares-mesh-in-certs name when viewer is empty).
func CertsVolumeForViewer(viewer string) corev1.Volume {
	secretName := "olares-mesh-in-certs"
	viewer = strings.ToLower(strings.TrimSpace(viewer))
	if viewer != "" {
		secretName = constants.MeshInTLSSecretNamePrefix + viewer
	}
	return corev1.Volume{
		Name: CertsVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
				Optional:   boolPtr(true),
			},
		},
	}
}

// SharedHostsVolume mounts auth-hosts and tls-hosts allowlists from ConfigMap.
func SharedHostsVolume() corev1.Volume {
	return corev1.Volume{
		Name: HostsVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: "olares-mesh-in-shared-hosts"},
				Optional:             boolPtr(true),
				Items: []corev1.KeyToPath{
					{Key: SharedHostsFileName, Path: SharedHostsFileName},
					{Key: TLSHostsFileName, Path: TLSHostsFileName},
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

// InsertNativeMeshIn starts mesh-in before application init containers. If
// Linkerd has already injected, keep its init/proxy prefix ahead of mesh-in.
func InsertNativeMeshIn(pod *corev1.Pod, gatewayIPs string) {
	index := 0
	for index < len(pod.Spec.InitContainers) {
		name := pod.Spec.InitContainers[index].Name
		if name != "linkerd-init" && name != "linkerd-proxy" {
			break
		}
		index++
	}
	pod.Spec.InitContainers = slices.Insert(pod.Spec.InitContainers, index,
		InitContainerSpecWithGatewayIPs(gatewayIPs), ContainerSpec())
}

// HasNativeMeshIn reports whether mesh-in remains running after init.
func HasNativeMeshIn(pod *corev1.Pod) bool {
	for _, container := range pod.Spec.InitContainers {
		if container.Name == ContainerName && container.RestartPolicy != nil &&
			*container.RestartPolicy == corev1.ContainerRestartPolicyAlways {
			return true
		}
	}
	return false
}
