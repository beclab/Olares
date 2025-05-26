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

var (
	// K8sCertsRenewService defines the template of k8s-certs-renew service for systemd.
	K8sCertsRenewService = template.Must(template.New("k8s-certs-renew.service").Parse(
		dedent.Dedent(`[Unit]
Description=Renew K8S control plane certificates
StartLimitIntervalSec=0
[Service]
Type=oneshot
ExecStart=/usr/local/bin/kube-scripts/k8s-certs-renew.sh
    `)))

	// K8sCertsRenewTimer defines the template of k8s-certs-renew timer for systemd.
	K8sCertsRenewTimer = template.Must(template.New("k8s-certs-renew.timer").Parse(
		dedent.Dedent(`[Unit]
Description=Timer to renew K8S control plane certificates
[Timer]
OnCalendar=Mon *-*-* 03:00:00
Unit=k8s-certs-renew.service
[Install]
WantedBy=multi-user.target
    `)))
)
