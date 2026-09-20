package credential

import (
	"fmt"
	"net/http"
)

func FormatHTTPAuthError(status int, _ []byte, olaresID string) error {
	if status == http.StatusForbidden {
		return fmt.Errorf(
			"server rejected the request (HTTP 403); the active identity does not have permission for this resource or action; inspect it with `olares-cli profile whoami --refresh`",
		)
	}
	if olaresID != "" {
		return fmt.Errorf(
			"server rejected the access token (HTTP %d); please run: olares-cli profile login --olares-id %s",
			status, olaresID,
		)
	}
	return fmt.Errorf(
		"server rejected the access token (HTTP %d); please re-run `olares-cli profile login`",
		status,
	)
}
