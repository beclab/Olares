{{/*
All namespaced resources for Envoy Gateway RBAC.
*/}}
{{- define "eg.rbac.namespaced" -}}
- {{ include "eg.rbac.namespaced.basic" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.apps" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.discovery" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.gateway.envoyproxy" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.gateway.envoyproxy.status" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.gateway.networking" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.gateway.networking.status" . | nindent 2 | trim }}
{{- if .Values.topologyInjector.enabled }}
- {{ include "eg.rbac.namespaced.topologyinjector" . | nindent 2 | trim }}
{{- end }}
{{- end }}

{{/*
All cluster scoped resources for Envoy Gateway RBAC.
*/}}
{{- define "eg.rbac.cluster" -}}
- {{ include "eg.rbac.cluster.basic" . | nindent 2 | trim }}
- {{ include "eg.rbac.cluster.gateway.networking" . | nindent 2 | trim }}
- {{ include "eg.rbac.cluster.gateway.networking.status" . | nindent 2 | trim }}
- {{ include "eg.rbac.cluster.multiclusterservices" . | nindent 2 | trim }}
{{- end }}

{{/*
Namespaced
*/}}

{{- define "eg.rbac.namespaced.basic" -}}
apiGroups:
- ""
resources:
- configmaps
- secrets
- services
verbs:
- get
- list
- watch
{{- end }}

{{- define "eg.rbac.namespaced.topologyinjector" -}}
apiGroups:
- ""
resources:
- pods
- pods/binding
verbs:
- get
- list
- patch
- update
- watch
{{- end }}

{{- define "eg.rbac.namespaced.apps" -}}
apiGroups:
- apps
resources:
- deployments
- daemonsets
verbs:
- get
- list
- watch
{{- end }}

{{- define "eg.rbac.namespaced.discovery" -}}
apiGroups:
- discovery.k8s.io
resources:
- endpointslices
verbs:
- get
- list
- watch
{{- end }}

{{- define "eg.rbac.namespaced.gateway.envoyproxy" -}}
apiGroups:
- gateway.envoyproxy.io
resources:
- envoyproxies
- envoypatchpolicies
- clienttrafficpolicies
- backendtrafficpolicies
- securitypolicies
- envoyextensionpolicies
- backends
- httproutefilters
verbs:
- get
- list
- watch
{{- end }}

{{- define "eg.rbac.namespaced.gateway.envoyproxy.status" -}}
apiGroups:
- gateway.envoyproxy.io
resources:
- envoyproxies/status
- envoypatchpolicies/status
- clienttrafficpolicies/status
- backendtrafficpolicies/status
- securitypolicies/status
- envoyextensionpolicies/status
- backends/status
verbs:
- update
{{- end }}

{{- define "eg.rbac.namespaced.gateway.networking" -}}
apiGroups:
- gateway.networking.k8s.io
resources:
- gateways
- listenersets
- grpcroutes
- httproutes
- referencegrants
- tcproutes
- tlsroutes
- udproutes
- backendtlspolicies
verbs:
- get
- list
- watch
{{- end }}

{{- define "eg.rbac.namespaced.gateway.networking.status" -}}
apiGroups:
- gateway.networking.k8s.io
resources:
- gateways/status
- listenersets/status
- grpcroutes/status
- httproutes/status
- tcproutes/status
- tlsroutes/status
- udproutes/status
- backendtlspolicies/status
verbs:
- update
{{- end }}

{{- define "eg.rbac.namespaced.gateway.networking.experimental" -}}
{{- end }}

{{- define "eg.rbac.namespaced.gateway.networking.experimental.status" -}}
{{- end }}

{{/*
Cluster scope
*/}}

{{- define "eg.rbac.cluster.basic" -}}
apiGroups:
- ""
resources:
- nodes
- namespaces
verbs:
- get
- list
- watch
{{- end }}

{{- define "eg.rbac.cluster.gateway.networking" -}}
apiGroups:
- gateway.networking.k8s.io
resources:
- gatewayclasses
verbs:
- get
- list
- patch
- update
- watch
{{- end }}


{{- define "eg.rbac.cluster.multiclusterservices" -}}
apiGroups:
- multicluster.x-k8s.io
resources:
- serviceimports
verbs:
- get
- list
- watch
{{- end }}

{{- define "eg.rbac.cluster.gateway.networking.status" -}}
apiGroups:
- gateway.networking.k8s.io
resources:
- gatewayclasses/status
verbs:
- update
{{- end }}

{{- define "eg.rbac.infra.basic" -}}
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  - services
  - configmaps
  verbs:
  - create
  - get
  - list
  - delete
  - deletecollection
  - patch
  - watch
- apiGroups:
  - apps
  resources:
  - deployments
  - daemonsets
  verbs:
  - create
  - get
  - list
  - delete
  - deletecollection
  - patch
  - watch
- apiGroups:
  - autoscaling
  - policy
  resources:
  - horizontalpodautoscalers
  - poddisruptionbudgets
  verbs:
  - create
  - get
  - list
  - delete
  - deletecollection
  - patch
  - watch
- apiGroups:
  - certificates.k8s.io
  resources:
  - clustertrustbundles
  verbs:
  - list
  - get
  - watch
{{- end }}

{{- define "eg.rbac.controllernamespace.secrets.read" -}}
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
  - watch
{{- end }}

{{- define "eg.rbac.infra.tokenreview" -}}
- apiGroups:
  - authentication.k8s.io
  resources:
  - tokenreviews
  verbs:
  - create
{{- end }}
