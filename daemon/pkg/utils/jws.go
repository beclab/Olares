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

func ValidateJWS(token string) (bool, string, error) {
	didServiceURL, err := url.JoinPath(commands.OLARES_REMOTE_SERVICE, "/did/1.0/name/")
	if err != nil {
		klog.Errorf("failed to parse DID gate service URL: %v, Olares remote service: %s", err, commands.OLARES_REMOTE_SERVICE)
		return false, "", err
	}
	jws.DIDGateURL = didServiceURL

	// Validate the JWS token with a 20-minute expiration time
	checkJWS, err := jws.CheckJWS(token, 20*60*1000)
	if err != nil {
		if strings.HasPrefix(err.Error(), "timestamp") {
			err = fmt.Errorf("%v, server time: %s", err, time.Now().UTC().Format(time.RFC3339))
		}
		klog.Errorf("failed to check JWS: %v, on %s", err, jws.DIDGateURL)
		return false, "", err
	}

	if checkJWS == nil {
		err := fmt.Errorf("JWS validation failed: JWS is nil")
		klog.Error(err)
		return false, "", err
	}

	// Convert to JSON with indentation
	bytes, err := json.MarshalIndent(checkJWS, "", "  ")
	if err != nil {
		klog.Errorf("failed to marshal result: %v", err)
	}

	klog.Infof("JWS validation successful: %s", string(bytes))
	return true, checkJWS.OlaresID, nil
}
