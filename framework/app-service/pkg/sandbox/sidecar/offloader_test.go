package sidecar

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/stretchr/testify/require"
)

func TestGetTLSOffloaderContainerSpec_Contract(t *testing.T) {
	configVolumeName := constants.D2ConfVolumeNamePrefix + "abc123"
	got := GetTLSOffloaderContainerSpec(configVolumeName)

	require.Equal(t, constants.D2SidecarContainerName, got.Name)
	require.Equal(t, constants.D2SidecarImageDigest, got.Image)
	require.NotNil(t, got.SecurityContext)
	require.NotNil(t, got.SecurityContext.RunAsUser)
	require.EqualValues(t, constants.D2SidecarUID, *got.SecurityContext.RunAsUser)
	require.NotNil(t, got.SecurityContext.RunAsNonRoot)
	require.True(t, *got.SecurityContext.RunAsNonRoot)
	require.NotNil(t, got.SecurityContext.ReadOnlyRootFilesystem)
	require.True(t, *got.SecurityContext.ReadOnlyRootFilesystem)
	require.NotNil(t, got.SecurityContext.AllowPrivilegeEscalation)
	require.False(t, *got.SecurityContext.AllowPrivilegeEscalation)
	require.NotNil(t, got.SecurityContext.Capabilities)
	require.ElementsMatch(t, []string{"ALL"}, []string{string(got.SecurityContext.Capabilities.Drop[0])})

	require.Len(t, got.Ports, 1)
	require.Equal(t, constants.D2StreamListenPortName, got.Ports[0].Name)
	require.Equal(t, constants.D2StreamListenPort, got.Ports[0].ContainerPort)

	require.Len(t, got.VolumeMounts, 4)
	require.Equal(t, constants.D2CertsVolumeName, got.VolumeMounts[0].Name)
	require.Equal(t, constants.D2NginxCertsDir, got.VolumeMounts[0].MountPath)
	require.True(t, got.VolumeMounts[0].ReadOnly)
	require.Equal(t, configVolumeName, got.VolumeMounts[1].Name)
	require.Equal(t, constants.D2NginxConfigMountPath, got.VolumeMounts[1].MountPath)
	require.Equal(t, constants.D2ConfNginxFileName, got.VolumeMounts[1].SubPath)
	require.True(t, got.VolumeMounts[1].ReadOnly)
	require.Equal(t, configVolumeName, got.VolumeMounts[2].Name)
	require.Equal(t, constants.D2NginxNJSMountPath, got.VolumeMounts[2].MountPath)
	require.Equal(t, constants.D2ConfSharedDecideJSFileName, got.VolumeMounts[2].SubPath)
	require.True(t, got.VolumeMounts[2].ReadOnly)
	require.Equal(t, constants.D2SharedHostsVolumeName, got.VolumeMounts[3].Name)
	require.Equal(t, constants.D2SharedHostsDir, got.VolumeMounts[3].MountPath)
	require.True(t, got.VolumeMounts[3].ReadOnly)

	require.Equal(t, "10m", got.Resources.Requests.Cpu().String())
	require.Equal(t, "48Mi", got.Resources.Requests.Memory().String())
	require.Equal(t, "192Mi", got.Resources.Limits.Memory().String())
	_, hasCPULimit := got.Resources.Limits["cpu"]
	require.False(t, hasCPULimit)
}

func TestGetTLSOffloaderInitContainerSpec_Contract(t *testing.T) {
	got := GetTLSOffloaderInitContainerSpec()
	require.Equal(t, constants.D2SidecarInitContainerName, got.Name)
	require.NotNil(t, got.SecurityContext)
	require.NotNil(t, got.SecurityContext.Privileged)
	require.True(t, *got.SecurityContext.Privileged)
	require.Len(t, got.Args, 2)
	require.Contains(t, got.Args[1], constants.D2IptablesChainName)
	require.Contains(t, got.Args[1], "--dport 443 -j REDIRECT --to-port 15443")
	require.Contains(t, got.Args[1], "--uid-owner 1556 -j RETURN")
	require.NotContains(t, got.Args[1], "--dport 80")
	require.NotContains(t, got.Args[1], "--dport 8080")
}

