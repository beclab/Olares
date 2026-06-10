{{- define "app-gateway.namespace" -}}
{{- .Values.namespace | required "values.namespace is required" -}}
{{- end -}}
