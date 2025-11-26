{{/*
Generate self-signed TLS certificates for search3-validation webhook
*/}}
{{- define "search3-validation.certs" -}}
{{- $altNames := list (printf "search3-validation.%s" .Release.Namespace) (printf "search3-validation.%s.svc" .Release.Namespace) -}}
{{- $cert := genSelfSignedCert "search3-validation" nil $altNames 36500 }}
{{- $result := dict "tlsCert" (b64enc $cert.Cert) "tlsKey" (b64enc $cert.Key) }}
{{- $result | toYaml }}
{{- end }}
