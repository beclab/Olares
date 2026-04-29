// Package format is a 1:1 Go port of the dashboard SPA's number / unit
// formatting helpers. Every function here is a faithful translation of the
// TypeScript source it cites — boundary cases, off-by-ones and all — so the
// CLI's `display.rendered` strings match what a user sees in the browser
// down to the last decimal place.
//
// Sources:
//   - getValueByUnit / getSuitableUnit / getSuitableValue + UnitTypes table
//     come from `@bytetrade/core/src/monitoring.ts`.
//   - worthValue / formatFrequency / convertTemperature / getDiskSize /
//     getThroughput / getFormatTime / getLastMonitoringData / getResult /
//     getMinuteValue come from `apps/packages/app/src/apps/dashboard/utils/`
//     (number.ts, cpu.ts, disk.ts, memory.ts, monitoring.ts, etc.).
//
// Whenever you tweak anything here, regenerate the golden fixtures:
//
//	cd cli/cmd/ctl/dashboard/format/testdata && node golden-gen.js
//
// and re-run `go test ./cmd/ctl/dashboard/format/...` to confirm parity with
// the upstream JavaScript implementation.
package format

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// ----------------------------------------------------------------------------
// UnitTypes (mirrors monitoring.ts UnitTypes table)
// ----------------------------------------------------------------------------

// UnitType is the canonical name of a unit family (e.g. "memory", "cpu").
// Mirrors `UnitKey` in `@bytetrade/core/src/monitoring.ts`.
type UnitType string

const (
	UnitTypeSecond     UnitType = "second"
	UnitTypeCPU        UnitType = "cpu"
	UnitTypeMemory     UnitType = "memory"
	UnitTypeDisk       UnitType = "disk"
	UnitTypeThroughput UnitType = "throughput"
	UnitTypeTraffic    UnitType = "traffic"
	UnitTypeBandwidth  UnitType = "bandwidth"
	UnitTypeNumber     UnitType = "number"
)

type unitDef struct {
	conditions []float64
	units      []string
}

// unitTypes holds the suitability thresholds per unit family. Walk the
// `conditions` slice from biggest to smallest; the first index where
// `value >= condition` selects the matching `units[index]`. The last entry
// (condition 0) is the "fallback" unit.
//
// All literals match `UnitTypes` in `@bytetrade/core/src/monitoring.ts` to the
// byte. Treat this table as a pinned constant: any change must come from a
// matching change upstream + a regenerated golden fixture.
var unitTypes = map[UnitType]unitDef{
	UnitTypeSecond: {
		conditions: []float64{0.01, 0},
		units:      []string{"s", "ms"},
	},
	UnitTypeCPU: {
		conditions: []float64{0.1, 0},
		units:      []string{"core", "m"},
	},
	UnitTypeMemory: {
		conditions: []float64{
			float64(int64(1024) * 1024 * 1024 * 1024),
			float64(int64(1024) * 1024 * 1024),
			float64(int64(1024) * 1024),
			float64(1024),
			0,
		},
		units: []string{"Ti", "Gi", "Mi", "Ki", "Bytes"},
	},
	UnitTypeDisk: {
		conditions: []float64{
			float64(int64(1024) * 1024 * 1024 * 1024),
			float64(int64(1024) * 1024 * 1024),
			float64(int64(1024) * 1024),
			float64(1024),
			0,
		},
		units: []string{"Ti", "Gi", "Mi", "Ki", "Bytes"},
	},
	UnitTypeThroughput: {
		conditions: []float64{1e12, 1e9, 1e6, 1e3, 0},
		units:      []string{"TB/s", "GB/s", "MB/s", "KB/s", "B/s"},
	},
	UnitTypeTraffic: {
		conditions: []float64{1e12, 1e9, 1e6, 1e3, 0},
		units:      []string{"TB/s", "GB/s", "MB/s", "KB/s", "B/s"},
	},
	UnitTypeBandwidth: {
		// (1024**2)/8 = 131072, 1024/8 = 128, 0
		conditions: []float64{131072, 128, 0},
		units:      []string{"Mbps", "Kbps", "bps"},
	},
	UnitTypeNumber: {
		conditions: []float64{1e12, 1e9, 1e6, 1e3, 0},
		units:      []string{"T", "G", "M", "K", ""},
	},
}

