{{/*
Generate CA-signed TLS certificates for search3-validation webhook
First generates a CA certificate, then uses it to sign the server certificate.
This ensures the caBundle (CA cert) can properly verify the server's TLS certificate.
*/}}
{{- define "search3-validation.certs" -}}
{{- $altNames := list (printf "search3-validation.%s" .Release.Namespace) (printf "search3-validation.%s.svc" .Release.Namespace) -}}
{{- $ca := genCA "search3-validation" 36500 }}
{{- $cert := genSignedCert "search3-validation" nil $altNames 36500 $ca }}
{{- $result := dict "caCert" (b64enc $ca.Cert) "tlsCert" (b64enc $cert.Cert) "tlsKey" (b64enc $cert.Key) }}
{{- $result | toYaml }}
{{- end }}
