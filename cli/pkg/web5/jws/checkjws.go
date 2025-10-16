package jws

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/web5/crypto/dsa"
	"github.com/beclab/Olares/cli/pkg/web5/dids/didcore"

	"github.com/syndtr/goleveldb/leveldb"
)

var (
	DIDGateURL     = "https://did-gate-v3.bttcdn.com/1.0/name/"
	DIDGateTimeout = 10 * time.Second
	DIDCachePath   = "/var/lib/olares"
)

var (
	db *leveldb.DB
)

func init() {
	var (
		err  error
		info os.FileInfo
	)
	info, err = os.Stat(DIDCachePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the directory if it doesn't exist
			if err := os.MkdirAll(DIDCachePath, 0755); err != nil {
				panic(fmt.Sprintf("failed to create directory: %v", err))
			}
		} else {
			panic(fmt.Sprintf("failed to check directory: %v", err))
		}
	}

	if info == nil || !info.IsDir() {
		err = os.RemoveAll(DIDCachePath)
		if err != nil {
			panic(fmt.Sprintf("failed to remove file: %v", err))
		}

		err = os.MkdirAll(DIDCachePath, 0755)
		if err != nil {
			panic(fmt.Sprintf("failed to create directory: %v", err))
		}
	}

	dbPath := DIDCachePath + "/did_cache.db"
	db, err = leveldb.OpenFile(dbPath, nil)
	if err != nil {
		// If file exists but can't be opened, try to remove it
		err := os.RemoveAll(dbPath)
		if err != nil {
			panic(fmt.Sprintf("failed to remove existing db: %v", err))
		}

		// Try to create a new database
		db, err = leveldb.OpenFile(dbPath, nil)
		if err != nil {
			panic(fmt.Sprintf("failed to create leveldb: %v", err))
		}
	}
}

// CheckJWSResult represents the result of checking a JWS
type CheckJWSResult struct {
	OlaresID string      `json:"olares_id"`
	Body     interface{} `json:"body"`
	KID      string      `json:"kid"`
}

// resolveDID resolves a DID either from cache or from the DID gate
func ResolveOlaresName(olares_id string) (*didcore.ResolutionResult, error) {
	name := strings.Replace(olares_id, "@", ".", -1)
	// Try to get from cache first
	cached, err := db.Get([]byte(name), nil)
	if err == nil {
		var result didcore.ResolutionResult
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	// If not in cache, fetch from DID gate
	client := &http.Client{
		Timeout: DIDGateTimeout,
	}
	resp, err := client.Get(DIDGateURL + name)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DID from gate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DID gate returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result didcore.ResolutionResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse DID document: %w", err)
	}

	// Cache the result
	if err := db.Put([]byte(name), body, nil); err != nil {
		// Log error but don't fail
		fmt.Printf("failed to cache DID document: %v\n", err)
	}

	return &result, nil
}

// CheckJWS verifies a JWS and returns the terminus name, body and kid
func CheckJWS(jws string, duration int64) (*CheckJWSResult, error) {
	var kid string
	var name string
	var timestamp int64

	// Split JWS into segments
	segs := strings.Split(jws, ".")
	if len(segs) != 3 {
		return nil, fmt.Errorf("invalid jws: wrong number of segments")
	}

	// Parse header
	headerBytes, err := base64.RawURLEncoding.DecodeString(segs[0])
	if err != nil {
		return nil, fmt.Errorf("invalid jws: failed to decode header: %w", err)
	}

	var header struct {
		KID string `json:"kid"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("invalid jws: failed to parse header: %w", err)
	}
	kid = header.KID

	// Parse payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(segs[1])
	if err != nil {
		return nil, fmt.Errorf("invalid jws: failed to decode payload: %w", err)
	}

	var payload struct {
		DID       string                 `json:"did"`
		Name      string                 `json:"name"`
		Time      string                 `json:"time"`
		Domain    string                 `json:"domain"`
		Challenge string                 `json:"challenge"`
		Body      map[string]interface{} `json:"body"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("invalid jws: failed to parse payload: %w", err)
	}

	name = payload.Name
	// Convert time string to int64
	timestamp, err = strconv.ParseInt(payload.Time, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid time format: %w", err)
	}

	// Validate required fields
	if name == "" || kid == "" || timestamp == 0 {
		return nil, fmt.Errorf("invalid jws: missing required fields")
	}

	// Check timestamp
	now := time.Now().UnixMilli()
	if now-timestamp > duration {
		return nil, fmt.Errorf("timestamp is out of range")
	}

	// Resolve DID
	resolutionResult, err := ResolveOlaresName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve DID: %w", err)
	}

	// Verify DID matches
	if resolutionResult.Document.ID != kid {
		sid := resolutionResult.Document.ID + resolutionResult.Document.VerificationMethod[0].ID
		if sid != kid {
			return nil, fmt.Errorf("DID does not match: expected %s, got %  s", sid, kid)
		}
	}
	// Get verification method
	if len(resolutionResult.Document.VerificationMethod) == 0 || resolutionResult.Document.VerificationMethod[0].PublicKeyJwk == nil {
		return nil, fmt.Errorf("invalid DID document: missing verification method")
	}

	// Verify signature
	toVerify := segs[0] + "." + segs[1]
	signature, err := base64.RawURLEncoding.DecodeString(segs[2])
	if err != nil {
		return nil, fmt.Errorf("invalid jws: failed to decode signature: %w", err)
	}

	verified, err := dsa.Verify([]byte(toVerify), signature, *resolutionResult.Document.VerificationMethod[0].PublicKeyJwk)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature: %w", err)
	}
	if !verified {
		return nil, fmt.Errorf("invalid signature")
	}

	result := CheckJWSResult{
		OlaresID: name,
		Body:     payload,
		KID:      kid,
	}

	return &result, nil
}
