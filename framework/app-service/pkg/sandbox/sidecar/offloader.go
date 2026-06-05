package sidecar

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"
)

// GetTLSOffloaderContainerSpec returns the olares-d2-sidecar specification.
func GetTLSOffloaderContainerSpec(configVolumeName string) corev1.Container {
	return corev1.Container{
		Name:            constants.D2SidecarContainerName,
		Image:           constants.D2SidecarImageDigest,
		// behavior: local-test image is imported into containerd by tag only (no
		// registry repoDigests), so PullNever forces the kubelet to use the
		// preloaded image and never reach out to a remote registry.
		ImagePullPolicy: corev1.PullNever,
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: pointer.BoolPtr(false),
			ReadOnlyRootFilesystem:   pointer.BoolPtr(true),
			RunAsNonRoot:             pointer.BoolPtr(true),
			RunAsUser:                pointer.Int64Ptr(constants.D2SidecarUID),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          constants.D2StreamListenPortName,
				ContainerPort: constants.D2StreamListenPort,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      constants.D2CertsVolumeName,
				MountPath: constants.D2NginxCertsDir,
				ReadOnly:  true,
			},
			{
				Name:      configVolumeName,
				MountPath: constants.D2NginxConfigMountPath,
				SubPath:   constants.D2ConfNginxFileName,
				ReadOnly:  true,
			},
			{
				Name:      configVolumeName,
				MountPath: constants.D2NginxNJSMountPath,
				SubPath:   constants.D2ConfSharedDecideJSFileName,
				ReadOnly:  true,
			},
			{
				Name:      constants.D2SharedHostsVolumeName,
				MountPath: constants.D2SharedHostsDir,
				ReadOnly:  true,
			},
			{
				Name:      constants.D2NginxCacheVolumeName,
				MountPath: constants.D2NginxCacheDir,
			},
			{
				Name:      constants.D2NginxRunVolumeName,
				MountPath: constants.D2NginxRunDir,
			},
		},
		Command: []string{"nginx"},
		Args: []string{
			"-g",
			"daemon off;",
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("48Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("192Mi"),
			},
		},
	}
}

// GetTLSOffloaderInitContainerSpec returns init container spec for d2 interception rules.
func GetTLSOffloaderInitContainerSpec() corev1.Container {
	enablePrivileged := true
	return corev1.Container{
		Name:            constants.D2SidecarInitContainerName,
		Image:           "beclab/init:v1.2.3",
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: &corev1.SecurityContext{
			Privileged: &enablePrivileged,
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"NET_ADMIN"},
			},
			RunAsNonRoot: pointer.BoolPtr(false),
			RunAsUser:    pointer.Int64Ptr(0),
		},
		Command: []string{"/bin/sh"},
		Args: []string{
			"-c",
			generateD2IptablesCommands(constants.D2StreamListenPort),
		},
	}
}

// GetTLSOffloaderVolumes returns cert/config/shared-host volume specs for d2.
func GetTLSOffloaderVolumes(viewer, nginxConfConfigMapName, nginxConfVolumeName string) []corev1.Volume {
	defaultMode := int32(0400)
	return []corev1.Volume{
		{
			Name: constants.D2CertsVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  constants.D2SharedTLSSecretNamePrefix + viewer,
					DefaultMode: &defaultMode,
				},
			},
		},
		{
			Name: nginxConfVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: nginxConfConfigMapName,
					},
					Items: []corev1.KeyToPath{
						{
							Key:  constants.D2ConfNginxFileName,
							Path: constants.D2ConfNginxFileName,
						},
						{
							Key:  constants.D2ConfSharedDecideJSFileName,
							Path: constants.D2ConfSharedDecideJSFileName,
						},
					},
				},
			},
		},
		{
			Name: constants.D2SharedHostsVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: constants.D2SharedHostsVolumeName,
					},
					Items: []corev1.KeyToPath{
						{
							Key:  constants.D2SharedHostsFileName,
							Path: constants.D2SharedHostsFileName,
						},
					},
				},
			},
		},
		{
			Name: constants.D2NginxCacheVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: constants.D2NginxRunVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
}

