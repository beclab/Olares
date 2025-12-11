package nats

import (
	"fmt"
	"testing"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
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
