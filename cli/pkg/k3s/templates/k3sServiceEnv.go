/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

// K3sServiceEnv defines the template of kubelet's Env for the kubelet's systemd service.
var K3sServiceEnv = template.Must(template.New("k3s.service.env").Parse(
	dedent.Dedent(`# Note: This dropin only works with k3s
{{ if .IsMaster }}
K3S_DATASTORE_ENDPOINT={{ .DataStoreEndPoint }}
{{- if .DataStoreCaFile }}
K3S_DATASTORE_CAFILE={{ .DataStoreCaFile }}
{{- end }}
{{- if .DataStoreCertFile }}
K3S_DATASTORE_CERTFILE={{ .DataStoreCertFile }}
{{- end }}
{{- if .DataStoreKeyFile }}
K3S_DATASTORE_KEYFILE={{ .DataStoreKeyFile }}
{{- end }}
K3S_KUBECONFIG_MODE=644
{{ end }}
{{ if .Token }}
K3S_TOKEN={{ .Token }}
{{ end }}
GODEBUG=netdns=go

    `)))
