/*
Copyright 2022 The KubeSphere Authors.
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

var (
	// RegistryServiceTempl defines the template of registry service for systemd.
	RegistryServiceTempl = template.Must(template.New("registryService").Parse(
		dedent.Dedent(`[Unit]
Description=v2 Registry server for Container
After=network.target
StartLimitIntervalSec=0
[Service]
Type=simple
ExecStart=/usr/local/bin/registry serve /etc/kubekey/registry/config.yaml
Restart=on-failure
[Install]
WantedBy=multi-user.target
    `)))

	// RegistryConfigTempl defines the template of registry's configuration file.
	RegistryConfigTempl = template.Must(template.New("registryConfig").Parse(
		dedent.Dedent(`version: 0.1
log:
  fields:
    service: registry
storage:
    cache:
        layerinfo: inmemory
    filesystem:
        rootdirectory: /mnt/registry
http:
    addr: :443
    tls:
      certificate: /etc/ssl/registry/ssl/{{ .Certificate }}
      key: /etc/ssl/registry/ssl/{{ .Key }}
    `)))
)
