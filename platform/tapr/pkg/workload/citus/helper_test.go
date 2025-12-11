package citus

import "testing"

func TestGenPassword(t *testing.T) {
	pwd, err := genPassword()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	t.Log(pwd)
}
