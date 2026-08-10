package utils

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/web5/jws"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"k8s.io/klog/v2"
)

// SignedRequest is a verified JWS: who signed it, and what their key covered.
//
// Body matters as much as OlaresID. A route that reads only the identity
// treats every signature as a bearer credential for everything that identity
// may do; the body is where a caller says which single operation it signed
// for, so a dangerous route can insist on being named.
type SignedRequest struct {
	OlaresID string
	Body     any
}

// ValidateJWSDetail verifies a JWS and reports both who signed it and what
// their key covered.
func ValidateJWSDetail(token string) (SignedRequest, error) {
	didServiceURL, err := url.JoinPath(commands.OLARES_REMOTE_SERVICE, "/did/1.0/name/")
	if err != nil {
		klog.Errorf("failed to parse DID gate service URL: %v, Olares remote service: %s", err, commands.OLARES_REMOTE_SERVICE)
		return SignedRequest{}, err
	}

	// Validate the JWS token with a 20-minute expiration time
	checkJWS, err := jws.CheckJWS(didServiceURL, token, 20*60*1000)
	if err != nil {
		if strings.HasPrefix(err.Error(), "timestamp") {
			err = fmt.Errorf("%v, server time: %s", err, time.Now().UTC().Format(time.RFC3339))
		}
		klog.Errorf("failed to check JWS: %v, on %s", err, didServiceURL)
		return SignedRequest{}, err
	}

	if checkJWS == nil {
		err := fmt.Errorf("JWS validation failed: JWS is nil")
		klog.Error(err)
		return SignedRequest{}, err
	}

	// Convert to JSON with indentation
	bytes, err := json.MarshalIndent(checkJWS, "", "  ")
	if err != nil {
		klog.Errorf("failed to marshal result: %v", err)
	}

	klog.Infof("JWS validation successful: %s", string(bytes))
	return SignedRequest{OlaresID: checkJWS.OlaresID, Body: checkJWS.Body}, nil
}
