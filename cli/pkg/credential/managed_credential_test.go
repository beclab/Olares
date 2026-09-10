package credential

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeCredential(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, credentialFilename), []byte(body), 0o600); err != nil {
		t.Fatalf("write credential.json: %v", err)
	}
	return dir
}

func TestLoadManagedCredentialHappyPath(t *testing.T) {
	dir := writeCredential(t, `{"refreshToken":"tok","olaresId":"alice@olares.com","appName":"lares"}`)
	t.Setenv(EnvCredentialsDir, dir)

	cred, ok := LoadManagedCredential()
	if !ok {
		t.Fatal("want ok=true for a complete credential")
	}
	if cred.RefreshToken != "tok" || cred.OlaresID != "alice@olares.com" || cred.AppName != "lares" {
		t.Fatalf("cred = %+v", cred)
	}
}

func TestLoadManagedCredentialEnvUnset(t *testing.T) {
	t.Setenv(EnvCredentialsDir, "")
	if _, ok := LoadManagedCredential(); ok {
		t.Fatal("want ok=false when the env var is unset")
	}
}

func TestLoadManagedCredentialFileMissing(t *testing.T) {
	t.Setenv(EnvCredentialsDir, t.TempDir())
	if _, ok := LoadManagedCredential(); ok {
		t.Fatal("want ok=false when credential.json is absent")
	}
}

func TestLoadManagedCredentialRejectsMalformed(t *testing.T) {
	cases := map[string]struct {
		body string
		want string
	}{
		"not json":         {`{`, "parse"},
		"empty object":     {`{}`, "refreshToken, olaresId, appName"},
		"no refresh token": {`{"olaresId":"a@b.c","appName":"lares"}`, "refreshToken"},
		"no olares id":     {`{"refreshToken":"tok","appName":"lares"}`, "olaresId"},
		"no app name":      {`{"refreshToken":"tok","olaresId":"a@b.c"}`, "appName"},
		"blank fields":     {`{"refreshToken":" ","olaresId":"a@b.c","appName":"lares"}`, "refreshToken"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dir := writeCredential(t, tc.body)
			if _, err := loadManagedCredentialFrom(dir); err == nil {
				t.Fatal("want an error")
			} else if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not mention %q", err, tc.want)
			}
			t.Setenv(EnvCredentialsDir, dir)
			if _, ok := LoadManagedCredential(); ok {
				t.Fatal("want ok=false")
			}
		})
	}
}