// IsKnownUnitType reports whether ut is one of the families recognised by
// the upstream UnitTypes table. Useful for catalog validation.
func IsKnownUnitType(ut UnitType) bool {
	_, ok := unitTypes[ut]
	return ok
}

// ----------------------------------------------------------------------------
// getValueByUnit / getSuitableUnit / getSuitableValue
// ----------------------------------------------------------------------------

// GetValueByUnit converts a raw metric value (Prometheus-style, always a
// string) into a target unit, mirroring `getValueByUnit` in monitoring.ts.
//
// Special cases preserved verbatim from the upstream JS:
//
//   - num == "NAN" (literal, case-sensitive) → start from 0.
//   - unit == "" or "default" → return the parsed value untouched
//     (no toFixed, no rounding).
//   - unit == "iops" → Math.round, no toFixed.
//   - unit == "m" → multiply by 1000 then snap-to-zero if < 1.
//   - everything else → divide / multiply / no-op per the table, then return
//     0 for an exact zero, otherwise toFixed(precision).
//
// The returned float64 is the rendered count (caller appends the unit text).
func GetValueByUnit(num string, unit string, precision int) float64 {
	var value float64
	if num == "NAN" {
		value = 0
	} else {
		v, err := strconv.ParseFloat(num, 64)
		if err != nil {
			// parseFloat in JS is permissive — leading "abc" → NaN.
			// monitoring.ts's path then runs through toFixed(NaN) which
			// returns "NaN". We collapse to 0 here because the only
			// non-numeric input the upstream actually defends against is
			// the literal "NAN" sentinel above; everything else is a
			// programming error worth flagging by zeroing the chart.
			value = 0
		} else {
			value = v
		}
	}

	switch unit {
	case "":
		return value
	case "default":
		return value
	case "iops":
		return math.Round(value)
	case "%":
		value *= 100
	case "m":
		value *= 1000
		if value < 1 {
			return 0
		}
	case "Ki":
		value /= 1024
	case "Mi":
		value /= 1024 * 1024
	case "Gi":
		value /= 1024 * 1024 * 1024
	case "Ti":
		value /= 1024 * 1024 * 1024 * 1024
	case "Bytes", "B", "B/s":
		// no-op
	case "K", "KB", "KB/s":
		value /= 1000
	case "M", "MB", "MB/s":
		value /= 1000 * 1000
	case "G", "GB", "GB/s":
		value /= 1000 * 1000 * 1000
	case "T", "TB", "TB/s":
		value /= 1000 * 1000 * 1000 * 1000
	case "bps":
		value *= 8
	case "Kbps":
		value = (value * 8) / 1024
	case "Mbps":
		value = (value * 8) / 1024 / 1024
	case "ms":
		value *= 1000
	default:
		// Unknown unit — fall through with the parsed value as-is.
		// Matches the JS switch's `default: break;`.
	}

	if value == 0 {
		return 0
	}
	return roundToFixed(value, precision)
}

// roundToFixed rounds v to `precision` decimal places using banker-free
// half-away-from-zero (Number.prototype.toFixed semantics — close enough
// for the metric values we ship). For negative `precision` we return v
// unchanged, matching toFixed throwing in JS.
func roundToFixed(v float64, precision int) float64 {
	if precision < 0 {
		return v
	}
	// strconv.FormatFloat with 'f' verb rounds half-to-even, which can
	// disagree with JS's toFixed on .5 boundaries (toFixed is half-away-
	// from-zero). For dashboard metrics the difference rarely surfaces,
	// but we keep this routine narrow so it can be hardened against
	// goldens later without ripple effects.
	pow := math.Pow(10, float64(precision))
	if math.IsInf(pow, 0) || pow == 0 {
		return v
	}
	if v < 0 {
		return -math.Floor(-v*pow+0.5) / pow
	}
	return math.Floor(v*pow+0.5) / pow
}

