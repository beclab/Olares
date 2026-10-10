package templates

import "text/template"

var SettingsValue = template.Must(template.New("values.yaml").Parse(`cluster_id: {{ .ClusterID }}
s3_sts: {{ .S3SessionToken }}
s3_ak: {{ .S3AccessKey }}
s3_sk: {{ .S3SecretKey }}
domainName: ''
selfHosted: 'false'
terminusd: '{{ .TerminusdInstalled }}'
`))
