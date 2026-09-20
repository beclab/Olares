package nodestatus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Path is where every node, master included, reports what it is and how it is
// doing. One protocol for the whole cluster is the point of the endpoint, so
// the address is the only thing that differs between two reads of it.
const Path = "/system/node-status"

// authHeader is the access token header the daemon's routes read.
const authHeader = "X-Authorization"

// Fetch reads one node's own report of itself from baseURL, e.g.
// "http://10.0.0.2:18088". The caller bounds the call through ctx: how long a
// node may take to answer depends on why it is being asked.
//
// The address it was given never comes back in the error text a caller shows a
// user; that is the caller's business, and both of them treat a failure here
// as "this node did not answer" rather than quoting the transport.
func Fetch(ctx context.Context, baseURL, token string) (Status, error) {
	return FetchWithHeaders(ctx, baseURL, map[string]string{authHeader: token})
}

// FetchWithHeaders reads a node status using the credential chosen by its
// caller. Interactive reads use an access token; a cluster-operation precheck
// uses the operation-bound owner signature from the scan callback.
func FetchWithHeaders(ctx context.Context, baseURL string, headers map[string]string) (Status, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+Path, http.NoBody)
	if err != nil {
		return Status{}, err
	}
	for name, value := range headers {
		if value != "" {
			req.Header.Set(name, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Status{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Status{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return Status{}, fmt.Errorf("node returned %d", resp.StatusCode)
	}
	return FromEnvelope(body)
}

// FromEnvelope unwraps the daemon response envelope. A reply that is not one
// is an error rather than an empty status: a zero-valued status declares no
// capabilities and no health, which a caller would read as a node that
// answered and has nothing to offer.
func FromEnvelope(body []byte) (Status, error) {
	var env struct {
		Data *Status `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return Status{}, fmt.Errorf("decode node status: %w", err)
	}
	if env.Data == nil {
		return Status{}, errors.New("node reply carried no status")
	}
	return *env.Data, nil
}