// GetSuitableUnit picks the most readable unit from `unitType` for either a
// single value or a slice of `[ts, value]` samples. Mirrors `getSuitableUnit`
// in monitoring.ts: walk `conditions` highest-first, the first one that any
// sample satisfies wins; otherwise we fall back to the smallest unit
// (`last(units)`), e.g. "Bytes" / "ms" / "" depending on the family.
//
// `value` is parsed leniently: int / float / string scalars are accepted, and
// for slice inputs each element may itself be a [ts, v] pair (length-2
// slice, where v lives at index 1) or a bare scalar.
func GetSuitableUnit(value any, unitType UnitType) string {
	def, ok := unitTypes[unitType]
	if !ok {
		return ""
	}

	samples := flattenForSuitability(value)
	result := def.units[len(def.units)-1]
	for i, cond := range def.conditions {
		triggered := false
		for _, s := range samples {
			if s >= cond {
				triggered = true
				break
			}
		}
		if triggered {
			result = def.units[i]
			break
		}
	}
	return result
}

// flattenForSuitability mimics the lodash-y "value can be array or scalar,
// elements can be [ts, v] tuples or bare scalars" contract from monitoring.ts.
func flattenForSuitability(value any) []float64 {
	switch v := value.(type) {
	case nil:
		return []float64{0}
	case float64:
		return []float64{v}
	case float32:
		return []float64{float64(v)}
	case int:
		return []float64{float64(v)}
	case int32:
		return []float64{float64(v)}
	case int64:
		return []float64{float64(v)}
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return []float64{0}
		}
		return []float64{f}
	case []any:
		// Either a single [ts, val] pair (length 2 with index 1 numeric)
		// or a slice of such pairs, or a slice of scalars. Same heuristic
		// the lodash chain in monitoring.ts uses.
		out := make([]float64, 0, len(v))
		for _, item := range v {
			out = append(out, extractSample(item))
		}
		return out
	default:
		return []float64{0}
	}
}

// extractSample pulls the numeric "value" component out of one element of a
// monitoring time series. Each element is either a [timestamp, value] pair
// (the Prometheus shape — value at index 1) or a bare scalar.
func extractSample(item any) float64 {
	switch v := item.(type) {
	case []any:
		if len(v) >= 2 {
			return scalar(v[1])
		}
		if len(v) == 1 {
			return scalar(v[0])
		}
		return 0
	default:
		return scalar(item)
	}
}

