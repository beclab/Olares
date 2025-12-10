{{/*
Generate CA-signed TLS certificates for search3-validation webhook
First generates a CA certificate, then uses it to sign the server certificate.
This ensures the caBundle (CA cert) can properly verify the server's TLS certificate.

IMPORTANT: Uses lookup to check if certificates already exist in the Secret.
If they exist, reuse them to maintain consistency between Secret and ValidatingWebhookConfiguration.
If they don't exist, generate new certificates.
*/}}
{{- define "search3-validation.certs" -}}
{{- $secretName := "search3-validation-tls" }}
{{- $existingSecret := lookup "v1" "Secret" .Release.Namespace $secretName }}
{{- $hasAllCerts := false }}
{{- if and $existingSecret $existingSecret.data }}
  {{- $hasCaCrt := and (index $existingSecret.data "ca.crt") (ne (index $existingSecret.data "ca.crt") "") }}
  {{- $hasTlsCrt := and (index $existingSecret.data "tls.crt") (ne (index $existingSecret.data "tls.crt") "") }}
  {{- $hasTlsKey := and (index $existingSecret.data "tls.key") (ne (index $existingSecret.data "tls.key") "") }}
  {{- $hasAllCerts = and $hasCaCrt $hasTlsCrt $hasTlsKey }}
{{- end }}
{{- if $hasAllCerts }}
  {{/* Reuse existing certificates from Secret - all three fields (ca.crt, tls.crt, tls.key) are present and non-empty */}}
  {{- $result := dict "caCert" (index $existingSecret.data "ca.crt") "tlsCert" (index $existingSecret.data "tls.crt") "tlsKey" (index $existingSecret.data "tls.key") }}
  {{- $result | toYaml }}
{{- else }}
  {{/* Generate new certificates - Secret doesn't exist or is incomplete */}}
  {{- $altNames := list "search3-validation" (printf "search3-validation.%s" .Release.Namespace) (printf "search3-validation.%s.svc" .Release.Namespace) (printf "search3-validation.%s.svc.cluster.local" .Release.Namespace) -}}
  {{- $ca := genCA "search3-validation" 36500 }}
  {{- $cert := genSignedCert "search3-validation" nil $altNames 36500 $ca }}
  {{- $result := dict "caCert" (b64enc $ca.Cert) "tlsCert" (b64enc $cert.Cert) "tlsKey" (b64enc $cert.Key) }}
  {{- $result | toYaml }}
{{- end }}
{{- end }}
