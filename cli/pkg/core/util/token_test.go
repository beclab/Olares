package util

import (
	"fmt"
	"testing"
)

func TestToken(t *testing.T) {
	var a = "n7X2dggXApH91fnVUzgPr1Fr1vAO0Upo"
	// var b = `{"email": "admin@kubesphere.io","username": "admin","token_type": "static_token"}`

	// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFkbWluQGt1YmVzcGhlcmUuaW8iLCJ1c2VybmFtZSI6ImFkbWluIiwidG9rZW5fdHlwZSI6InN0YXRpY190b2tlbiJ9.iwsRH37tcqE8HyI_S98AEM6KUH7bVdxDasR3V8QasXI
	var data, _ = EncryptToken(a)

	fmt.Println("---data---", data)
}
