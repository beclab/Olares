package rag

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

type fakeDoer struct {
	method     string
	path       string
	body       interface{}
	result     string
	requestErr error
}

func (d *fakeDoer) DoJSON(_ context.Context, method, path string, body, out interface{}) error {
	d.method, d.path, d.body = method, path, body
	if d.requestErr != nil {
		return d.requestErr
	}
	return json.Unmarshal([]byte(d.result), out)
}

func TestCallPreservesRAGEnvelopeData(t *testing.T) {
	d := &fakeDoer{result: `{"code":0,"message":null,"data":["drive/Home/Documents/"]}`}
	response, err := call(context.Background(), d, "GET", "/api/rag/directories", nil)
	if err != nil {
		t.Fatal(err)
	}
	if d.method != "GET" || d.path != "/api/rag/directories" {
		t.Fatalf("unexpected request %s %s", d.method, d.path)
	}
	if string(response.Data) != `["drive/Home/Documents/"]` {
		t.Fatalf("data = %s", response.Data)
	}
}

func TestCallReturnsServiceCodeAndMessage(t *testing.T) {
	d := &fakeDoer{result: `{"code":-7,"message":"another directory operation is active","data":null}`}
	_, err := call(context.Background(), d, "DELETE", "/api/rag/directories", map[string]interface{}{"roots": []string{"drive/Home/"}})
	if err == nil || !strings.Contains(err.Error(), "another directory operation") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCallAllowsDirectoryEmptyStateOnlyWhenRequested(t *testing.T) {
	d := &fakeDoer{result: `{"code":-2,"message":"no RAG monitor setting","data":[]}`}
	if _, err := call(context.Background(), d, "GET", "/api/rag/directories", nil); err == nil {
		t.Fatal("generic call unexpectedly accepted NotFound")
	}
	response, err := callAllowCodes(context.Background(), d, "GET", "/api/rag/directories", nil, -2)
	if err != nil {
		t.Fatal(err)
	}
	if string(response.Data) != `[]` {
		t.Fatalf("data = %s", response.Data)
	}
}

func TestValidateRootNormalizesAndRejectsTrustedFields(t *testing.T) {
	root, err := validateRoot("/drive/Home/Documents")
	if err != nil {
		t.Fatal(err)
	}
	if root != "drive/Home/Documents/" {
		t.Fatalf("root = %q", root)
	}
	for _, invalid := range []string{"files_v2:bob:drive/Home/", "drive/Home/../Other/", `drive\\Home`} {
		if _, err := validateRoot(invalid); err == nil {
			t.Fatalf("accepted invalid root %q", invalid)
		}
	}
}

func TestRAGCommandTree(t *testing.T) {
	root := NewRAGCommand(nil)
	want := []string{"dirs", "failed", "pause", "rebuild", "resume", "retry", "status"}
	var got []string
	for _, command := range root.Commands() {
		got = append(got, command.Name())
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %v, want %v", got, want)
	}
	dirs, _, err := root.Find([]string{"dirs"})
	if err != nil {
		t.Fatal(err)
	}
	var dirCommands []string
	for _, command := range dirs.Commands() {
		dirCommands = append(dirCommands, command.Name())
	}
	if !reflect.DeepEqual(dirCommands, []string{"add", "list", "rm"}) {
		t.Fatalf("dir commands = %v", dirCommands)
	}
}
