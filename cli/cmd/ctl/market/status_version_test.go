package market

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
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

// `market status <app>` renders a detail block rather than the table, so the
// version has to be printed there too or the single-app reader — the one who
// asked about exactly this app — is the only one who still cannot see it.
func TestRenderStatusMatchesPrintsTheVersion(t *testing.T) {
	withVersion := captureStdout(t, func() {
		if err := renderStatusMatches(&MarketOptions{}, []statusRow{{
			Name: "clitest", Source: "upload", Version: "0.1.8", State: "running",
		}}); err != nil {
			t.Fatalf("renderStatusMatches: %v", err)
		}
	})
	if !strings.Contains(withVersion, "Version:    0.1.8") {
		t.Fatalf("detail block does not report the version:\n%s", withVersion)
	}

	// A synthesized watch row can carry no version; an empty label is worse
	// than no label, so the line is dropped entirely.
	withoutVersion := captureStdout(t, func() {
		if err := renderStatusMatches(&MarketOptions{}, []statusRow{{
			Name: "clitest", Source: "upload", State: "running",
		}}); err != nil {
			t.Fatalf("renderStatusMatches: %v", err)
		}
	})
	if strings.Contains(withoutVersion, "Version:") {
		t.Fatalf("detail block prints an empty version line:\n%s", withoutVersion)
	}
}

// captureStdout collects what a renderer wrote, for the paths that print with
// fmt.Printf straight to os.Stdout.
func captureStdout(t *testing.T, run func()) string {
	t.Helper()
	saved := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	run()
	os.Stdout = saved
	w.Close()
	defer r.Close()

	var sb strings.Builder
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
		sb.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return sb.String()
}