func TestGetTLSOffloaderVolumes_Contract(t *testing.T) {
	got := GetTLSOffloaderVolumes("alice", "olares-d2-conf-abc123", "olares-d2-conf-abc123")
	require.Len(t, got, 3)

	require.Equal(t, constants.D2CertsVolumeName, got[0].Name)
	require.NotNil(t, got[0].Secret)
	require.Equal(t, "shared-entrance-tls-alice", got[0].Secret.SecretName)
	require.NotNil(t, got[0].Secret.DefaultMode)
	require.EqualValues(t, 0400, *got[0].Secret.DefaultMode)

	require.Equal(t, "olares-d2-conf-abc123", got[1].Name)
	require.NotNil(t, got[1].ConfigMap)
	require.Equal(t, "olares-d2-conf-abc123", got[1].ConfigMap.Name)
	require.Len(t, got[1].ConfigMap.Items, 2)
	require.Equal(t, constants.D2ConfNginxFileName, got[1].ConfigMap.Items[0].Key)
	require.Equal(t, constants.D2ConfSharedDecideJSFileName, got[1].ConfigMap.Items[1].Key)

	require.Equal(t, constants.D2SharedHostsVolumeName, got[2].Name)
	require.NotNil(t, got[2].ConfigMap)
	require.Equal(t, constants.D2SharedHostsVolumeName, got[2].ConfigMap.Name)
	require.Len(t, got[2].ConfigMap.Items, 1)
	require.Equal(t, constants.D2SharedHostsFileName, got[2].ConfigMap.Items[0].Key)
}

func TestRenderNginxConf_Contract(t *testing.T) {
	got := RenderNginxConf("alice", []string{"bob"}, "example.com", "user-space-alice", 8081)
	require.Contains(t, got, "load_module /etc/nginx/modules/ngx_stream_js_module.so;")
	require.Contains(t, got, "worker_processes 1;")
	require.Contains(t, got, "worker_shutdown_timeout 30s;")
	require.Contains(t, got, "js_path /etc/nginx/njs/;")
	require.Contains(t, got, "js_import shared_decide.js;")
	require.Contains(t, got, "js_set $offload_upstream shared_decide.decideOffload;")
	require.Contains(t, got, "ssl_preread on;")
	require.Contains(t, got, "ssl_certificate_cache max=16 inactive=10m valid=1m;")
	require.Contains(t, got, "server app-gateway-data.user-space-alice:8081;")
	require.Contains(t, got, "if ($host ~* \"^[a-z0-9-]+\\.shared\\.example\\.com$\") {")
	require.Contains(t, got, "if ($host !~* \"^[a-z0-9-]+\\.(alice|bob)\\.example\\.com$\") {")
	require.Contains(t, got, "proxy_set_header X-Real-IP $remote_addr;")
	require.Contains(t, got, "proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;")
	require.Contains(t, got, "proxy_set_header X-Forwarded-Proto https;")
	require.Contains(t, got, "proxy_set_header X-Forwarded-Host $host;")
	require.Contains(t, got, "proxy_set_header X-Original-URI $request_uri;")
	require.NotContains(t, got, ":8080;")
	require.NotContains(t, got, ":80;")
	require.Less(t,
		strings.Index(got, "load_module /etc/nginx/modules/ngx_stream_js_module.so;"),
		strings.Index(got, "js_import shared_decide.js;"),
	)
}

func TestRenderSharedDecideJS_Contract(t *testing.T) {
	got := RenderSharedDecideJS("Example.COM", constants.D2SidecarHostsFilePath)
	require.Contains(t, got, "const HOSTS_FILE = '/etc/d2/shared-hosts.txt';")
	require.Contains(t, got, "const CACHE_TTL_MS = 5000;")
	require.Contains(t, got, "const V2_GUARD = new RegExp('^[a-z0-9-]+")
	require.Contains(t, got, "shared")
	require.Contains(t, got, "ESCAPED_PLATFORM_DOMAIN + '$', 'i');")
	require.Contains(t, got, "function decideOffload(s)")
	require.Contains(t, got, "return '127.0.0.1:15080';")
	require.Contains(t, got, "return passthrough(host);")
}

