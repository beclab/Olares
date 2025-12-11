package controllers

import (
	"testing"

	"bytetrade.io/web3os/tapr/pkg/vault/infisical"
	"github.com/google/uuid"
	"k8s.io/utils/pointer"
)

func TestCreateSecret(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2NDgxYzE2N2VlNTI3Yzc1ZTk2MjJmYWIiLCJpYXQiOjE2ODYyODAyOTUsImV4cCI6MTY4NzE0NDI5NX0.6nmz4RTcjQyHCcXGx0Sb5PliGkZz1nKEKsV7rjoich0"
	infisical.InfisicalAddr = "https://infisical.liuy102.snowinning.com"

	user := &infisical.UserEncryptionKeysPG{
		ID:                  pointer.String(uuid.New().String()),
		EncryptionVersion:   1,
		EncryptedPrivateKey: "ITMdDXtLoxib4+53U/qzvIV/T/UalRwimogFCXv/UsulzEoiKM+aK2aqOb0=",
		IV:                  "9fp0dZHI+UuHeKkWMDvD6w==",
		PublicKey:           "cf44BhkybbBfsE0fZHe2jvqtCj6KLXvSq4hVjV0svzk=",
		Salt:                "d8099dc70958090346910fb9639262b83cf526fc9b4555a171b36a9e1bcd0240",
		Tag:                 "bQ/UTghqcQHRoSMpLQD33g==",
	}

	s := &secretClient{}

	err := s.CreateSecretInWorkspace(user, token, "64882ef65819cc5941c7d3d5", "748b8bf6769e72c65d7a74ecf3eb0ada", "test-client", "secret value", "dev")
	if err != nil {
		t.Log("create error: ", err)
		t.Fail()
		return
	}

}

func TestRetrieveSecret(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2NDgxYzE2N2VlNTI3Yzc1ZTk2MjJmYWIiLCJpYXQiOjE2ODYyODAyOTUsImV4cCI6MTY4NzE0NDI5NX0.6nmz4RTcjQyHCcXGx0Sb5PliGkZz1nKEKsV7rjoich0"
	infisical.InfisicalAddr = "https://infisical.liuy102.snowinning.com"

	s := &secretClient{}

	name, value, err := s.GetSecretInWorkspace(token, "64882ef65819cc5941c7d3d5", "748b8bf6769e72c65d7a74ecf3eb0ada", "dev", "test-client")
	if err != nil {
		t.Log("get error, ", err)
		t.Fail()
		return
	}

	t.Log("get secret", name, ",", value)
}
