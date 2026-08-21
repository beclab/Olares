package router

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The two credentials a caller can name have to beat the platform, and the flag
// has to beat the environment: a scripted run exports a key once and then
// overrides it for one command, which only works in that order.
func TestAnExplicitKeyBeatsTheEnvironmentAndBothBeatThePlatform(t *testing.T) {
	t.Setenv(dataPlaneKeyEnv, "sk-from-env")

	if got := resolveDataPlaneAuth("sk-from-flag"); got.Mode != authKey || got.Key != "sk-from-flag" {
		t.Errorf("--api-key did not win over %s: got %+v", dataPlaneKeyEnv, got)
	}
	if got := resolveDataPlaneAuth("   "); got.Mode != authKey || got.Key != "sk-from-env" {
		t.Errorf("a blank --api-key should fall through to %s, got %+v", dataPlaneKeyEnv, got)
	}

	t.Setenv(dataPlaneKeyEnv, "")
	if got := resolveDataPlaneAuth(""); got.Mode != authPlatform || got.Key != "" {
		t.Errorf("with no key named the call should go as the platform's caller, got %+v", got)
	}
}

// A keyless call is the normal one now, and what makes it keyless is the
// absence of the header rather than any value in it. Sending an empty or
// malformed Authorization would be read by Router as a Bearer attempt and
// refused, never reaching the X-BFL-USER branch that is the point of this.
func TestAKeylessCallSendsNoAuthorizationHeaderAtAll(t *testing.T) {
	t.Setenv(dataPlaneKeyEnv, "")
	pc := &preparedClient{router: newRouterClient(nil, "https://router.example", "someone@olares")}

	keyless := dataPlane(pc, "")
	if _, ok := keyless.headers["Authorization"]; ok {
		t.Errorf("a keyless call set Authorization to %q; Router would read that as a Bearer attempt",
			keyless.headers["Authorization"])
	}

	keyed := dataPlane(pc, "sk-abcd")
	if got := keyed.headers["Authorization"]; got != "Bearer sk-abcd" {
		t.Errorf("a named key should travel as a Bearer, got %q", got)
	}
	if _, ok := pc.router.headers["Authorization"]; ok {
		t.Error("presenting a key mutated the shared client instead of cloning it")
	}
}

// Calling a model used to create a credential: the first `router call` of any
// kind on a laptop minted an unrestricted, never-expiring key and wrote it to
// the keychain. Both halves are gone, and each is checked where it can be seen
// — a key is created by posting to the keys route, and saved by writing the
// keychain, so the only place either may appear is the verb that was asked to
// do it.
func TestNothingButKeyIssueCreatesACredential(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list the package: %v", err)
	}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		body, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		src := string(body)
		if strings.Contains(src, "keychain.Set") {
			t.Errorf("%s writes the keychain. Nothing saves a data-plane key now: a call that "+
				"names no key goes as the platform's caller instead", name)
		}
		// key.go is where issuing one is the verb the user asked for.
		if name != "key.go" && strings.Contains(src, `"POST", epAPIKeys`) {
			t.Errorf("%s issues an API key. Reaching the data plane must not create a credential "+
				"as a side effect of using one", name)
		}
	}
}

// Three refusals send a reader somewhere different, and the two that arrived
// with keyless calling are the ones a stale message would answer wrongly: an
// old Router needs upgrading rather than a key minted around it, and a person
// Router has not met needs the console plane rather than a credential.
func TestTheCredentialRefusalsSayWhichOneToFix(t *testing.T) {
	for _, tc := range []struct {
		code  string
		wants []string
	}{
		{"missing_credentials", []string{"v2.2.1", "--api-key", dataPlaneKeyEnv}},
		{"unknown_bfl_user", []string{"router model list"}},
		{"invalid_api_key", []string{"key current --forget", dataPlaneKeyEnv}},
	} {
		err := callErr(&RouterError{Status: 401, Type: "authentication_error", Code: tc.code, Message: "no"})
		if err == nil {
			t.Fatalf("%s: callErr swallowed the error", tc.code)
		}
		for _, want := range tc.wants {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s does not mention %q:\n%s", tc.code, want, err)
			}
		}
		if strings.Contains(err.Error(), "issues a fresh one") {
			t.Errorf("%s still promises a key will be issued:\n%s", tc.code, err)
		}
	}
}
