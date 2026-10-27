package nats

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	raw := `{
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
          "password": "secret",
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
}`
	c, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if c.HTTPPort != 8222 {
		t.Fatalf("http_port: got %d", c.HTTPPort)
	}
	if len(c.Accounts.Terminus.Users) != 1 || c.Accounts.Terminus.Users[0].Username != "admin" {
		t.Fatalf("users: %#v", c.Accounts.Terminus.Users)
	}
}

func TestRenderConfigFile(t *testing.T) {
	config := Config{
		HTTPPort: 8222,
		Jetstream: Jetstream{
			MaxFileStore:   13124323,
			MaxMemoryStore: 34243243,
			StoreDir:       "/data",
		},
		Port:       4222,
		PidFile:    "/var/run/nats/nats.pid",
		ServerName: "$ServerName",
		Accounts: Accounts{
			Terminus: Terminus{
				Jetstream: "enabled",
				Users: []User{
					{
						Username: "admin",
						Password: "$ADMIN_PASSWORD",
						Permissions: Permissions{
							Publish: Publish{
								Allow: []string{">"},
							},
							Subscribe: Subscribe{
								Allow: []string{">"},
							},
						},
					},
					{
						Username: "user",
						Password: "hello",
						Permissions: Permissions{
							Publish: Publish{
								Allow: []string{">"},
							},
							Subscribe: Subscribe{
								Allow: []string{">"},
							},
						},
					},
				},
			},
		},
	}
	data, err := renderConfigFile(&config)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}
