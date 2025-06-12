package utils

import (
	"encoding/json"

	"github.com/beclab/Olares/cli/pkg/web5/jws"
	"k8s.io/klog/v2"
)

func ValidateJWS(token string) bool {
	checkJWS, err := jws.CheckJWS(token, 20*60*1000)
	if err != nil {
		klog.Errorf("failed to check JWS: %v", err)
		return false
	}

	if checkJWS == nil {
		klog.Error("JWS validation failed: JWS is nil")
		return false
	}

	// Convert to JSON with indentation
	bytes, err := json.MarshalIndent(checkJWS, "", "  ")
	if err != nil {
		klog.Errorf("failed to marshal result: %v", err)
	}

	klog.Infof("JWS validation successful: %s", string(bytes))
	return true
}
