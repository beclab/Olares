package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var RedisConf = template.Must(template.New("redis.conf").Parse(
	dedent.Dedent(`protected-mode no
bind {{ .LocalIP }}
port 6379
daemonize no
supervised no
pidfile {{ .RootPath }}/run/redis.pid
logfile {{ .RootPath }}/log/redis-server.log
save 900 1
save 600 50
save 300 100
save 180 300
save 60 1000
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir {{ .RootPath }}/data
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 32mb
requirepass {{ .Password }}
	`),
))

var RedisService = template.Must(template.New("redis.service").Parse(
	dedent.Dedent(`[Unit]
Description=Redis
Documentation=https://redis.io/
Wants=network-online.target
After=network-online.target
AssertFileIsExecutable={{ .RedisBinPath }}
StartLimitIntervalSec=0

[Service]
WorkingDirectory={{ .RootPath }}

User=root
Group=root

EnvironmentFile=
ExecStartPre=/bin/sh -c "test -f /sys/kernel/mm/transparent_hugepage/enabled && /bin/echo never > /sys/kernel/mm/transparent_hugepage/enabled; test -f {{ .RootPath }}/data/appendonly.aof && (echo y | /usr/local/bin/redis-check-aof --fix {{ .RootPath }}/data/appendonly.aof); true"
ExecStart={{ .RedisBinPath }} {{ .RedisConfPath }}

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
