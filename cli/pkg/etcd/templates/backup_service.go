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
	"fmt"
	"strconv"
	"text/template"

	"github.com/lithammer/dedent"
)

var (
	// BackupETCDService defines the template of backup-etcd service for systemd.
	BackupETCDService = template.Must(template.New("backup-etcd.service").Parse(
		dedent.Dedent(`[Unit]
Description=Backup ETCD
[Service]
Type=oneshot
ExecStart={{ .ScriptPath }}
    `)))

	// BackupETCDTimer defines the template of backup-etcd timer for systemd.
	BackupETCDTimer = template.Must(template.New("backup-etcd.timer").Parse(
		dedent.Dedent(`[Unit]
Description=Timer to backup ETCD
StartLimitIntervalSec=0
[Timer]
{{- if .OnCalendarStr }}
OnCalendar={{ .OnCalendarStr }}
{{- else }}
OnCalendar=*-*-* *:*/30:00
{{- end }}
Unit=backup-etcd.service
[Install]
WantedBy=multi-user.target
    `)))
)

func BackupTimeOnCalendar(period int) string {
	var onCalendar string
	if period <= 0 {
		onCalendar = "*-*-* *:00/30:00"
	} else if period < 60 {
		onCalendar = fmt.Sprintf("*-*-* *:00/%d:00", period)
	} else if period >= 60 && period < 1440 {
		hour := strconv.Itoa(period / 60)
		minute := strconv.Itoa(period % 60)
		if period%60 == 0 {
			onCalendar = fmt.Sprintf("*-*-* 00/%s:00:00", hour)
		} else {
			onCalendar = fmt.Sprintf("*-*-* 00/%s:%s:00", hour, minute)
		}
	} else {
		onCalendar = "*-*-* 00:00:00"
	}
	return onCalendar
}
