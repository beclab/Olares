package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

func TestProfileListJSONReportsEveryProfile(t *testing.T) {
	t.Setenv("OLARES_CLI_HOME", t.TempDir())
	if err := cliconfig.SaveMultiProfileConfig(managedConfig()); err != nil {
		t.Fatalf("seed config: %v", err)
	}
	var out bytes.Buffer
	if err := runList(context.Background(), nil, false, true, &out); err != nil {
		t.Fatalf("runList: %v", err)
	}
	var rows []profileRow
	if err := json.Unmarshal(out.Bytes(), &rows); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out.String())
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[0].OlaresID != managedID || !rows[0].Current || rows[0].Source != "platform(lares)" || rows[0].Status != "pending" {
		t.Fatalf("managed row = %+v", rows[0])
	}
	if rows[1].Current || rows[1].Source != "local" || rows[1].Status != "never" {
		t.Fatalf("local row = %+v", rows[1])
	}
}

func TestProfileListJSONWithoutProfilesIsAnEmptyArray(t *testing.T) {
	t.Setenv("OLARES_CLI_HOME", t.TempDir())
	var out bytes.Buffer
	if err := runList(context.Background(), nil, false, true, &out); err != nil {
		t.Fatalf("runList: %v", err)
	}
	if got := bytes.TrimSpace(out.Bytes()); string(got) != "[]" {
		t.Fatalf("got %q, want []", got)
	}
}
