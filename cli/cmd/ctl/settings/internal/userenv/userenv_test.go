package userenv

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/bflenvelope"
)

func TestBodySortsByName(t *testing.T) {
	got := Body(map[string]string{"ZED": "3", "ALPHA": "1", "MID": "2"})

	wantOrder := []string{"ALPHA", "MID", "ZED"}
	if len(got) != len(wantOrder) {
		t.Fatalf("body has %d entries; want %d: %+v", len(got), len(wantOrder), got)
	}
	for i, name := range wantOrder {
		if got[i].EnvName != name {
			t.Errorf("entry %d = %q; want %q", i, got[i].EnvName, name)
		}
	}
}

// An empty value is a legitimate write (clearing a variable), so it must
// survive into the body rather than being dropped as "unset".
func TestBodyKeepsEmptyValues(t *testing.T) {
	got := Body(map[string]string{"FOO": ""})
	if len(got) != 1 || got[0].EnvName != "FOO" || got[0].Value != "" {
		t.Errorf("body = %+v; want a single cleared FOO", got)
	}
}

type recordedCall struct {
	method string
	path   string
	body   interface{}
}

type fakeDoer struct {
	list  []Entry
	calls []recordedCall
}

func (d *fakeDoer) DoJSON(_ context.Context, method, path string, body, out interface{}) error {
	d.calls = append(d.calls, recordedCall{method: method, path: path, body: body})
	env, ok := out.(*bflenvelope.Envelope)
	if !ok {
		return nil
	}
	env.Code = 200
	if method == "GET" {
		data, err := json.Marshal(d.list)
		if err != nil {
			return err
		}
		env.Data = data
	}
	return nil
}

// The upstream upserts what it is given, and 400s the whole batch if any
// entry in the body is non-editable. Sending the untouched variables back
// would turn an unrelated read-only entry into a failed write, so the
// write must be a single PUT carrying only the named variables.
func TestSetValuesSendsOnlyTheNamedVariables(t *testing.T) {
	d := &fakeDoer{list: []Entry{
		{EnvName: "OLARES_USER_LANGUAGE", Value: "en-US"},
		{EnvName: ThemeEnvName, Value: "light"},
		{EnvName: "OLARES_USER_PASSWORD", Value: "secret"},
	}}

	if err := SetValues(context.Background(), d, UserEnvsPath, map[string]string{ThemeEnvName: "dark"}); err != nil {
		t.Fatalf("SetValues errored: %v", err)
	}

	if len(d.calls) != 1 {
		t.Fatalf("made %d calls; want a single PUT with no preceding read: %+v", len(d.calls), d.calls)
	}
	if d.calls[0].method != "PUT" || d.calls[0].path != UserEnvsPath {
		t.Errorf("call = %s %s; want PUT %s", d.calls[0].method, d.calls[0].path, UserEnvsPath)
	}

	raw, err := json.Marshal(d.calls[0].body)
	if err != nil {
		t.Fatalf("marshal PUT body: %v", err)
	}
	want := `[{"envName":"OLARES_USER_THEME","value":"dark"}]`
	if string(raw) != want {
		t.Errorf("PUT body = %s; want %s", raw, want)
	}
}

func TestSetValuesRefusesAnEmptyUpdateSet(t *testing.T) {
	d := &fakeDoer{}
	err := SetValues(context.Background(), d, UserEnvsPath, nil)
	if err == nil || !strings.Contains(err.Error(), "no environment variable updates") {
		t.Fatalf("SetValues(nil) = %v; want a refusal before any request", err)
	}
	if len(d.calls) != 0 {
		t.Errorf("made %d calls; want none", len(d.calls))
	}
}

func TestListDecodesTheVector(t *testing.T) {
	d := &fakeDoer{list: []Entry{{EnvName: "A", Value: "1"}, {EnvName: "B", Value: "2"}}}
	got, err := List(context.Background(), d, UserEnvsPath)
	if err != nil {
		t.Fatalf("List errored: %v", err)
	}
	if len(got) != 2 || got[0].EnvName != "A" || got[1].Value != "2" {
		t.Errorf("List = %+v; want the upstream vector verbatim", got)
	}
}

func TestListSurfacesAnUpstreamErrorCode(t *testing.T) {
	d := &failingDoer{code: 500, message: "boom"}
	_, err := List(context.Background(), d, UserEnvsPath)
	if err == nil {
		t.Fatal("List returned no error on upstream code 500")
	}
	for _, sub := range []string{"GET " + UserEnvsPath, "500", "boom"} {
		if !strings.Contains(err.Error(), sub) {
			t.Errorf("error %q does not contain %q", err, sub)
		}
	}
}

type failingDoer struct {
	code    int
	message string
}

func (d *failingDoer) DoJSON(_ context.Context, _, _ string, _, out interface{}) error {
	if env, ok := out.(*bflenvelope.Envelope); ok {
		env.Code = d.code
		env.Message = bflenvelope.Message{Text: d.message}
	}
	return nil
}
