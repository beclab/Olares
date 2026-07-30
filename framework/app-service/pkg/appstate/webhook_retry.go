package appstate

import (
	"strings"
)

// IsRetryableWebhookError reports whether err is a transient admission-webhook
// reachability failure (e.g. Service Endpoints still pointing at a dead pod IP
// after reboot). Callers should requeue instead of transitioning to StopFailed.
func IsRetryableWebhookError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "webhook") && !strings.Contains(msg, "8433") {
		return false
	}
	for _, needle := range retryableWebhookNeedles {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

var retryableWebhookNeedles = []string{
	"failed calling webhook",
	"502 bad gateway",
	"code 502",
	"connection refused",
	"no endpoints available",
	"no endpoints",
}
