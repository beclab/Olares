package clusterop

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"
)

var emptyParamsDigest = mustDigestParams(nil)

// CanonicalParams normalizes caller-supplied params without persisting them.
// Missing or whitespace-only params mean "no params", which is represented as
// the empty JSON object for idempotency.
func CanonicalParams(raw json.RawMessage) (json.RawMessage, error) {
	if strings.TrimSpace(string(raw)) == "" {
		return json.RawMessage(`{}`), nil
	}

	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(canonical), nil
}

// DigestParams returns the SHA-256 hex digest of canonicalized params.
func DigestParams(raw json.RawMessage) (string, error) {
	canonical, err := CanonicalParams(raw)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}

func mustDigestParams(raw json.RawMessage) string {
	digest, err := DigestParams(raw)
	if err != nil {
		panic(err)
	}
	return digest
}
