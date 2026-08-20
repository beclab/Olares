{{/*
Expand the name of the chart.
*/}}
{{- define "helm-charts-exporter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "helm-charts-exporter.fullname" -}}
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
{{- define "helm-charts-exporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "helm-charts-exporter.labels" -}}
helm.sh/chart: {{ include "helm-charts-exporter.chart" . }}
{{ include "helm-charts-exporter.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "helm-charts-exporter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "helm-charts-exporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "helm-charts-exporter.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "helm-charts-exporter.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Determine the exporter app label based on monitoring configuration
*/}}
{{- define "metrics-exporter.appLabelWithReleaseName" }}
{{- if .Values.monitor.resources.gpu }}
app: {{ .Release.Name }}-amdgpu-metrics-exporter
{{- else }}
app: {{ .Release.Name }}-amd-device-metrics-exporter
{{- end }}
{{- end }}

{{/*
Determine the exporter app label based on monitoring configuration
*/}}
{{- define "metrics-exporter.appLabel" }}
{{- if .Values.monitor.resources.gpu }}
app: amdgpu-metrics-exporter
{{- else }}
app: amd-device-metrics-exporter
{{- end }}
{{- end }}

{{/*
Determine the exporter name based on monitoring configuration
*/}}
{{- define "metrics-exporter.nameWithRelease" }}
{{- if .Values.monitor.resources.gpu }}
name: {{ .Release.Name }}-amdgpu-metrics-exporter
{{- else }}
name: {{ .Release.Name }}-amd-device-metrics-exporter
{{- end }}
{{- end }}

{{/*
Determine the exporter name based on monitoring configuration
*/}}
{{- define "metrics-exporter.name" }}
{{- if .Values.monitor.resources.gpu }}
name: amdgpu-metrics-exporter
{{- else }}
name: amd-device-metrics-exporter
{{- end }}
{{- end }}

{{/*
Determine the exporter labels based on monitoring configuration
*/}}
{{- define "metrics-exporter.labels" -}}
{{- if .Values.monitor.resources.gpu }}
app.kubernetes.io/component: amd-gpu
app.kubernetes.io/part-of: amd-gpu
{{- else }}
app.kubernetes.io/component: amd-device
app.kubernetes.io/part-of: amd-device
{{- end }}
{{- end }}