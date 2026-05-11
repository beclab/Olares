package users

import (
	"bytes"
	"testing"
)

func TestDecodeUsersMutateBodyBareName(t *testing.T) {
	var got struct {
		Name string `json:"name"`
	}
	if err := decodeUsersMutateBody([]byte(`{"name":"alice"}`), &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "alice" {
		t.Fatalf("name %q", got.Name)
	}
}

func TestDecodeUsersMutateBodyEnvelope200(t *testing.T) {
	var got map[string]string
	raw := bytes.TrimSpace([]byte(`{"code":200,"data":{"name":"bob"}}`))
	if err := decodeUsersMutateBody(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got["name"] != "bob" {
		t.Fatalf("got %+v", got)
	}
}

func TestDecodeUsersMutateBodyErrorCode(t *testing.T) {
	var junk struct{}
	err := decodeUsersMutateBody([]byte(`{"code":400,"message":"no good"}`), &junk)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "no good" {
		t.Fatalf("err %q", err.Error())
	}
}
