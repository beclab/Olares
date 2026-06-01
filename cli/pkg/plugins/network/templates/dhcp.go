package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var CniDhcpService = template.Must(template.New("cni-dhcp.service").Parse(
	dedent.Dedent(`[Unit]
Description=CNI DHCP IPAM daemon (for Multus/macvlan DHCP)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/opt/cni/bin/dhcp daemon
Restart=on-failure
RestartSec=2
User=root

[Install]
WantedBy=multi-user.target
	`),
))
