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
	// ETCDService defines the template of etcd's service for systemd.
	ETCDService = template.Must(template.New("etcd.service").Parse(
		dedent.Dedent(`[Unit]
Description=etcd
After=network.target
StartLimitIntervalSec=0

[Service]
User=root
Type=notify
EnvironmentFile=/etc/etcd.env
ExecStart=/usr/local/bin/etcd
NotifyAccess=all
RestartSec=10s
LimitNOFILE=40000
Restart=always

[Install]
WantedBy=multi-user.target
    `)))
)
