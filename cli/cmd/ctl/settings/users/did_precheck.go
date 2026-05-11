package users

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// olaresInfoProbe fields used for create-user parity with Termipass
// (CreateUserDialog + adminStore.olaresd + olaresId domain suffix).
type olaresInfoProbe struct {
	OsVersion    string `json:"osVersion"`
	OlaresID     string `json:"olaresId"`
	TerminusName string `json:"terminusName"`
	Olaresd      string `json:"olaresd"`
	Terminusd    string `json:"terminusd"`
}

func olaresDModeLikeSPA(p olaresInfoProbe) bool {
	return strings.TrimSpace(p.Olaresd) == "1" || strings.TrimSpace(p.Terminusd) == "1"
}

func loggedInOlaresCanonicalID(p olaresInfoProbe) string {
	if s := strings.TrimSpace(p.OlaresID); s != "" {
		return s
	}
	return strings.TrimSpace(p.TerminusName)
}

func domainSuffixFromOlaresID(canonical string) (string, error) {
	canonical = strings.TrimSpace(canonical)
	i := strings.Index(canonical, "@")
	if i <= 0 || i >= len(canonical)-1 {
		return "", fmt.Errorf(
			"cannot derive domain suffix from Olares ID %q — expected logged-in identity like user@example.olares.com", canonical)
	}
	return canonical[i+1:], nil
}

func fullOlaresIDDotted(localUsername string, probe olaresInfoProbe) (fullAt string, dotted string, err error) {
	canonical := loggedInOlaresCanonicalID(probe)
	suffix, err := domainSuffixFromOlaresID(canonical)
	if err != nil {
		return "", "", err
	}
	local := strings.TrimSpace(localUsername)
	if local == "" {
		return "", "", fmt.Errorf("username is empty")
	}
	fullAt = local + "@" + suffix
	dotted = strings.ReplaceAll(fullAt, "@", ".")
	return fullAt, dotted, nil
}

// didGateBase mirrors @bytetrade/core GolbalHost.userNameToEnvironment URLs.
func didGateBase(fullNameAt string) string {
	fn := strings.ToLower(strings.TrimSpace(fullNameAt))
	if strings.HasSuffix(fn, "olares.cn") {
		return "https://api.olares.cn/did"
	}
	return "https://api.olares.com/did"
}

// precheckNewUserOlaresIDDID mirrors Termipass didStore.resolve_name /
// resolve_name_by_did before POST /api/users.
func precheckNewUserOlaresIDDID(ctx context.Context, pc *preparedClient, probe olaresInfoProbe, username string) error {
	if pc == nil {
		return fmt.Errorf("internal error: preparedClient is nil")
	}
	fullAt, dotted, err := fullOlaresIDDotted(username, probe)
	if err != nil {
		return err
	}
	if olaresDModeLikeSPA(probe) {
		path := "/api/mdns/system/1.0/name/" + url.PathEscape(dotted)
		return getExpectOKJSON(ctx, pc.Doer, path, "Olares DID name precheck")
	}
	gateURL := strings.TrimRight(didGateBase(fullAt), "/") + "/1.0/name/" + url.PathEscape(dotted)
	return externalDIDGateGET(ctx, gateURL)
}

func getExpectOKJSON(ctx context.Context, d Doer, path, what string) error {
	type empty struct{}
	var out json.RawMessage
	if err := d.DoJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return fmt.Errorf("%s (desktop %s): %w", what, path, err)
	}
	if len(strings.TrimSpace(string(out))) == 0 || string(out) == "null" {
		return fmt.Errorf("%s returned empty JSON", what)
	}
	return nil
}

func externalDIDGateGET(ctx context.Context, rawURL string) error {
	cli := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("DID gateway precheck: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("DID gateway GET %s: %w", rawURL, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	var outer struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	msg := ""
	if json.Unmarshal(body, &outer) == nil {
		msg = strings.TrimSpace(outer.Message)
		if msg == "" {
			msg = strings.TrimSpace(outer.Error)
		}
	}
	if msg == "" {
		msg = strings.TrimSpace(string(body))
	}
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	if strings.Contains(strings.ToLower(msg), "failed to resolve did") {
		return fmt.Errorf("Olares ID not resolved on blockchain (DID gateway): %s", msg)
	}
	return fmt.Errorf("DID gateway precheck failed: %s", msg)
}
