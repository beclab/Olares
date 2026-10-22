package timeout

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/timeoutprofile"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// Platform Shared (east-west EG) timeouts come from the timeout catalog so this
// hop stays aligned with the rest of the platform.
var (
	DefaultSharedBackendResponseTimeout = timeoutprofile.Default().Profile(timeoutprofile.Interactive).EdgeIdle
	MinResponseTimeout                  = timeoutprofile.Default().Bounds.TotalMin
	MaxResponseTimeout                  = timeoutprofile.Default().Bounds.TotalMax
	defaultTimeoutGatewayString         = mustSeconds(DefaultSharedBackendResponseTimeout)
)

const EnvDefaultResponseTimeout = "SHARED_BACKEND_RESPONSE_TIMEOUT"

var configSecondsPattern = regexp.MustCompile(`^(\d+)s?$`)

func mustSeconds(d time.Duration) string {
	s, err := FormatSeconds(d)
	if err != nil {
		panic(err)
	}
	return s
}

// ParseSeconds parses a duration configured in whole seconds (e.g. "600" or "600s").
func ParseSeconds(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if strings.ContainsAny(s, "hm") || strings.HasSuffix(s, "ms") {
		return 0, fmt.Errorf("duration must use seconds only (e.g. 600 or 600s)")
	}
	m := configSecondsPattern.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("duration must be a whole number of seconds (e.g. 600 or 600s)")
	}
	secs, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse seconds: %w", err)
	}
	return time.Duration(secs) * time.Second, nil
}

// FormatSeconds formats a duration as a Gateway API seconds string (e.g. "600s").
func FormatSeconds(d time.Duration) (string, error) {
	if d <= 0 {
		return "", fmt.Errorf("duration must be positive")
	}
	if d%time.Second != 0 {
		return "", fmt.Errorf("duration must be a whole number of seconds: %s", d)
	}
	return fmt.Sprintf("%ds", d/time.Second), nil
}

func DefaultTimeout() time.Duration {
	v := strings.TrimSpace(os.Getenv(EnvDefaultResponseTimeout))
	if v == "" {
		return DefaultSharedBackendResponseTimeout
	}
	d, err := ParseSeconds(v)
	if err != nil {
		klog.Warningf("invalid %s=%q, using default %s: %v", EnvDefaultResponseTimeout, v, defaultTimeoutGatewayString, err)
		return DefaultSharedBackendResponseTimeout
	}
	if err := ValidateResponseTimeout(d); err != nil {
		klog.Warningf("invalid %s=%q, using default %s: %v", EnvDefaultResponseTimeout, v, defaultTimeoutGatewayString, err)
		return DefaultSharedBackendResponseTimeout
	}
	return d
}

// EffectiveResponseTimeout returns the Gateway API seconds string for HTTPRoute
// request/backendRequest timeouts.
//
// When manifest (entrance timeouts.total) is set and in bounds, it wins over
// the platform default and any EG floor. Otherwise the effective value is
// max(defaultTimeout, egConfigured) within [MinResponseTimeout, MaxResponseTimeout].
func EffectiveResponseTimeout(defaultTimeout time.Duration, manifest *metav1.Duration, egConfigured time.Duration) (string, error) {
	if err := ValidateResponseTimeout(defaultTimeout); err != nil {
		return defaultTimeoutGatewayString, fmt.Errorf("invalid default timeout: %w", err)
	}

	if m := candidateDuration(manifest); m > 0 {
		if err := ValidateResponseTimeout(m); err != nil {
			klog.Warningf("invalid entrance timeouts.total=%s, falling back to platform/floor: %v", m, err)
		} else {
			formatted, err := FormatSeconds(m)
			if err != nil {
				return defaultTimeoutGatewayString, fmt.Errorf("format manifest timeout: %w", err)
			}
			return formatted, nil
		}
	}

	effective := defaultTimeout
	if egConfigured > 0 {
		if err := ValidateResponseTimeout(egConfigured); err != nil {
			return defaultTimeoutGatewayString, fmt.Errorf("invalid timeout %s: %w", egConfigured, err)
		}
		if egConfigured > effective {
			effective = egConfigured
		}
	}

	formatted, err := FormatSeconds(effective)
	if err != nil {
		return defaultTimeoutGatewayString, fmt.Errorf("format effective timeout: %w", err)
	}
	return formatted, nil
}

func ValidateResponseTimeout(d time.Duration) error {
	if d < MinResponseTimeout {
		return fmt.Errorf("timeout %s is below min %s", d, MinResponseTimeout)
	}
	if d > MaxResponseTimeout {
		return fmt.Errorf("timeout %s exceeds max %s", d, MaxResponseTimeout)
	}
	return nil
}

func candidateDuration(manifest *metav1.Duration) time.Duration {
	if manifest == nil {
		return 0
	}
	return manifest.Duration
}