// RenderNginxConf renders d2 nginx.conf with stream preread and http offload.
func RenderNginxConf(viewer string, allowset []string, platformDomain, gatewayNamespace string, strongIdentityServicePort int32) string {
	normalizedViewerSet := normalizeViewerAllowset(viewer, allowset)
	viewerHostPattern := buildViewerHostPattern(normalizedViewerSet, platformDomain)
	v2HostPattern := buildV2SharedHostPattern(platformDomain)

	return fmt.Sprintf(`load_module /etc/nginx/modules/ngx_stream_js_module.so;
pid %s/nginx.pid;
error_log /dev/stderr warn;
worker_processes 1;
worker_shutdown_timeout %s;

events {
  worker_connections 1024;
}

stream {
  js_path %s;
  js_import shared_decide.js;
  js_set $offload_upstream shared_decide.decideOffload;
  resolver kube-dns.kube-system.svc.cluster.local ipv6=off valid=30s;

  server {
    listen %d;
    proxy_connect_timeout 5s;
    proxy_timeout 3600s;
    ssl_preread on;
    proxy_pass $offload_upstream;
  }
}

http {
  client_body_temp_path %s/client_temp;
  proxy_temp_path %s/proxy_temp;
  fastcgi_temp_path %s/fastcgi_temp;
  uwsgi_temp_path %s/uwsgi_temp;
  scgi_temp_path %s/scgi_temp;

  map $ssl_server_name $tls_cert_path {
    default %s/tls.crt;
  }
  map $ssl_server_name $tls_key_path {
    default %s/tls.key;
  }

  upstream d2_strong_identity_upstream {
    server app-gateway-data.%s:%d;
    keepalive 64;
  }

  server {
    listen 127.0.0.1:%d ssl;
    ssl_certificate $tls_cert_path;
    ssl_certificate_key $tls_key_path;
    ssl_certificate_cache max=%d inactive=%s valid=%s;
    ssl_session_cache shared:d2_ssl_cache:20m;
    ssl_session_timeout 10m;

    if ($host ~* "%s") {
      return 421;
    }
    if ($host !~* "%s") {
      return 421;
    }

    location / {
      proxy_http_version 1.1;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto https;
      proxy_set_header X-Forwarded-Host $host;
      proxy_set_header X-Original-URI $request_uri;
      proxy_pass http://d2_strong_identity_upstream;
    }
  }
}
`,
		constants.D2NginxRunDir,
		constants.D2WorkerShutdownTimeout,
		constants.D2NginxNJSDir,
		constants.D2StreamListenPort,
		constants.D2NginxCacheDir,
		constants.D2NginxCacheDir,
		constants.D2NginxCacheDir,
		constants.D2NginxCacheDir,
		constants.D2NginxCacheDir,
		constants.D2NginxCertsDir,
		constants.D2NginxCertsDir,
		gatewayNamespace,
		strongIdentityServicePort,
		constants.D2HTTPLoopbackPort,
		constants.D2CertCacheMax,
		constants.D2CertCacheInactive,
		constants.D2CertCacheValid,
		v2HostPattern,
		viewerHostPattern,
	)
}

