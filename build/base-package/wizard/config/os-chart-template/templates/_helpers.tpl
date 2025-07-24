{{/*
Expand the name of the chart.
*/}}
{{- define "system.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "system.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "system.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "system.labels" -}}
helm.sh/chart: {{ include "system.chart" . }}
{{ include "system.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "system.selectorLabels" -}}
app.kubernetes.io/name: {{ include "system.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "system.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "system.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "opentelemetry-operator.fullname" -}}
{{- "otel-opentelemetry-operator" }}
{{- end }}

{{- define "opentelemetry-operator.WebhookCert" -}}
{{- $caCertEnc := "" }}
{{- $certCrtEnc := "" }}
{{- $certKeyEnc := "" }}
{{- $prevSecret := (lookup "v1" "Secret" .Release.Namespace (printf "%s-controller-manager-service-cert" (include "opentelemetry-operator.fullname" .) )) }}
{{- if $prevSecret }}
{{- $certCrtEnc = index $prevSecret "data" "tls.crt" }}
{{- $certKeyEnc = index $prevSecret "data" "tls.key" }}
{{- $caCertEnc = index $prevSecret "data" "ca.crt" }}
{{- else }}
{{- $altNames := list ( printf "%s-webhook.%s" (include "opentelemetry-operator.fullname" .) .Release.Namespace ) ( printf "%s-webhook.%s.svc" (include "opentelemetry-operator.fullname" .) .Release.Namespace ) -}}
{{- $tmpperioddays := 36500 }}
{{- $ca := genCA "opentelemetry-operator-operator-ca" $tmpperioddays }}
{{- $cert := genSignedCert (include "opentelemetry-operator.fullname" .) nil $altNames $tmpperioddays $ca }}
{{- $certCrtEnc = b64enc $cert.Cert }}
{{- $certKeyEnc = b64enc $cert.Key }}
{{- $caCertEnc = b64enc $ca.Cert }}
{{- end }}
{{- $result := dict "crt" $certCrtEnc "key" $certKeyEnc "ca" $caCertEnc }}
{{- $result | toYaml }}
{{- end }}
