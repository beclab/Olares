package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var SwapServiceTmpl = template.Must(template.New("olares-swap.service").Parse(
	dedent.Dedent(`[Unit]
Description=Olares Swap Configuring Service
After=local-fs.target
StartLimitIntervalSec=0

[Service]
Type=oneshot

{{- if .EnableZRAM }}
ExecStart=/usr/sbin/modprobe zram
ExecStart=-/usr/sbin/swapoff /dev/zram0
ExecStart=-/usr/sbin/zramctl -r /dev/zram0
ExecStart=/usr/sbin/zramctl -f -s {{ .ZRAMSize }}
ExecStart=/usr/sbin/mkswap /dev/zram0
ExecStart=/usr/sbin/swapon -p {{ .ZRAMSwapPriority }} /dev/zram0
{{- end }}
{{- if .Swappiness }}
ExecStart=/usr/sbin/sysctl vm.swappiness={{ .Swappiness }}
{{- end }}

{{ if .EnableZRAM }}
ExecStop=-/usr/sbin/swapoff /dev/zram0
ExecStop=-/usr/sbin/zramctl -r /dev/zram0
{{ end }}

RemainAfterExit=yes
Delegate=yes
Restart=on-failure
RestartSec=5

[Install]
WantedBy=sysinit.target
`)))
