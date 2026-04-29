package dashboard

import (
	"fmt"
	"strconv"
	"time"

	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// FormatFloat renders v with the minimum number of decimals required for a
// round-trip representation (Go's 'f' verb with prec=-1). Used everywhere
// SPA's number column expects an unbounded-precision string before being
// run through format.GetDiskSize / format.GetThroughput / format.WorthValue.
//
// Hoisted to the pkg layer so cmd area subpackages don't each duplicate
// the one-liner.
func FormatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// SafeRatio returns num/den, but yields 0 (instead of NaN/Inf) when den
// is zero. Mirrors the SPA's many `total > 0 ? value / total : 0` guards.
func SafeRatio(num, den float64) float64 {
	if den == 0 {
		return 0
	}
	return num / den
}

// FormatRateAny coerces an arbitrary tx/rx value (could be number or
// string) to a SPA-style "X B/s" throughput line. The system-ifs payload
// returns strings on some Olares versions and numbers on others; this
// function unifies both shapes.
func FormatRateAny(v any) string {
	if v == nil {
		return "-"
	}
	switch x := v.(type) {
	case string:
		if x == "" {
			return "-"
		}
		return format.GetThroughput(x)
	case float64:
		return format.GetThroughput(FormatFloat(x))
	case int, int64, int32:
		return format.GetThroughput(fmt.Sprintf("%d", x))
	default:
		return fmt.Sprintf("%v", x)
	}
}

// ParseRFCTimestamp converts an RFC3339 timestamp string to milliseconds
// since epoch (the unit format.FormatTime expects). Returns 0 on parse
// failure.
func ParseRFCTimestamp(s string) int64 {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0
	}
	return t.UnixMilli()
}

// SampleFloat extracts the numeric reading from a single last-sample
// entry returned by format.GetLastMonitoringData. Empty samples (status
// "no_data" or absent metric) yield 0 rather than NaN; non-numeric raw
// values also fall through to 0 so leaf renderers can keep their
// arithmetic simple.
func SampleFloat(s format.LastMonitoringSample) float64 {
	if s.Empty {
		return 0
	}
	v, err := strconv.ParseFloat(s.RawValue, 64)
	if err != nil {
		return 0
	}
	return v
}

// LastSampleFromRow picks the last (timestamp, value) tuple out of a
// PromQL-style range result. It prefers `values[-1]` (the actual range
// data) and falls back to `value` (the instant-vector shape some leaves
// also accept) when the range is empty. Returns Empty=true when neither
// shape carries usable data.
//
// Hoisted to pkg so disk's per-partition rendering and overview's
// per-node rendering share one canonical decoder.
func LastSampleFromRow(values [][]any, value []any) format.LastMonitoringSample {
	if len(values) > 0 {
		row := values[len(values)-1]
		if len(row) >= 2 {
			ts, _ := row[0].(float64)
			s := fmt.Sprintf("%v", row[1])
			return format.LastMonitoringSample{Timestamp: ts, RawValue: s}
		}
	}
	if len(value) >= 2 {
		ts, _ := value[0].(float64)
		s := fmt.Sprintf("%v", value[1])
		return format.LastMonitoringSample{Timestamp: ts, RawValue: s}
	}
	return format.LastMonitoringSample{Empty: true}
}
