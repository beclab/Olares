package nats

import (
	"context"
	"fmt"
	"testing"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateOrUpdateUser(t *testing.T) {
	testCases := []struct {
		Req      aprv1.Nats
		Expected *Config
	}{
		{
			Req: aprv1.Nats{
				User: "test1",
				Subjects: []aprv1.Subject{
					{
						Name: "subject1",
						Permission: aprv1.Permission{
							Pub: "allow",
							Sub: "allow",
						},
					},
				},
			},
			Expected: &Config{
				Accounts: Accounts{
					Terminus: Terminus{
						Jetstream: "enabled",
						Users: []User{
							{
								Username: "admin",
								Password: "password",
							},
							{
								Username: "subject1",
								Permissions: Permissions{
									Publish: Publish{
										Allow: []string{"subject1"},
									},
									Subscribe: Subscribe{
										Allow: []string{"subject1"},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		ret, err := createOrUpdateUser(&aprv1.MiddlewareRequest{
			Spec: aprv1.MiddlewareSpec{
				Nats: testCase.Req,
			},
		}, "namespace", "password", func() (*Config, error) {
			return &Config{
				Accounts: Accounts{
					Terminus: Terminus{
						Jetstream: "enabled",
						Users: []User{
							{
								Username: "admin",
								Password: "password",
							},
						},
					},
				},
			}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%#v\n", ret)

	}

}

func TestGetOriginSubjectName(t *testing.T) {
	testCases := []struct {
		originalName string
		expectedName string
	}{
		{
			"terminus.aaa-bustyleg0.aaa.subject1",
			"subject1",
		},
		{
			"terminus.aaa-bustyleg0.aaa.subject1.qqq",
			"subject1.qqq",
		},
	}
	for _, testCase := range testCases {
		if testCase.expectedName != GetOriginSubjectName(testCase.originalName) {
			t.Fatalf("expetd: %s, but got: %s", testCase.expectedName, GetOriginSubjectName(testCase.originalName))
		}
	}
}

func TestEncryptPassword(t *testing.T) {
	password := "0OHhUJAsEbxDZRsOluTOwsK1z2tLT3Ti3xZ8KjHlcRQwlff7gAs011igXxEUdJno"
	encrypted, err := encryptPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(encrypted)
}

const testNatsConf = `{
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
            "publish": { "allow": [">"] },
            "subscribe": { "allow": [">"] }
          }
        }
      ]
    }
  },
  "port": 4222,
  "pid_file": "/var/run/nats/nats.pid",
  "server_name": "nats-0"
}`

func TestEnsureConfigMap(t *testing.T) {
	t.Setenv("ADMIN_PASSWORD", "secret")
	client := fake.NewSimpleClientset()
	if err := EnsureConfigMap(context.TODO(), client); err != nil {
		t.Fatal(err)
	}
	cm, err := client.CoreV1().ConfigMaps(constants.PlatformNamespace).Get(context.TODO(), ConfigMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if cm.Name != "nats-cfg" {
		t.Fatalf("configmap name: got %s", cm.Name)
	}
	if cm.Data[ConfigMapKey] != DefaultNatsConf {
		t.Fatalf("unexpected default nats.conf:\n%s", cm.Data[ConfigMapKey])
	}
	parsed, err := Parse(cm.Data[ConfigMapKey])
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.Accounts.Terminus.Users) != 1 || parsed.Accounts.Terminus.Users[0].Username != "admin" {
		t.Fatalf("users: %#v", parsed.Accounts.Terminus.Users)
	}

	cm.Data[ConfigMapKey] = testNatsConf
	if _, err := client.CoreV1().ConfigMaps(constants.PlatformNamespace).Update(context.TODO(), cm, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := EnsureConfigMap(context.TODO(), client); err != nil {
		t.Fatal(err)
	}
	kept, err := client.CoreV1().ConfigMaps(constants.PlatformNamespace).Get(context.TODO(), ConfigMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if kept.Data[ConfigMapKey] != testNatsConf {
		t.Fatal("EnsureConfigMap overwrote existing nats.conf")
	}
}

func TestLoadAndPersistConfigMap(t *testing.T) {
	t.Setenv("ADMIN_PASSWORD", "secret")
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConfigMapName,
			Namespace: constants.PlatformNamespace,
		},
		Data: map[string]string{
			ConfigMapKey: testNatsConf,
		},
	}
	client := fake.NewSimpleClientset(cm)

	cfg, err := Parse(testNatsConf)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Accounts.Terminus.Users = append(cfg.Accounts.Terminus.Users, User{
		Username: "app-user",
		Password: "hashed",
		Permissions: Permissions{
			Publish:   Publish{Allow: []string{"app.>"}},
			Subscribe: Subscribe{Allow: []string{"app.>"}},
		},
	})
	if err := persistConfig(client, cfg); err != nil {
		t.Fatal(err)
	}

	updated, err := client.CoreV1().ConfigMaps(constants.PlatformNamespace).Get(context.TODO(), ConfigMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(updated.Data[ConfigMapKey])
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.Accounts.Terminus.Users) != 2 {
		t.Fatalf("expected 2 users in configmap, got %#v", parsed.Accounts.Terminus.Users)
	}
}
