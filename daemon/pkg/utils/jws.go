package utils

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/beclab/Olares/cli/pkg/web5/jws"
	"k8s.io/klog/v2"
)

func ValidateJWS(token string) (bool, string, error) {
	didGateDomain := os.Getenv("DID_GATE_URL")
	if didGateDomain != "" {
		newUrl := fmt.Sprintf("%s/1.0/name/", didGateDomain)
		_, err := url.Parse(newUrl)
		if err != nil {
			klog.Warning("failed to parse DID gate URL in environment variable: %v", err)
		} else {
			jws.DIDGateURL = newUrl
		}
	}

	// Validate the JWS token with a 20-minute expiration time
	checkJWS, err := jws.CheckJWS(token, 20*60*1000)
	if err != nil {
		klog.Errorf("failed to check JWS: %v", err)
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
