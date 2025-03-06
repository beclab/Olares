{{/*
Expand the name of the chart.
*/}}
{{- define "hami-vgpu.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "hami-vgpu.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
The app name for Scheduler
*/}}
{{- define "hami-vgpu.scheduler" -}}
{{- printf "%s-scheduler" ( include "hami-vgpu.fullname" . ) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
The app name for DevicePlugin
*/}}
{{- define "hami-vgpu.device-plugin" -}}
{{- printf "%s-device-plugin" ( include "hami-vgpu.fullname" . ) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
The tls secret name for Scheduler
*/}}
{{- define "hami-vgpu.scheduler.tls" -}}
{{- printf "%s-scheduler-tls" ( include "hami-vgpu.fullname" . ) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
The webhook name
*/}}
{{- define "hami-vgpu.scheduler.webhook" -}}
{{- printf "%s-webhook" ( include "hami-vgpu.fullname" . ) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "hami-vgpu.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "hami-vgpu.labels" -}}
helm.sh/chart: {{ include "hami-vgpu.chart" . }}
{{ include "hami-vgpu.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "hami-vgpu.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hami-vgpu.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Image registry secret name
*/}}
{{- define "hami-vgpu.imagePullSecrets" -}}
imagePullSecrets: {{ toYaml .Values.imagePullSecrets | nindent 2 }}
{{- end }}

{{/*
    Resolve the tag for kubeScheduler.
*/}}
{{- define "resolvedKubeSchedulerTag" -}}
{{- if .Values.scheduler.kubeScheduler.imageTag }}
{{- .Values.scheduler.kubeScheduler.imageTag | trim -}}
{{- else }}
{{- include "strippedKubeVersion" . | trim -}}
{{- end }}
{{- end }}


{{/*
    Return the stripped Kubernetes version string by removing extra parts after semantic version number.
    v1.31.1+k3s1 -> v1.31.1
    v1.30.8-eks-2d5f260 -> v1.30.8
    v1.31.1 -> v1.31.1
*/}}
{{- define "strippedKubeVersion" -}}
{{ regexReplaceAll "^(v[0-9]+\\.[0-9]+\\.[0-9]+)(.*)$" .Capabilities.KubeVersion.Version "$1" }}
{{- end -}}
