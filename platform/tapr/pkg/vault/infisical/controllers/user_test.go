package controllers

import (
	"testing"

	"bytetrade.io/web3os/tapr/pkg/vault/infisical"
)

func TestUserOrgs(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2NDgxYzE2N2VlNTI3Yzc1ZTk2MjJmYWIiLCJpYXQiOjE2ODYyODAyOTUsImV4cCI6MTY4NzE0NDI5NX0.6nmz4RTcjQyHCcXGx0Sb5PliGkZz1nKEKsV7rjoich0"
	infisical.InfisicalAddr = "https://infisical.liuy102.snowinning.com"
	u := &userClient{}
	id, err := u.GetUserOrganizationId(token)
	if err != nil {
		t.Log(err)
		t.Fail()
	} else {
		t.Log("org id: ", id)
	}
}
