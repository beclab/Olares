package controllers

import "testing"

func TestDecode(t *testing.T) {
	op := DecodeOps("CreateSecret?workspace=testws")

	t.Log(op)
}
