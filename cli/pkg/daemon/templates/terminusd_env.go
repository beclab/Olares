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
    `)))
