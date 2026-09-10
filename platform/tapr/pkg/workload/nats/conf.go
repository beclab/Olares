package nats

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/mitchellh/mapstructure"
	load "github.com/nats-io/nats-server/v2/conf"
	"k8s.io/klog/v2"
)

const DefaultNatsConf = `{
  "http_port": 8222,
  "jetstream": {
    "max_file_store": 10102410241024,
    "max_memory_store": 0,
    "store_dir": "/data"
  },
  "accounts": {
    "terminus": {
      "jetstream": enabled,
      "users": [
        {
          "user": "admin",
          "password": $ADMIN_PASSWORD,
          "permissions": {
            "publish": {
              "allow": [">"]
            },
            "subscribe": {
              "allow": [">"]
            }
          }
        }
      ]
    }
  },
  "port": 4222,
  "pid_file": "/var/run/nats/nats.pid",
  "server_name": "nats-0"
}
`

const tmpl = `{
  "http_port": {{.HTTPPort}},
  "jetstream": {
    "max_file_store": {{.Jetstream.MaxFileStore}},
    "max_memory_store": {{.Jetstream.MaxMemoryStore}},
    "store_dir": "{{.Jetstream.StoreDir}}"
  },
  "accounts": {
    "terminus": {
      "jetstream": enabled,
      "users": [
        {{- range $index, $user := .Accounts.Terminus.Users }}
        {{- if $index}},{{ end }}
        {
          "user": "{{ $user.Username }}",
          "password": {{ if eq $user.Username "admin" }}$ADMIN_PASSWORD{{ else }}{{ $user.Password | quoteOrNot }}{{ end }},
          "permissions": {
            "publish": {
              "allow": [{{ range $i, $allow := $user.Permissions.Publish.Allow }}{{ if $i }}, {{ end }}"{{ $allow }}"{{ end }}]
            },
            "subscribe": {
              "allow": [{{ range $i, $allow := $user.Permissions.Subscribe.Allow }}{{ if $i }}, {{ end }}"{{ $allow }}"{{ end }}]
            }
          }
        }
        {{- end }}
      ]
    }
  },
  "port": {{ .Port }},
  "pid_file": "{{ .PidFile }}",
  "server_name": "{{ .ServerName }}"
}
`

func quoteOrNot(s string) string {
	if strings.HasPrefix(s, "$2a") {
		return s
	}
	if len(s) > 0 && s[0] == '$' {
		return s
	}
	return `"` + s + `"`
}

func renderConfigFile(config *Config) ([]byte, error) {
	funcMap := template.FuncMap{
		"quoteOrNot": quoteOrNot,
	}
	// klog.Infof("renderConfigFile: %##v\n", config)
	var buf bytes.Buffer
	tpl := template.Must(template.New("config").Funcs(funcMap).Parse(tmpl))
	err := tpl.Execute(&buf, config)
	if err != nil {
		return buf.Bytes(), err
	}
	return buf.Bytes(), nil
}

func Parse(data string) (*Config, error) {
	m, err := load.Parse(data)
	if err != nil {
		return nil, err
	}
	return decodeConfig(m)
}

func ParseFile(fp string) (*Config, error) {
	m, err := load.ParseFile(fp)
	if err != nil {
		return nil, err
	}
	return decodeConfig(m)
}

func decodeConfig(m map[string]interface{}) (*Config, error) {
	var config Config
	err := mapstructure.Decode(m, &config)
	if err != nil {
		klog.Infof("mapstructure decode: err=%v", err)
		return nil, err
	}
	return &config, nil
}
