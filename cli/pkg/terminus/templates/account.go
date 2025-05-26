package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var AccountValues = template.Must(template.New("values.yaml").Parse(
	dedent.Dedent(`user:
  name: '{{ .UserName }}'
  password: '{{ .Password }}'
  email: '{{ .Email }}'
  terminus_name: '{{ .UserName }}@{{ .DomainName }}'
`),
))
