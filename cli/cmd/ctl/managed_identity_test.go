package ctl

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/spf13/cobra"
)

// The fifteen importer unit tests all call managedImporter.run directly, so
// none of them noticed that Cobra had stopped calling it: a subtree with a
// PersistentPreRun of its own shadows the root's, and eighteen subtrees had
// grown one to set SilenceUsage. `market list` in a fresh container reported
// that no profile was configured, and `profile list` — the one tree without a
// hook — repaired the container for every command that came after it.
//
// These tests are the missing layer. They drive the real tree the way a
// container does and only look at what ends up on disk.

const (
	managedOlaresID = "tester@example.invalid"
	managedAppName  = "lares"
)

// managedContainer stands up what app-service gives a container: a credential
// mount, and a cache directory to persist into. Everything the import touches
// is redirected under t.TempDir(), and the grant exchange goes to a local
// server that refuses it, which keeps the test off both DNS and the system
// keychain — a refusal creates the profile entry without a token to store
// (pkg/credential: TestImport_TransientRefreshFailureStillCreatesTheEntry).
func managedContainer(t *testing.T) {
	t.Helper()

	// Several verbs print straight to os.Stdout / os.Stderr rather than to
	// the command's streams, and a fake profile table in the middle of the
	// suite output reads as a real one.
	discardProcessOutput(t)

	// Empty reads as unset in both resolvers, and t.Setenv restores the
	// developer's real values afterwards.
	t.Setenv("OLARES_CLI_HOME", "")
	t.Setenv("OLARES_CLI_DATA_DIR", "")
	t.Setenv("OLARES_CLI_CACHE_DIR", t.TempDir())

	refusing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "refresh unavailable", http.StatusInternalServerError)
	}))
	t.Cleanup(refusing.Close)

	credentialsDir := t.TempDir()
	mount, err := json.Marshal(map[string]string{
		"refreshToken": "grant-from-the-mount",
		"olaresId":     managedOlaresID,
		"appName":      managedAppName,
	})
	if err != nil {
		t.Fatalf("encode credential.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(credentialsDir, "credential.json"), mount, 0o600); err != nil {
		t.Fatalf("write credential.json: %v", err)
	}
	t.Setenv(credential.EnvCredentialsDir, credentialsDir)

	// The profile the import takes over exists only to carry the auth URL:
	// the real one is derived from the Olares ID, and an exchange against
	// it would leave the test asking the network what it thinks of
	// example.invalid.
	if err := cliconfig.SaveMultiProfileConfig(&cliconfig.MultiProfileConfig{
		Profiles: []cliconfig.ProfileConfig{{
			OlaresID:        managedOlaresID,
			AuthURLOverride: refusing.URL,
		}},
	}); err != nil {
		t.Fatalf("seed config.json: %v", err)
	}
}

func discardProcessOutput(t *testing.T) {
	t.Helper()

	sink, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	stdout, stderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = sink, sink
	t.Cleanup(func() {
		os.Stdout, os.Stderr = stdout, stderr
		sink.Close()
	})
}

// assertImported reads config.json back. A profile is only marked managed by
// the importer — resolving one never writes the file — so the flag on disk
// says the prologue ran, whatever the command itself went on to do.
func assertImported(t *testing.T, verb string) {
	t.Helper()

	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("`olares-cli %s`: read config.json: %v", verb, err)
	}
	profile := cfg.FindByOlaresID(managedOlaresID)
	if profile == nil {
		t.Fatalf("`olares-cli %s`: config.json lost the profile entirely", verb)
	}
	if !profile.Managed {
		t.Fatalf("`olares-cli %s` ran without importing the mounted credential: "+
			"the profile is still unmanaged, which is what a container sees as "+
			"\"no Olares profile is configured\"", verb)
	}
	if profile.AppName != managedAppName {
		t.Errorf("`olares-cli %s`: profile says app %q, mount says %q",
			verb, profile.AppName, managedAppName)
	}
}

// TestEveryKindOfSubtreeImportsTheMountedCredential covers the four shapes a
// path through the tree can have: no hook anywhere below the root, a hook on
// the group, a hook on the command that runs, and a group two levels down.
// Each one gets its own container, because one invocation is one container.
func TestEveryKindOfSubtreeImportsTheMountedCredential(t *testing.T) {
	// Every verb here fails — the grant exchange is refused before the
	// command reaches its own backend. That failure is the container's
	// second line of defence working, not a problem for the assertion.
	// Every verb has to be a leaf that runs: Cobra answers a group with
	// help and returns before it looks for a pre-run at all.
	for _, verb := range [][]string{
		{"market", "list"},
		{"files", "nfs", "history", "list"},
		{"profile", "list"},
		{"cluster", "pod", "list"},
	} {
		t.Run(strings.Join(verb, " "), func(t *testing.T) {
			managedContainer(t)

			root := NewDefaultCommand()
			root.SetOut(io.Discard)
			root.SetErr(io.Discard)
			root.SetArgs(verb)
			_ = root.Execute()

			assertImported(t, strings.Join(verb, " "))
		})
	}
}

// TestNoSubtreeHidesTheIdentityImport is the tripwire for the shape of the
// bug rather than the bug itself. wireManagedIdentity wraps whatever hooks it
// finds, so a new one is covered the moment it is added — but only the list
// below says which paths the test above actually exercises, and a new hook
// that nobody adds a case for is a subtree nobody is checking.
func TestNoSubtreeHidesTheIdentityImport(t *testing.T) {
	root := NewDefaultCommand()

	if !root.SilenceUsage {
		t.Error("the root lost SilenceUsage: subtrees will start declaring " +
			"PersistentPreRun again to silence themselves, which is exactly " +
			"how the identity import went missing")
	}

	declared := []string{}
	var walk func(*cobra.Command)
	walk = func(cmd *cobra.Command) {
		if cmd.PersistentPreRun != nil || cmd.PersistentPreRunE != nil {
			declared = append(declared, cmd.CommandPath())
		}
		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(root)

	// The root binds viper, `files nfs` gates on the backend version, and
	// `backups` arrives with a hook of its own from backups-sdk — that one
	// is host-side and registered only when OLARES_CLI_REMOTE_ONLY is unset.
	known := map[string]bool{
		"olares-cli":           true,
		"olares-cli files nfs": true,
		"olares-cli backups":   true,
	}
	found := map[string]bool{}
	for _, path := range declared {
		found[path] = true
		if !known[path] {
			t.Errorf("`%s` declares a persistent pre-run. wireManagedIdentity "+
				"already wraps it, so identity is fine — add the subtree to "+
				"TestEveryKindOfSubtreeImportsTheMountedCredential and to this "+
				"list so it stays that way. If the hook only sets SilenceUsage, "+
				"delete it: the root covers the whole tree.", path)
		}
	}
	for _, path := range []string{"olares-cli", "olares-cli files nfs"} {
		if !found[path] {
			t.Errorf("`%s` no longer declares a persistent pre-run; found them on %v", path, declared)
		}
	}
}
