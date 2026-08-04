package timeout

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	DefaultSharedBackendResponseTimeout = 10 * time.Minute
	MinResponseTimeout                  = DefaultSharedBackendResponseTimeout
	MaxResponseTimeout                  = 24 * time.Hour
	EnvDefaultResponseTimeout           = "SHARED_BACKEND_RESPONSE_TIMEOUT"
	defaultTimeoutGatewayString         = "600s"
)

var configSecondsPattern = regexp.MustCompile(`^(\d+)s?$`)

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

// ClampResponseTimeout raises values below the platform default to 600s.
func ClampResponseTimeout(d time.Duration) time.Duration {
	if d < DefaultSharedBackendResponseTimeout {
		return DefaultSharedBackendResponseTimeout
	}
	return d
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
	if clamped := ClampResponseTimeout(d); clamped != d {
		klog.Warningf("%s=%q below minimum %s, using %s", EnvDefaultResponseTimeout, v, defaultTimeoutGatewayString, defaultTimeoutGatewayString)
		return clamped
	}
	return d
}

func EffectiveResponseTimeout(defaultTimeout time.Duration, manifest *metav1.Duration, egConfigured time.Duration) (string, error) {
	if err := ValidateResponseTimeout(defaultTimeout); err != nil {
		return defaultTimeoutGatewayString, fmt.Errorf("invalid default timeout: %w", err)
	}
	effective := defaultTimeout

	for _, candidate := range []time.Duration{candidateDuration(manifest), egConfigured} {
		if candidate <= 0 {
			continue
		}
		if err := ValidateResponseTimeout(candidate); err != nil {
			return defaultTimeoutGatewayString, fmt.Errorf("invalid timeout %s: %w", candidate, err)
		}
		if candidate > effective {
			effective = candidate
		}
	}

	effective = ClampResponseTimeout(effective)
	formatted, err := FormatSeconds(effective)
	if err != nil {
		return defaultTimeoutGatewayString, fmt.Errorf("format effective timeout: %w", err)
	}
	return formatted, nil
}

func ValidateResponseTimeout(d time.Duration) error {
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
