package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

// TerminusdEnv defines the template of terminusd's env.
var TerminusdEnv = template.Must(template.New("olaresd.service.env").Parse(
	dedent.Dedent(`# Environment file for olaresd
INSTALLED_VERSION={{ .Version }}
KUBE_TYPE={{ .KubeType }}
REGISTRY_MIRRORS={{ .RegistryMirrors }}
BASE_DIR={{ .BaseDir }}
LOCAL_GPU_ENABLE={{ .GpuEnable }}
LOCAL_GPU_SHARE={{ .GpuShare }}
PUBLICLY_ACCESSIBLE={{ .PubliclyAccessible }}
CLOUDFLARE_ENABLE={{ .CloudflareEnable }}
FRP_ENABLE={{ .FrpEnable }}
FRP_SERVER={{ .FrpServer }}
FRP_PORT={{ .FrpPort }}
FRP_AUTH_METHOD={{ .FrpAuthMethod }}
FRP_AUTH_TOKEN=
TOKEN_MAX_AGE={{ .TokenMaxAge }}
MARKET_PROVIDER={{ .MarketProvider }}
TERMINUS_CERT_SERVICE_API={{ .TerminusCertServiceAPI }}
TERMINUS_DNS_SERVICE_API={{ .TerminusDNSServiceAPI }}
{{- range $key, $val := .TerminusGlobalEnvs }}
{{ $key }}={{ $val }}
{{- end }}
    `)))