func TestGenerateD2IptablesCommands_Contract(t *testing.T) {
	got := generateD2IptablesCommands(constants.D2StreamListenPort)
	require.Contains(t, got, ":D2_OUTBOUND - [0:0]")
	require.Contains(t, got, "-A OUTPUT -p tcp -j D2_OUTBOUND")
	require.Contains(t, got, "-A D2_OUTBOUND -m owner --uid-owner 1556 -j RETURN")
	require.Contains(t, got, "-A D2_OUTBOUND -d 127.0.0.1/32 -j RETURN")
	require.Contains(t, got, "-A D2_OUTBOUND -p tcp --dport 443 -j REDIRECT --to-port 15443")
	require.Contains(t, got, "-A D2_OUTBOUND -j RETURN")
	require.NotContains(t, got, "--dport 80")
	require.NotContains(t, got, "--dport 8080")
}

func TestD2Constants_UIDAndNames(t *testing.T) {
	require.EqualValues(t, 1556, constants.D2SidecarUID)
	require.NotEqual(t, constants.EnvoyUID, constants.D2SidecarUID)
	require.NotEqual(t, constants.LinkerdProxyUID, constants.D2SidecarUID)
	require.Equal(t, "olares-d2-sidecar", constants.D2SidecarContainerName)
	require.Equal(t, "/etc/d2/shared-hosts.txt", constants.D2SidecarHostsFilePath)
}

func TestEnvoyGoldenDiff(t *testing.T) {
	expectedCmd := `iptables-restore --noflush <<EOF
# sidecar interception rules
*nat
:PROXY_IN_REDIRECT - [0:0]
:PROXY_INBOUND - [0:0]
:PROXY_OUTBOUND - [0:0]
:PROXY_OUT_REDIRECT - [0:0]

-A PREROUTING -p tcp -j PROXY_INBOUND
-A OUTPUT -p tcp -j PROXY_OUTBOUND
-A PROXY_INBOUND -p tcp --dport 15000 -j RETURN
-A PROXY_INBOUND -s 20.20.20.21 -j RETURN
-A PROXY_INBOUND -s 172.30.0.0/16 -j RETURN
-A PROXY_INBOUND -p tcp -j PROXY_IN_REDIRECT
-A PROXY_IN_REDIRECT -p tcp -j REDIRECT --to-port 15003
-A PROXY_OUTBOUND -d ${POD_IP}/32 -j RETURN
-A PROXY_OUTBOUND -o lo ! -d 127.0.0.1/32 -m owner --uid-owner 1555 -j PROXY_IN_REDIRECT
-A PROXY_OUTBOUND -o lo -m owner ! --uid-owner 1555 -j RETURN
-A PROXY_OUTBOUND -m owner --uid-owner 1555 -j RETURN
-A PROXY_OUTBOUND -m owner --uid-owner 2102 -j RETURN
-A PROXY_OUTBOUND -d 127.0.0.1/32 -j RETURN
-A PROXY_OUTBOUND -p tcp -m multiport ! --dports 80,8080 -j RETURN
-A PROXY_OUTBOUND -j PROXY_OUT_REDIRECT
-A PROXY_OUT_REDIRECT -p tcp -j REDIRECT --to-port 15001

COMMIT
EOF
`
	require.Equal(t, expectedCmd, generateIptablesCommands(nil, false))

	initSpec := GetInitContainerSpec(nil, false)
	require.Equal(t, constants.SidecarInitContainerName, initSpec.Name)
	require.Equal(t, "beclab/init:v1.2.3", initSpec.Image)
	require.Len(t, initSpec.Args, 2)
	require.Equal(t, expectedCmd, initSpec.Args[1])
	require.Len(t, initSpec.Env, 1)
	require.Equal(t, "POD_IP", initSpec.Env[0].Name)
	require.NotNil(t, initSpec.Env[0].ValueFrom)
	require.NotNil(t, initSpec.Env[0].ValueFrom.FieldRef)
	require.Equal(t, "v1", initSpec.Env[0].ValueFrom.FieldRef.APIVersion)
	require.Equal(t, "status.podIP", initSpec.Env[0].ValueFrom.FieldRef.FieldPath)
}
