{{/*
Allow the release namespace to be overridden.
*/}}
{{- define "eg.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Expand the name of the chart.
*/}}
{{- define "eg.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "eg.fullname" -}}
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
{{- define "eg.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "eg.labels" -}}
helm.sh/chart: {{ include "eg.chart" . }}
{{ include "eg.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "eg.selectorLabels" -}}
app.kubernetes.io/name: {{ include "eg.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "eg.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "eg.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
The name of the Envoy Gateway image.
*/}}
{{- define "eg.image" -}}
{{/* if deployment-specific repository is defined, it takes precedence */}}
{{- if .Values.deployment.envoyGateway.image.repository -}}
{{/*  if global.imageRegistry is defined, it takes precedence always */}}
{{-   if .Values.global.imageRegistry -}}
{{-     $repositoryParts := splitn "/" 2 .Values.deployment.envoyGateway.image.repository -}}
{{-     $registryName := .Values.global.imageRegistry -}}
{{-     $repositoryName := $repositoryParts._1 -}}
{{-     $imageTag := default .Chart.AppVersion .Values.deployment.envoyGateway.image.tag -}}
{{-     printf "%s/%s:%s" $registryName $repositoryName $imageTag -}}
{{/*  if global.imageRegistry is undefined, take repository as is */}}
{{-   else -}}
{{-     $imageTag := default .Chart.AppVersion .Values.deployment.envoyGateway.image.tag -}}
{{-     printf "%s:%s" .Values.deployment.envoyGateway.image.repository $imageTag -}}
{{-   end -}}
{{/* else, global image is used if defined */}}
{{- else if .Values.global.images.envoyGateway.image -}}
{{-   $imageParts := splitn "/" 2 .Values.global.images.envoyGateway.image -}}
{{/*    if global.imageRegistry is defined, it takes precedence always */}}
{{-   $registryName := default $imageParts._0 .Values.global.imageRegistry -}}
{{-   $repositoryTag := $imageParts._1 -}}
{{-   $repositoryParts := splitn ":" 2 $repositoryTag -}}
{{-   $repositoryName := $repositoryParts._0 -}}
{{-   $imageTag := $repositoryParts._1 -}}
{{-   printf "%s/%s:%s" $registryName $repositoryName $imageTag -}}
{{- else -}}
docker.io/envoyproxy/gateway:{{ .Chart.Version }}
{{- end -}}
{{- end -}}

{{/*
Pull policy for the Envoy Gateway image.
*/}}
{{- define "eg.image.pullPolicy" -}}
{{- default .Values.deployment.envoyGateway.imagePullPolicy .Values.global.images.envoyGateway.pullPolicy -}}
{{- end }}

{{/*
Pull secrets for the Envoy Gateway image.
*/}}
{{- define "eg.image.pullSecrets" -}}
{{- if .Values.global.imagePullSecrets -}}
imagePullSecrets:
{{ toYaml .Values.global.imagePullSecrets }}
{{- else if .Values.deployment.envoyGateway.imagePullSecrets -}}
imagePullSecrets:
{{ toYaml .Values.deployment.envoyGateway.imagePullSecrets }}
{{- else if .Values.global.images.envoyGateway.pullSecrets -}}
imagePullSecrets:
{{ toYaml .Values.global.images.envoyGateway.pullSecrets }}
{{- else -}}
imagePullSecrets: {{ toYaml list }}
{{- end }}
{{- end }}

{{/*
The name of the Envoy Ratelimit image.
*/}}
{{- define "eg.ratelimit.image" -}}
{{-   $imageParts := splitn "/" 2 .Values.global.images.ratelimit.image -}}
{{/*    if global.imageRegistry is defined, it takes precedence always */}}
{{-   $registryName := default $imageParts._0 .Values.global.imageRegistry -}}
{{-   $repositoryTag := $imageParts._1 -}}
{{-   $repositoryParts := splitn ":" 2 $repositoryTag -}}
{{-   $repositoryName := $repositoryParts._0 -}}
{{-   $imageTag := default "master" $repositoryParts._1 -}}
{{-   printf "%s/%s:%s" $registryName $repositoryName $imageTag -}}
{{- end -}}

{{/*
Pull secrets for the Envoy Ratelimit image.
*/}}
{{- define "eg.ratelimit.image.pullSecrets" -}}
{{- if .Values.global.imagePullSecrets }}
imagePullSecrets:
{{ toYaml .Values.global.imagePullSecrets }}
{{- else if .Values.global.images.ratelimit.pullSecrets -}}
imagePullSecrets:
{{ toYaml .Values.global.images.ratelimit.pullSecrets }}
{{- else }}
imagePullSecrets: {{ toYaml list }}
{{- end }}
{{- end }}

{{/*
Resolve the Envoy Proxy image.
*/}}
{{- define "eg.envoyProxy.image" -}}
{{-   $imageParts := splitn "/" 2 .Values.global.images.envoyProxy.image -}}
{{/*    if global.imageRegistry is defined, it takes precedence always */}}
{{-   $registryName := default $imageParts._0 .Values.global.imageRegistry -}}
{{-   $repositoryTag := $imageParts._1 -}}
{{-   $repositoryParts := splitn ":" 2 $repositoryTag -}}
{{-   $repositoryName := $repositoryParts._0 -}}
{{-   $imageTag := default "distroless-dev" $repositoryParts._1 -}}
{{-   printf "%s/%s:%s" $registryName $repositoryName $imageTag -}}
{{- end -}}

{{/*
Resolve the Envoy Proxy image pull secrets.
*/}}
{{- define "eg.envoyProxy.image.pullSecrets" -}}
{{- if .Values.global.imagePullSecrets }}
imagePullSecrets:
{{ toYaml .Values.global.imagePullSecrets }}
{{- else if .Values.global.images.envoyProxy.pullSecrets -}}
imagePullSecrets:
{{ toYaml .Values.global.images.envoyProxy.pullSecrets }}
{{- else }}
imagePullSecrets: {{ toYaml list }}
{{- end }}
{{- end }}

{{/*
The default Envoy Gateway configuration.
*/}}
{{- define "eg.default-envoy-gateway-config" -}}
{{- if or .Values.global.images.envoyProxy.image .Values.config.envoyGateway.envoyProxy }}
{{- $envoyProxyBase := .Values.config.envoyGateway.envoyProxy | default dict }}
{{- $imageOverride := dict }}
{{- if .Values.global.images.envoyProxy.image }}
  {{- $container := dict "image" (include "eg.envoyProxy.image" .) }}
  {{- if .Values.global.images.envoyProxy.pullPolicy }}
    {{- $_ := set $container "imagePullPolicy" .Values.global.images.envoyProxy.pullPolicy }}
  {{- end }}
  {{- $deployment := dict "container" $container }}
  {{- if or .Values.global.imagePullSecrets .Values.global.images.envoyProxy.pullSecrets }}
    {{- $pullSecretsYaml := include "eg.envoyProxy.image.pullSecrets" . }}
    {{- $pullSecrets := dict "imagePullSecrets" ($pullSecretsYaml | fromYaml).imagePullSecrets }}
    {{- $_ := set $deployment "pod" $pullSecrets }}
  {{- end }}
  {{- $kubernetes := dict "envoyDeployment" $deployment }}
  {{- $provider := dict "type" "Kubernetes" "kubernetes" $kubernetes }}
  {{- $imageOverride = dict "provider" $provider }}
{{- end }}
{{- $merged := mustMergeOverwrite (dict) $envoyProxyBase $imageOverride }}
envoyProxy:
{{ toYaml $merged | indent 2 }}
{{- end }}
provider:
  type: Kubernetes
  kubernetes:
    rateLimitDeployment:
      container:
        image: {{ include "eg.ratelimit.image" . }}
      {{- if (or .Values.global.imagePullSecrets .Values.global.images.ratelimit.pullSecrets) }}
      pod:
        {{- include "eg.ratelimit.image.pullSecrets" . | nindent 8 }}
      {{- end }}
      {{- with .Values.global.images.ratelimit.pullPolicy }}
      patch:
        type: StrategicMerge
        value:
          spec:
            template:
              spec:
                containers:
                - name: envoy-ratelimit
                  imagePullPolicy: {{ . }}
      {{- end }}
    shutdownManager:
      image: {{ include "eg.image" . }}
{{- with .Values.config.envoyGateway.extensionApis }}
extensionApis:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- if not .Values.topologyInjector.enabled }}
proxyTopologyInjector:
  disabled: true
{{- end }}
{{- end }}
