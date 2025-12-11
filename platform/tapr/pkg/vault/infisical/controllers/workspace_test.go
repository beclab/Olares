package controllers

import (
	"crypto/rand"
	"io"
	"testing"

	"bytetrade.io/web3os/tapr/pkg/utils"
	"bytetrade.io/web3os/tapr/pkg/vault/infisical"
	"github.com/google/uuid"
)

func TestGetWorkspace(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2NDgxYzE2N2VlNTI3Yzc1ZTk2MjJmYWIiLCJpYXQiOjE2ODYyODAyOTUsImV4cCI6MTY4NzE0NDI5NX0.6nmz4RTcjQyHCcXGx0Sb5PliGkZz1nKEKsV7rjoich0"
	infisical.InfisicalAddr = "https://infisical.liuy102.snowinning.com"

	w := &workspaceClient{}

	id, err := w.GetWorkspace(token, "6481c2e3971ab38ac64cd693", "Message")
	if err != nil {
		t.Log(err)
		t.Fail()
	} else {
		t.Log("workspace id: ", id)
	}
}

func TestCreateWorkspace(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2NDgxYzE2N2VlNTI3Yzc1ZTk2MjJmYWIiLCJpYXQiOjE2ODYyODAyOTUsImV4cCI6MTY4NzE0NDI5NX0.6nmz4RTcjQyHCcXGx0Sb5PliGkZz1nKEKsV7rjoich0"
	infisical.InfisicalAddr = "https://infisical.liuy102.snowinning.com"

	w := &workspaceClient{}

	user := &infisical.UserEncryptionKeysPG{
		UserID:              uuid.New().String(),
		EncryptionVersion:   1,
		EncryptedPrivateKey: "ITMdDXtLoxib4+53U/qzvIV/T/UalRwimogFCXv/UsulzEoiKM+aK2aqOb0=",
		IV:                  "9fp0dZHI+UuHeKkWMDvD6w==",
		PublicKey:           "cf44BhkybbBfsE0fZHe2jvqtCj6KLXvSq4hVjV0svzk=",
		Salt:                "d8099dc70958090346910fb9639262b83cf526fc9b4555a171b36a9e1bcd0240",
		Tag:                 "bQ/UTghqcQHRoSMpLQD33g==",
	}

	randName := make([]byte, 6)
	if _, err := io.ReadFull(rand.Reader, randName); err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	wname := "create " + utils.Hex(randName)

	u := &userClient{}
	pk, err := u.GetUserPrivateKey(user, "testInfisical1")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	wid, err := w.CreateWorkspace(user, token, "6481c2e3971ab38ac64cd693", wname, pk)
	if err != nil {
		t.Log(err)
		t.Fail()
	} else {
		t.Log("workspace id: ", wid, " name: ", wname)
	}
}
