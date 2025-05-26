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

// KubeletService defines the template of kubelete service for systemd.
var KubeletService = template.Must(template.New("kubelet.service").Parse(
	dedent.Dedent(`[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=http://kubernetes.io/docs/
{{ if .JuiceFSPreCheckEnabled }}
After={{ .JuiceFSServiceUnit }}
{{ end }}
StartLimitIntervalSec=0

[Service]
CPUAccounting=true
MemoryAccounting=true
{{ if .JuiceFSPreCheckEnabled }}
ExecStartPre={{ .JuiceFSBinPath }} summary {{ .JuiceFSMountPoint }}
{{ end }}
ExecStart=/usr/local/bin/kubelet
Restart=always
StartLimitInterval=0
RestartSec=10

[Install]
WantedBy=multi-user.target
    `)))
