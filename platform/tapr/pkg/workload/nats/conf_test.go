package nats

import (
	"fmt"
	"testing"
)

func TestParseFile(t *testing.T) {
	c, err := ParseFile("/Users/hys/code/beclab/tapr/pkg/workload/nats/nats.conf")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(c)
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
