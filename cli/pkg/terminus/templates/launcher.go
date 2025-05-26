package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var BFLValues = template.Must(template.New("values.yaml").Parse(
	dedent.Dedent(`bfl:
  terminus_cert_service_api: {{ .TerminusCertServiceAPI }}
  terminus_dns_service_api: {{ .TerminusDNSServiceAPI }}  
`),
))
