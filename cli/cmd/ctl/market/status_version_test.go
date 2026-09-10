package market

import (
	"encoding/json"
	"testing"
)

// status is where a reader looks first to answer "what is this app doing", and
// it used to answer everything except which version. Finding that out meant a
// second command against a different endpoint, which is also how the two came
// to disagree without anyone noticing.
//
// The version lives beside `status` on the state row rather than inside it (see
// AppStateLatest.Version), which is the reason it was missed.
func TestParseStatusRowsCarriesTheRowVersion(t *testing.T) {
	const state = `{
        "user_data": {
            "sources": {
                "upload": {
                    "type": "local",
                    "app_state_latest": [
                        {"version": "0.1.8", "status": {"name": "clitest", "state": "running"}},
                        {"status": {"name": "noversion", "state": "running"}}
                    ]
                }
            }
        }
    }`

	rows, err := parseStatusRows(&APIResponse{Data: json.RawMessage(state)}, "upload", false)
	if err != nil {
		t.Fatalf("parseStatusRows: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}

	byName := map[string]statusRow{}
	for _, r := range rows {
		byName[r.Name] = r
	}
	if got := byName["clitest"].Version; got != "0.1.8" {
		t.Fatalf("clitest version = %q, want 0.1.8", got)
	}
	// A row without a version must stay empty rather than borrow the catalog's
	// latest, which is the wrong answer as soon as the catalog moves ahead.
	if got := byName["noversion"].Version; got != "" {
		t.Fatalf("noversion version = %q, want empty", got)
	}
}
