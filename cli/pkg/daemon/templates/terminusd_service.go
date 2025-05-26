package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var (
	// TerminusdService defines the template of terminusd's service for systemd.
	TerminusdService = template.Must(template.New("olaresd.service").Parse(
		dedent.Dedent(`[Unit]
Description=olaresd
After=network.target
StartLimitIntervalSec=0

[Service]
User=root
EnvironmentFile=/etc/systemd/system/olaresd.service.env
ExecStart=/usr/local/bin/olaresd
RestartSec=10s
LimitNOFILE=40000
Restart=always

[Install]
WantedBy=multi-user.target
    `)))
)
