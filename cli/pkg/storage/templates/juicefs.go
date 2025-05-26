package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var JuicefsService = template.Must(template.New("juicefs.service").Parse(
	dedent.Dedent(`[Unit]
Description=JuicefsMount
Documentation=https://juicefs.com/docs/zh/community/introduction/
Wants=redis-online.target
After=redis-online.target
AssertFileIsExecutable={{ .JuiceFsBinPath }}
StartLimitIntervalSec=0

[Service]
WorkingDirectory=/usr/local

EnvironmentFile=
ExecStart={{ .JuiceFsBinPath }} mount -o writeback_cache --entry-cache 300 --attr-cache 300 --cache-dir {{ .JuiceFsCachePath }} {{ .JuiceFsMetaDb }} {{ .JuiceFsMountPoint }}

# Let systemd restart this service always
Restart=always

# Specifies the maximum file descriptor number that can be opened by this process
LimitNOFILE=65536

# Specifies the maximum number of threads this process can create
TasksMax=infinity

# Disable timeout logic and wait until process is stopped
TimeoutStopSec=infinity
SendSIGKILL=no

[Install]
WantedBy=multi-user.target
	`),
))