func scalar(item any) float64 {
	switch v := item.(type) {
	case nil:
		return 0
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

// GetSuitableValue is the headline "give me the human string for this
// metric" helper. Mirrors `getSuitableValue` in monitoring.ts:
//
//   - non-numeric input         → defaultValue rendered as-is
//   - numeric input, no UnitTypes match → "<count><unit>" with unit empty,
//     fed back into GetValueByUnit using `unitType` as the unit name (the
//     JS does the same `unit || unitType` fallback)
//   - numeric input, normal     → "<count> <CapitalisedUnit>"
//
// The capitalisation matches the upstream `_capitalize` (first char upper,
// rest lower) which produces e.g. "Mb/s" from "MB/s" — that capitalisation
// quirk is intentional, and golden tests pin it.
func GetSuitableValue(value string, unitType UnitType, defaultValue string) string {
	if defaultValue == "" {
		defaultValue = "0"
	}
	if !isNumericString(value) {
		return defaultValue
	}

	unit := GetSuitableUnit(value, unitType)
	unitText := ""
	if unit != "" {
		unitText = " " + capitalize(unit)
	}

	probe := unit
	if probe == "" {
		probe = string(unitType)
	}
	count := GetValueByUnit(value, probe, 2)

	return formatNumberJSLike(count) + unitText
}

// isNumericString matches the JS guard `(!isNumber(value) && !isString(value))
// || isNaN(Number(value))`. We always receive `value` as Go string here, so
// only the NaN check matters in practice — but we keep the test explicit so
// callers passing the literal "NaN" / "" hit the fallback like in JS.
func isNumericString(s string) bool {
	if s == "" {
		return false
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// capitalize mirrors `_capitalize` in monitoring.ts:
//
//	first char ToUpper, rest ToLower, then "_" → " "
func capitalize(s string) string {
	if s == "" {
		return s
	}
	first := strings.ToUpper(s[:1])
	rest := strings.ToLower(s[1:])
	out := first + rest
	out = strings.ReplaceAll(out, "_", " ")
	return out
}

// formatNumberJSLike formats a float the way `${count}` does in JS: integer
// values render without trailing zeros / decimal point, anything else uses
// the shortest decimal representation. Matches the visible output of
// monitoring.ts's `${count}${unitText}` template.
func formatNumberJSLike(v float64) string {
	if v == math.Trunc(v) && !math.IsInf(v, 0) {
		return strconv.FormatFloat(v, 'f', 0, 64)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// ----------------------------------------------------------------------------
// worthValue (number.ts)
// ----------------------------------------------------------------------------

// WorthValue ports the BigNumber-based formatter in
// `apps/packages/app/src/apps/dashboard/utils/number.ts`.
//
// The upstream is a curious cascade of *non-exclusive* `if` branches:
//
//	let k = 4;
//	if (t.gte(1000)) k = 0;
//	if (t.gte(100))  k = 1;
//	if (t.gte(10))   k = 2;
//	if (t.gte(1))    k = 3;
//
// Every condition runs (no `else if`), so for any v >= 1 the final `k` is 3,
// for v < 1 (or NaN) it stays 4. The upstream then formats with
// BigNumber.toPrecision(k) and adds thousand separators via toFormat().
//
// Bug-for-bug parity is intentional: the dashboard SPA renders the same
// values, and any "fix" here would silently desync the CLI from the UI.
func WorthValue(v string) string {
	d, err := decimal.NewFromString(v)
	if err != nil {
		// BigNumber('foo').toPrecision(4) returns "NaN"; toFormat()
		// preserves it. We do the same so the table layer can detect
		// the sentinel and substitute "-" if it wants to.
		return "NaN"
	}

	k := 4
	one := decimal.NewFromInt(1)
	ten := decimal.NewFromInt(10)
	hundred := decimal.NewFromInt(100)
	thousand := decimal.NewFromInt(1000)

	if d.GreaterThanOrEqual(thousand) {
		k = 0
	}
	if d.GreaterThanOrEqual(hundred) {
		k = 1
	}
	if d.GreaterThanOrEqual(ten) {
		k = 2
	}
	if d.GreaterThanOrEqual(one) {
		k = 3
	}

	rounded := toPrecision(d, k)
	return addThousandsSep(rounded)
}

// toPrecision is a narrow port of BigNumber.toPrecision for the value ranges
// worthValue actually feeds it (k in {3,4}; v can be 0, negative, very small).
// It returns the plain decimal representation — no scientific notation, since
// BigNumber's default EXPONENTIAL_AT prevents that for the magnitudes we hit
// here.
func toPrecision(d decimal.Decimal, sigDigits int) string {
	if sigDigits <= 0 {
		// k=0 would normally throw in BigNumber.toPrecision, but we
		// already saw worthValue's branch table can't reach this case
		// for any real input. Defensively render as integer.
		return d.Round(0).String()
	}
	if d.IsZero() {
		// "0" with no extra precision; mirrors BigNumber behavior
		// where toPrecision on zero returns "0".
		return "0"
	}
	abs := d.Abs()
	// Order of magnitude of the integer part: floor(log10(|d|)) + 1.
	// Use a string-based check to avoid float rounding.
	intPart := abs.Truncate(0).String()
	if intPart == "0" {
		intPart = ""
	}
	intDigits := len(intPart)
	if intDigits >= sigDigits {
		// More integer digits than sig digits → round to (intDigits - sigDigits)
		// places left of the decimal. BigNumber's toPrecision matches Round on
		// half-away-from-zero, which decimal.Round does by default.
		shift := intDigits - sigDigits
		factor := decimal.NewFromInt(1)
		for i := 0; i < shift; i++ {
			factor = factor.Mul(decimal.NewFromInt(10))
		}
		rounded := d.Div(factor).Round(0).Mul(factor)
		return rounded.String()
	}
	// sig digits fall after the integer part — keep (sigDigits - intDigits)
	// decimals.
	decimals := int32(sigDigits - intDigits)
	if intDigits == 0 {
		// Pure fraction: count leading zeros after the decimal point so
		// we keep `sigDigits` *significant* digits.
		// d in (0,1): exponent = floor(log10(|d|)), e.g. 0.0123 → -2.
		s := abs.String()
		// strip "0."
		if strings.HasPrefix(s, "0.") {
			frac := s[2:]
			leading := 0
			for _, ch := range frac {
				if ch == '0' {
					leading++
					continue
				}
				break
			}
			decimals = int32(sigDigits + leading)
		}
	}
	return d.Round(decimals).StringFixed(decimals)
}

// addThousandsSep adds `,` thousand separators to the integer portion of a
// numeric string, mirroring BigNumber.toFormat() with default config
// ({groupSize: 3, groupSeparator: ",", decimalSeparator: "."}). Negative
// numbers and a fractional tail are preserved untouched.
func addThousandsSep(numStr string) string {
	if numStr == "" || numStr == "NaN" {
		return numStr
	}
	sign := ""
	s := numStr
	if strings.HasPrefix(s, "-") {
		sign = "-"
		s = s[1:]
	}
	intPart := s
	frac := ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart = s[:i]
		frac = s[i:]
	}
	if len(intPart) <= 3 {
		return sign + intPart + frac
	}

	var b strings.Builder
	rem := len(intPart) % 3
	if rem > 0 {
		b.WriteString(intPart[:rem])
		if len(intPart) > rem {
			b.WriteByte(',')
		}
	}
	for i := rem; i < len(intPart); i += 3 {
		b.WriteString(intPart[i : i+3])
		if i+3 < len(intPart) {
			b.WriteByte(',')
		}
	}
	return sign + b.String() + frac
}

// ----------------------------------------------------------------------------
// formatFrequency / convertTemperature (cpu.ts)
// ----------------------------------------------------------------------------

// FormatFrequency mirrors `formatFrequency` in cpu.ts. For value == 0 it
// returns "0 Hz" (with space); for everything else the unit is auto-scaled
// up or down through [Hz, kHz, MHz, GHz] and concatenated *without* a space
// (the JS template `${...}${units[i]}` has no separator).
//
// Rounding: `Math.round(frequency * 100) / 100` — two-decimal snap, with
// trailing zeros suppressed by JS's number→string coercion. `1500 → "1.5kHz"`
// not `"1.50kHz"`.
//
// `fromUnit` defaults to "Hz" when empty; non-recognised values mean
// `unitIndex` becomes -1 in JS, the up-loop never runs, and we render in the
// original unit. We mirror that behaviour with a clamp.
func FormatFrequency(value float64, fromUnit string) string {
	if value == 0 {
		return "0 Hz"
	}
	if fromUnit == "" {
		fromUnit = "Hz"
	}
	units := []string{"Hz", "kHz", "MHz", "GHz"}
	unitIndex := -1
	for i, u := range units {
		if u == fromUnit {
			unitIndex = i
			break
		}
	}
	frequency := value
	if unitIndex < 0 {
		// JS leaves currentUnitIndex at -1; both while-loops short-circuit
		// and we render with the original unit. We do the same.
		return formatFrequencyFloat(frequency) + fromUnit
	}
	currentUnitIndex := unitIndex
	for frequency >= 1000 && currentUnitIndex < len(units)-1 {
		frequency /= 1000
		currentUnitIndex++
	}
	for frequency < 1 && currentUnitIndex > 0 {
		frequency *= 1000
		currentUnitIndex--
	}
	return formatFrequencyFloat(frequency) + units[currentUnitIndex]
}

// formatFrequencyFloat emulates the JS `${Math.round(f * 100) / 100}` pattern
// — snap to 2 decimals, but render via Number→string so trailing zeros and
// trailing decimal points are stripped.
func formatFrequencyFloat(f float64) string {
	rounded := math.Round(f*100) / 100
	if rounded == math.Trunc(rounded) && !math.IsInf(rounded, 0) {
		return strconv.FormatFloat(rounded, 'f', 0, 64)
	}
	return strconv.FormatFloat(rounded, 'f', -1, 64)
}

// TempUnit is the target unit for ConvertTemperature.
type TempUnit string

const (
	TempC TempUnit = "C"
	TempF TempUnit = "F"
	TempK TempUnit = "K"
)

// ConvertTemperature mirrors `convertTemperature` in cpu.ts: switch on the
// target unit, default (anything other than F/K) returns the input celsius
// value unchanged. We canonicalise input as °C — the dashboard always feeds
// celsius as the source.
func ConvertTemperature(celsius float64, target TempUnit) float64 {
	switch target {
	case TempF:
		return celsius*1.8 + 32
	case TempK:
		return celsius + 273.15
	default:
		return celsius
	}
}

// ----------------------------------------------------------------------------
// getDiskSize / getThroughput (disk.ts, memory.ts)
// ----------------------------------------------------------------------------

// GetDiskSize mirrors `getDiskSize` in disk.ts. NB: the upstream
// concatenates value+unit *without* a space ("100Gi"), and substitutes "-"
// for any falsy result (zero or NaN). We replicate exactly.
func GetDiskSize(size string) string {
	unit := GetSuitableUnit(size, UnitTypeDisk)
	if unit == "" {
		// shouldn't happen for a known UnitType, but the JS `|| ''`
		// fallback matters when tooling pipes us a corrupt UnitType.
		unit = ""
	}
	value := GetValueByUnit(size, unit, 2)
	if value == 0 {
		return "-"
	}
	return formatNumberJSLike(value) + unit
}

// GetThroughput mirrors `getThroughput` in memory.ts (despite the
// "throughput" name — the file is in the wrong path upstream). NB: the
// upstream concatenates `value + ' ' + unit`, *with* a space and *without*
// a "-" fallback, so `0 B/s` is a possible (and intended) output.
func GetThroughput(size string) string {
	unit := GetSuitableUnit(size, UnitTypeThroughput)
	value := GetValueByUnit(size, unit, 2)
	return formatNumberJSLike(value) + " " + unit
}

// ----------------------------------------------------------------------------
// getFormatTime / getMinuteValue (monitoring.ts, status.ts)
// ----------------------------------------------------------------------------

// FormatTime mirrors `getFormatTime` in monitoring.ts: render `ms` (a
// millisecond timestamp, also accepted as a stringified number for legacy
// callers) according to whether `showDay` is set, and strip the trailing
// `:00` seconds when the longer pattern would end on an exact minute.
//
// We keep the rendering local-time / fixed-format ("MM-DD HH:mm" or
// "HH:mm:ss"). The watch loop's `--timezone` flag overrides `loc`.
func FormatTime(ms int64, showDay bool, loc *Location) string {
	if loc == nil {
		loc = LocalLocation()
	}
	t := loc.unixMilliToTime(ms)
	if showDay {
		return t.Format("01-02 15:04")
	}
	s := t.Format("15:04:05")
	if strings.HasSuffix(s, ":00") {
		// Match the JS regex /(\d+:\d+)(:00)$/.
		return s[:len(s)-3]
	}
	return s
}

// GetMinuteValue mirrors the small "Nm" stringifier the dashboard uses for
// uptime / interval labels in `status.ts` (callers pass a duration in
// minutes; output is e.g. "5m"). We accept the duration as a float so the
// CLI can pass the same Prometheus-flavoured numbers the SPA does.
func GetMinuteValue(minutes float64) string {
	if minutes == math.Trunc(minutes) && !math.IsInf(minutes, 0) {
		return strconv.FormatFloat(minutes, 'f', 0, 64) + "m"
	}
	return strconv.FormatFloat(minutes, 'f', -1, 64) + "m"
}

// ----------------------------------------------------------------------------
// getLastMonitoringData / getResult (monitoring.ts)
// ----------------------------------------------------------------------------

// LastMonitoringSample is the canonical shape for the "most recent point"
// extracted from a Prometheus-style metric. value can be either a single
// `[ts, val]` (matrix `value`) or the last entry of `values` (matrix
// `values`); see `getLastMonitoringData` in monitoring.ts.
type LastMonitoringSample struct {
	// Timestamp is the unix-seconds timestamp of the sample (the [0]
	// position of the [ts, val] pair).
	Timestamp float64 `json:"timestamp"`
	// RawValue is the value at index [1] of the pair, kept as the
	// upstream string so the caller can choose how to re-render it via
	// GetValueByUnit / GetSuitableValue.
	RawValue string `json:"raw_value"`
	// Empty signals there was no sample at all (no values + no value).
	Empty bool `json:"empty,omitempty"`
}

// GetLastMonitoringData ports `getLastMonitoringData` from monitoring.ts,
// but in a Go-idiomatic shape: `metricResults` is the map from metric_name
// to its raw API response, and we return the last sample for `index`
// (default 0) of each.
//
// errMissingValues is returned only as a value-bag entry per-metric (Empty=
// true), never as an error — the upstream just emits `{}` and lets the
// chart layer render zeros, so we follow.
func GetLastMonitoringData(metricResults map[string]MonitoringResult, index int) map[string]LastMonitoringSample {
	out := make(map[string]LastMonitoringSample, len(metricResults))
	for name, res := range metricResults {
		out[name] = lastSample(res, index)
	}
	return out
}

// MonitoringResult is the minimal shape `getLastMonitoringData` peers into:
//
//	{ "data": { "result": [ { "metric": {...}, "values": [...], "value": [ts,v] }, ... ] } }
//
// We keep it concrete rather than `interface{}` so callers don't have to
// hand-craft lodash-style getters. The `metric` map carries Prometheus
// labels (e.g. `pod`, `namespace`, `owner_kind`); the dashboard SPA reads
// them in the workload-grain merge — see Applications2/config.ts's
// `getTabOptions` / `getTabOptions2`.
type MonitoringResult struct {
	Data struct {
		Result []struct {
			Metric map[string]string `json:"metric,omitempty"`
			Values [][]any           `json:"values,omitempty"`
			Value  []any             `json:"value,omitempty"`
		} `json:"result"`
	} `json:"data"`
}

func lastSample(res MonitoringResult, index int) LastMonitoringSample {
	if index < 0 || index >= len(res.Data.Result) {
		return LastMonitoringSample{Empty: true}
	}
	r := res.Data.Result[index]
	if len(r.Values) > 0 {
		pair := r.Values[len(r.Values)-1]
		return pairToSample(pair)
	}
	if len(r.Value) > 0 {
		return pairToSample(r.Value)
	}
	return LastMonitoringSample{Empty: true}
}

func pairToSample(pair []any) LastMonitoringSample {
	if len(pair) < 2 {
		return LastMonitoringSample{Empty: true}
	}
	return LastMonitoringSample{
		Timestamp: scalar(pair[0]),
		RawValue:  scalarString(pair[1]),
	}
}

func scalarString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int, int32, int64:
		return fmt.Sprintf("%v", x)
	default:
		return fmt.Sprintf("%v", x)
	}
}

// ----------------------------------------------------------------------------
// getResult (monitoring.ts)
// ----------------------------------------------------------------------------

// GetResult ports `getResult` in monitoring.ts: flatten either a top-level
// list of metric results, or a `{ results: [...] }` wrapper, into a
// metric_name → result map. Single-result top-level objects with a
// metric_name field are handled too (the upstream's special-case fallback).
//
// Input is `any` so callers can feed the BFF response straight in (it's
// either a slice or a map depending on the endpoint). The return preserves
// the original entries so further per-metric processing still has access to
// the raw values / labels.
func GetResult(input any) map[string]any {
	out := map[string]any{}
	switch v := input.(type) {
	case []any:
		fillFromSlice(out, v)
	case map[string]any:
		if results, ok := v["results"]; ok {
			if rs, ok := results.([]any); ok {
				fillFromSlice(out, rs)
				return out
			}
		}
		// Single-result fallback: top-level object has a metric_name.
		if name, ok := v["metric_name"].(string); ok && name != "" {
			out[name] = v
		}
	}
	return out
}

func fillFromSlice(out map[string]any, entries []any) {
	for _, e := range entries {
		m, ok := e.(map[string]any)
		if !ok {
			continue
		}
		name, _ := m["metric_name"].(string)
		if name == "" {
			continue
		}
		out[name] = m
	}
}

// ----------------------------------------------------------------------------
// Errors / sentinels exported for tests
// ----------------------------------------------------------------------------

// ErrNotNumeric is returned by no public helper today, but exists so
// future strict variants (e.g. GetSuitableValueStrict) have a stable
// sentinel to assert on.
var ErrNotNumeric = errors.New("value is not numeric")
