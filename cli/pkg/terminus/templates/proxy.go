package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var ReverseProxyConfigMap = template.Must(template.New("cm-default-reverse-proxy-config.yaml").Parse(
	dedent.Dedent(`apiVersion: v1
data:
  cloudflare.enable: "{{ .EnableCloudflare }}"
  frp.enable: "{{ .EnableFrp }}"
  frp.server: "{{ .FrpServer }}"
  frp.port: "{{ .FrpPort }}"
  frp.auth_method: "{{ .FrpAuthMethod }}"
  frp.auth_token: "{{ .FrpAuthToken }}"
kind: ConfigMap
metadata:
  name: default-reverse-proxy-config
  namespace: os-system`),
))
