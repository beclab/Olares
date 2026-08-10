package model

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Credential input for the provider write verbs.
//
// Two ways in, because they fail differently. `--credential key=value` is
// what a person types once at a terminal, and every value it carries is a
// string — which is what every credential field in Router's vendor catalog
// actually is. `--credentials-json` takes the whole object, from a file or
// from stdin, and preserves types for the rare non-string field.
//
// The second one is also the only form that keeps a key out of the shell
// history and out of `ps` on a shared machine, which is why the help text
// on every verb points at it.

// credentialFlagUsage is repeated verbatim on each verb that accepts
// credentials, so a person reading `--help` on any one of them learns the
// same thing.
const credentialFlagUsage = "credential field as key=value; repeatable"

const credentialsJSONFlagUsage = "credentials as a JSON object, from a file or - for stdin; " +
	"preferred over --credential, which leaves the secret in your shell history"

// collectCredentials merges the two input forms into the object Router
// expects. Returns nil when neither flag was given, which the update verb
// reads as "leave the stored credentials alone". An explicit empty object is
// not the same thing and survives as an empty map, because an upstream that
// wants no key at all is a real case.
func collectCredentials(pairs []string, jsonSource string) (map[string]any, error) {
	out := map[string]any{}
	given := len(pairs) > 0

	if s := strings.TrimSpace(jsonSource); s != "" {
		given = true
		raw, err := readCredentialSource(s)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("--credentials-json must contain a JSON object of credential fields: %w", err)
		}
	}

	for _, p := range pairs {
		key, value, ok := strings.Cut(p, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf("--credential %q is not in key=value form", p)
		}
		out[key] = value
	}

	if !given {
		return nil, nil
	}
	return out, nil
}

func readCredentialSource(source string) ([]byte, error) {
	if source == "-" {
		raw, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read credentials from stdin: %w", err)
		}
		return raw, nil
	}
	raw, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("read credentials file: %w", err)
	}
	return raw, nil
}

// maskCredentialValue renders a stored credential for display. Router already
// masks what it returns from the credentials-form route, replacing secrets
// with a sentinel; this collapses whatever came back — sentinel, empty, or a
// field the vendor does not treat as secret — into one line that never grows
// long enough to wrap.
func maskCredentialValue(v any) string {
	switch t := v.(type) {
	case nil:
		return "(unset)"
	case string:
		if t == "" {
			return "(empty)"
		}
		return truncate(t, 60)
	default:
		raw, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprintf("%v", t)
		}
		return truncate(string(raw), 60)
	}
}