// RenderSharedDecideJS renders the njs offload decision script.
func RenderSharedDecideJS(platformDomain string, hostsFilePath string) string {
	script := fmt.Sprintf(`
var fs = require('fs');

const HOSTS_FILE = '%s';
const CACHE_TTL_MS = 5000;

let cachedHosts = null;
let cachedMtimeMs = 0;
let cachedAtMs = 0;

const ESCAPED_PLATFORM_DOMAIN = '__ESCAPED_PLATFORM_DOMAIN__';
const PLATFORM_SUFFIX = '.' + ESCAPED_PLATFORM_DOMAIN;
const V2_GUARD = new RegExp('^[a-z0-9-]+\\\\.shared\\\\.' + ESCAPED_PLATFORM_DOMAIN + '$', 'i');

function nowMs() {
  return Date.now();
}

function normalizeHost(v) {
  if (!v) {
    return '';
  }
  return ('' + v).trim().toLowerCase();
}

function passthrough(host) {
  if (!host) {
    return 'invalid.local:443';
  }
  return host + ':443';
}

function isPlatformHost(host) {
  return host.length > PLATFORM_SUFFIX.length && host.endsWith(PLATFORM_SUFFIX);
}

function parseHosts(content) {
  const out = {};
  const lines = content.split(/\r?\n/);
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i].trim().toLowerCase();
    if (!line || line.startsWith('#')) {
      continue;
    }
    out[line] = true;
  }
  return out;
}

function reloadHostsIfNeeded() {
  const now = nowMs();
  if (cachedHosts !== null && now-cachedAtMs < CACHE_TTL_MS) {
    return cachedHosts;
  }

  try {
    const st = fs.statSync(HOSTS_FILE);
    if (cachedHosts !== null && st.mtimeMs === cachedMtimeMs && now-cachedAtMs < CACHE_TTL_MS) {
      return cachedHosts;
    }
    const content = fs.readFileSync(HOSTS_FILE, 'utf8');
    cachedHosts = parseHosts(content);
    cachedMtimeMs = st.mtimeMs;
    cachedAtMs = now;
    return cachedHosts;
  } catch (e) {
    cachedAtMs = now;
    if (cachedHosts === null) {
      cachedHosts = {};
    }
    return cachedHosts;
  }
}

function decideOffload(s) {
  // njs stream js_set handler receives the session object; the SNI parsed by
  // "ssl_preread on" is exposed as s.variables.ssl_preread_server_name.
  // Stringifying the session object itself yields "[object stream session]",
  // which nginx then rejects as an invalid upstream address.
  const host = normalizeHost(s.variables.ssl_preread_server_name);
  if (!host) {
    return passthrough(host);
  }
  if (V2_GUARD.test(host)) {
    return passthrough(host);
  }
  if (!isPlatformHost(host)) {
    return passthrough(host);
  }

  const hosts = reloadHostsIfNeeded();
  if (hosts[host]) {
    return '127.0.0.1:%d';
  }
  return passthrough(host);
}

export default { decideOffload };
`, hostsFilePath, constants.D2HTTPLoopbackPort)

	escapedPlatformDomain := regexp.QuoteMeta(strings.ToLower(strings.TrimSpace(platformDomain)))
	return strings.ReplaceAll(script, "__ESCAPED_PLATFORM_DOMAIN__", escapedPlatformDomain)
}

func generateD2IptablesCommands(tlsOffloadPort int32) string {
	return fmt.Sprintf(`iptables-restore --noflush <<EOF
*nat
:%s - [0:0]

-A OUTPUT -p tcp -j %s
-A %s -m owner --uid-owner %d -j RETURN
-A %s -d 127.0.0.1/32 -j RETURN
-A %s -p tcp --dport 443 -j REDIRECT --to-port %d
-A %s -j RETURN

COMMIT
EOF
`,
		constants.D2IptablesChainName,
		constants.D2IptablesChainName,
		constants.D2IptablesChainName, constants.D2SidecarUID,
		constants.D2IptablesChainName,
		constants.D2IptablesChainName, tlsOffloadPort,
		constants.D2IptablesChainName,
	)
}

func normalizeViewerAllowset(viewer string, allowset []string) []string {
	normalized := make([]string, 0, len(allowset)+1)
	seen := make(map[string]struct{}, len(allowset)+1)
	addViewer := func(v string) {
		clean := strings.ToLower(strings.TrimSpace(v))
		if clean == "" {
			return
		}
		if _, ok := seen[clean]; ok {
			return
		}
		seen[clean] = struct{}{}
		normalized = append(normalized, clean)
	}

	addViewer(viewer)
	for _, v := range allowset {
		addViewer(v)
	}
	if len(normalized) == 0 {
		normalized = append(normalized, "invalid-viewer")
	}
	sort.Strings(normalized)
	return normalized
}

func buildViewerHostPattern(viewers []string, platformDomain string) string {
	escapedDomain := regexp.QuoteMeta(strings.ToLower(strings.TrimSpace(platformDomain)))
	return fmt.Sprintf("^[a-z0-9-]+\\.(%s)\\.%s$", strings.Join(viewers, "|"), escapedDomain)
}

func buildV2SharedHostPattern(platformDomain string) string {
	escapedDomain := regexp.QuoteMeta(strings.ToLower(strings.TrimSpace(platformDomain)))
	return fmt.Sprintf("^[a-z0-9-]+\\.shared\\.%s$", escapedDomain)
}
