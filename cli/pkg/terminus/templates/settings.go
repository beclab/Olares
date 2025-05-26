package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var SettingsValue = template.Must(template.New("values.yaml").Parse(
	dedent.Dedent(`namespace:
  name: 'user-space-{{ .UserName }}'
  role: admin

cluster_id: {{ .ClusterID }}
s3_sts: {{ .S3SessionToken }}
s3_ak: {{ .S3AccessKey }}
s3_sk: {{ .S3SecretKey }}
domainName: '{{ .DomainName }}'
selfHosted: '{{ .SelfHosted }}'
terminusd: '{{ .TerminusdInstalled }}'

user:
  name: '{{ .UserName }}'
`),
))
