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

// ContainerdConfig renders /etc/containerd/config.toml in containerd config
// version 3 (containerd 2.x) format. It is intentionally minimal.
//
// Registry mirrors and auth are NOT inlined here: config version 2's
// registry.mirrors / registry.configs are DEPRECATED and ignored by containerd
// 2.x whenever config_path is set (and it is an error to specify both). Instead:
//
//   - Mirrors live under config_path (/etc/containerd/certs.d/<host>/hosts.toml),
//     seeded by the CLI at install (docker.io -> configured dockerhub mirror,
//     consistent across all nodes) and reconciled by olaresd at runtime. Because
//     hosts.toml is read per-pull, mirror changes need no containerd restart.
//   - Drop-in files under /etc/containerd/conf.d/*.toml are imported. This is
//     where nvidia-container-toolkit writes its runtime settings
//     (conf.d/99-nvidia.toml) via `nvidia-ctk runtime configure --drop-in-config`.
//     Those only touch the runtime section, so Olares-managed registry config in
//     certs.d always survives an nvidia configure, and nodes without a GPU (which
//     never run nvidia configure) get an identical base config.
var ContainerdConfig = template.Must(template.New("config.toml").Parse(
	dedent.Dedent(`version = 3
{{- if .DataRoot }}
root = {{ .DataRoot }}
{{- else }}
root = "/var/lib/containerd"
{{- end }}

imports = ["/etc/containerd/conf.d/*.toml"]

[plugins]

  [plugins.'io.containerd.cri.v1.images']
    snapshotter = "{{ .FsType }}"

    [plugins.'io.containerd.cri.v1.images'.pinned_images]
      sandbox = "{{ .SandBoxImage }}"

    [plugins.'io.containerd.cri.v1.images'.registry]
      config_path = "/etc/containerd/certs.d"

  [plugins.'io.containerd.cri.v1.runtime'.containerd]
    default_runtime_name = "runc"

    [plugins.'io.containerd.cri.v1.runtime'.containerd.runtimes.runc]
      runtime_type = 'io.containerd.runc.v2'
      sandboxer = 'podsandbox'

      [plugins.'io.containerd.cri.v1.runtime'.containerd.runtimes.runc.options]
        SystemdCgroup = true

  [plugins."io.containerd.snapshotter.v1.zfs"]
    root_path = "{{ .ZfsRootPath }}"
`)))

// ContainerdRegistryHosts renders /etc/containerd/certs.d/docker.io/hosts.toml,
// the initial docker.io pull-through configuration written at install/prepare
// time. It points docker.io at the configured dockerhub mirror(s) (the same on
// every node) and falls back to the canonical upstream. olaresd may reconcile
// this file at runtime (e.g. prepend a higher-priority mirror).
//
// Note: `server` is the canonical fallback upstream; each `[host."..."]` is a
// mirror tried before it, in order. For docker.io the upstream registry host is
// registry-1.docker.io (NOT docker.io, which does not serve the /v2 API);
// containerd only auto-maps docker.io -> registry-1.docker.io when `server` is
// empty, so we must spell it out here since we always write `server`.
var ContainerdRegistryHosts = template.Must(template.New("hosts.toml").Parse(
	dedent.Dedent(`server = "https://registry-1.docker.io"
{{- range .Mirrors }}

[host."{{ . }}"]
  capabilities = ["pull", "resolve"]
{{- end }}
`)))
