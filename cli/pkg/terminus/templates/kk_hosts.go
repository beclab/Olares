package templates

import (
	"github.com/lithammer/dedent"
	"text/template"
)

var UpdateKKHostsScriptTmpl = template.Must(template.New("initOS.sh").Parse(
	dedent.Dedent(`#!/usr/bin/env bash
sed -i ':a;$!{N;ba};s@# kubekey hosts BEGIN.*# kubekey hosts END@@' /etc/hosts
sed -i '/^$/N;/\n$/N;//D' /etc/hosts

cat >>/etc/hosts<<EOF
# kubekey hosts BEGIN
{{- range .Hosts }}
{{ . }}
{{- end }}
# kubekey hosts END
EOF
`)))
