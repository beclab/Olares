// Package format tests pin the Go port to the upstream JS source. Every
// case here either:
//
//   - mirrors a literal value computed by hand from monitoring.ts (the
//     "manual" cases below), or
//   - is sourced from cli/cmd/ctl/dashboard/format/testdata/golden.json,
//     which is regenerated via testdata/golden-gen.js using
//     @bytetrade/core/src/monitoring.ts directly.
//
// When you tweak format.go, regenerate golden.json:
//
//	cd cli/cmd/ctl/dashboard/format/testdata && node golden-gen.js
//
// and re-run `go test ./cmd/ctl/dashboard/format/...`.

package format

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestGetValueByUnit_KnownUnits(t *testing.T) {
	cases := []struct {
		name      string
		num       string
		unit      string
		precision int
		want      float64
	}{
		{"empty unit returns parsed value", "12.345", "", 2, 12.345},
		{"default unit returns parsed value", "12.345", "default", 2, 12.345},
		{"NAN sentinel collapses to zero", "NAN", "Bytes", 2, 0},
		{"non-numeric collapses to zero", "abc", "Bytes", 2, 0},
		{"percent multiplies by 100", "0.4567", "%", 2, 45.67},
		{"millicore multiplies by 1000", "0.123", "m", 2, 123},
		{"millicore < 1 snaps to zero", "0.0001", "m", 2, 0},
		{"Ki divides by 1024", "2048", "Ki", 2, 2},
		{"Mi divides by 1024^2", "1048576", "Mi", 2, 1},
		{"Gi divides by 1024^3", "1073741824", "Gi", 2, 1},
		{"Ti divides by 1024^4", "1099511627776", "Ti", 2, 1},
		{"Bytes is no-op", "1024", "Bytes", 2, 1024},
		{"K divides by 1000", "1500", "K", 2, 1.5},
		{"M divides by 1000^2", "1500000", "M", 2, 1.5},
		{"G divides by 1000^3", "1500000000", "G", 2, 1.5},
		{"T divides by 1000^4", "1500000000000", "T", 2, 1.5},
		{"bps multiplies by 8", "100", "bps", 2, 800},
		{"Kbps mul8 div1024", "1024", "Kbps", 2, 8},
		{"Mbps mul8 div(1024^2)", "131072", "Mbps", 2, 1},
		{"ms multiplies by 1000", "0.1234", "ms", 2, 123.4},
		{"iops rounds to integer (no toFixed)", "12.7", "iops", 2, 13},
		{"unknown unit returns parsed value", "0.45", "fishtacos", 2, 0.45},
		{"exact zero short-circuits to 0", "0", "Mi", 2, 0},
		// 1.234567 → 1.23 (toFixed(2)). NB: don't use 1.005 here —
		// it lives in IEEE-754 as 1.00499... so JS's toFixed *also*
		// returns "1.00", not "1.01". Pin a value where the rounding
		// is unambiguous in either floating-point representation.
		{"precision applies via toFixed", "1.234567", "Bytes", 2, 1.23},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := GetValueByUnit(tc.num, tc.unit, tc.precision)
			if got != tc.want {
				t.Fatalf("GetValueByUnit(%q, %q, %d) = %v, want %v", tc.num, tc.unit, tc.precision, got, tc.want)
			}
		})
	}
}

func TestGetSuitableUnit_PicksHighestSatisfiedCondition(t *testing.T) {
	cases := []struct {
		name string
		v    any
		ut   UnitType
		want string
	}{
		{"memory 0 falls back to smallest unit", 0, UnitTypeMemory, "Bytes"},
		{"memory 1KiB picks Ki", 1024, UnitTypeMemory, "Ki"},
		{"memory 1MiB picks Mi", 1024 * 1024, UnitTypeMemory, "Mi"},
		{"memory 1GiB picks Gi", float64(1024 * 1024 * 1024), UnitTypeMemory, "Gi"},
		{"memory 1TiB picks Ti", float64(int64(1024) * 1024 * 1024 * 1024), UnitTypeMemory, "Ti"},
		{"throughput 1MB/s", 1_000_000, UnitTypeThroughput, "MB/s"},
		{"bandwidth 1Kbps", 128, UnitTypeBandwidth, "Kbps"},
		{"bandwidth 1Mbps", 131072, UnitTypeBandwidth, "Mbps"},
		{"second 0.005 = ms", 0.005, UnitTypeSecond, "ms"},
		{"second 1 = s", 1, UnitTypeSecond, "s"},
		{"unknown unit type returns empty", 1, UnitType("nonsense"), ""},
		{"slice picks max sample", []any{0, 1024, 0}, UnitTypeMemory, "Ki"},
		{"timeseries pair extracts index 1", []any{[]any{1700000000, 2048}}, UnitTypeMemory, "Ki"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := GetSuitableUnit(tc.v, tc.ut)
			if got != tc.want {
				t.Fatalf("GetSuitableUnit(%v, %s) = %q, want %q", tc.v, tc.ut, got, tc.want)
			}
		})
	}
}

func TestGetSuitableValue_RendersHumanString(t *testing.T) {
	cases := []struct {
		name string
		v    string
		ut   UnitType
		want string
	}{
		// 1.5 GiB rounded to 1.5 since precision=2.
		{"memory 1.5GiB", strconv.FormatFloat(1.5*1024*1024*1024, 'f', -1, 64), UnitTypeMemory, "1.5 Gi"},
		// 1024 bytes → 1 Ki (formatNumberJSLike strips ".00")
		{"memory 1KiB", "1024", UnitTypeMemory, "1 Ki"},
		// Sub-1024 → Bytes
		{"memory 100B", "100", UnitTypeMemory, "100 Bytes"},
		// CPU 0.05 -> 50m (under 0.1 threshold)
		{"cpu 0.05 core → 50m", "0.05", UnitTypeCPU, "50 M"},
		// Throughput 1500 → 1.5 KB/s
		{"throughput 1500", "1500", UnitTypeThroughput, "1.5 Kb/s"},
		// Non-numeric → defaultValue (0).
		{"non-numeric falls back to 0", "totally not a number", UnitTypeMemory, "0"},
		// Empty unit means we use the unitType name as the unit, fed
		// through GetValueByUnit's default branch — value stays.
		{"number 5 falls back to lowest unit ''", "5", UnitTypeNumber, "5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := GetSuitableValue(tc.v, tc.ut, "0")
			if got != tc.want {
				t.Fatalf("GetSuitableValue(%q, %s) = %q, want %q", tc.v, tc.ut, got, tc.want)
			}
		})
	}
}

func TestWorthValue_BugForBugWithBigNumberCascade(t *testing.T) {
	// The upstream cascade always lands at k=3 for v >= 1, k=4 for v < 1
	// (or negatives / NaN). We pin both branches plus the formatting
	// quirks (thousand separator + sign).
	cases := []struct {
		name string
		v    string
		want string
	}{
		{"v >= 1: 1234.5678 -> 3 sig digits with thousand sep", "1234.5678", "1,230"},
		{"v >= 1: 1.23456 -> 3 sig digits", "1.23456", "1.23"},
		{"v < 1: 0.001234 -> 4 sig digits", "0.001234", "0.001234"},
		{"v < 1: 0.5 -> 4 sig digits", "0.5", "0.5000"},
		{"zero -> '0'", "0", "0"},
		{"negative: -1234 -> 4 sig digits (cascade misses for negatives)", "-1234", "-1,234"},
		{"non-numeric returns NaN sentinel", "abc", "NaN"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := WorthValue(tc.v)
			if got != tc.want {
				t.Fatalf("WorthValue(%q) = %q, want %q", tc.v, got, tc.want)
			}
		})
	}
}

func TestFormatFrequency_AutoScalesAndRounds(t *testing.T) {
	cases := []struct {
		name string
		v    float64
		from string
		want string
	}{
		{"zero special case", 0, "Hz", "0 Hz"},
		{"1500 Hz -> 1.5kHz", 1500, "Hz", "1.5kHz"},
		{"3500000000 Hz -> 3.5GHz", 3.5e9, "Hz", "3.5GHz"},
		{"0.5 kHz -> 500Hz", 0.5, "kHz", "500Hz"},
		{"empty fromUnit defaults to Hz", 1500, "", "1.5kHz"},
		{"unknown fromUnit short-circuits", 100, "blub", "100blub"},
		{"trailing zeros stripped (no '.50')", 2_500, "Hz", "2.5kHz"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatFrequency(tc.v, tc.from)
			if got != tc.want {
				t.Fatalf("FormatFrequency(%v, %q) = %q, want %q", tc.v, tc.from, got, tc.want)
			}
		})
	}
}

func TestConvertTemperature(t *testing.T) {
	if got := ConvertTemperature(0, TempC); got != 0 {
		t.Fatalf("ConvertTemperature(0, C) = %v, want 0", got)
	}
	if got := ConvertTemperature(100, TempC); got != 100 {
		t.Fatalf("ConvertTemperature(100, C) = %v, want 100", got)
	}
	if got := ConvertTemperature(0, TempF); got != 32 {
		t.Fatalf("ConvertTemperature(0, F) = %v, want 32", got)
	}
	if got := ConvertTemperature(100, TempF); got != 212 {
		t.Fatalf("ConvertTemperature(100, F) = %v, want 212", got)
	}
	if got := ConvertTemperature(0, TempK); got != 273.15 {
		t.Fatalf("ConvertTemperature(0, K) = %v, want 273.15", got)
	}
	// Default branch (unknown target) returns C unchanged — matches
	// monitoring.ts's `default: return celsius;`.
	if got := ConvertTemperature(42, TempUnit("X")); got != 42 {
		t.Fatalf("ConvertTemperature(42, X) = %v, want 42", got)
	}
}

func TestGetDiskSize_ConcatsValuePlusUnitNoSpace(t *testing.T) {
	cases := []struct {
		name string
		size string
		want string
	}{
		{"1024 -> 1Ki", "1024", "1Ki"},
		{"1.5GiB literal", strconv.FormatFloat(1.5*1024*1024*1024, 'f', -1, 64), "1.5Gi"},
		{"zero -> '-'", "0", "-"},
		{"NaN -> '-'", "abc", "-"},
		{"Bytes-range value", "100", "100Bytes"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := GetDiskSize(tc.size)
			if got != tc.want {
				t.Fatalf("GetDiskSize(%q) = %q, want %q", tc.size, got, tc.want)
			}
		})
	}
}

func TestGetThroughput_ConcatsWithSpace(t *testing.T) {
	if got := GetThroughput("1500"); got != "1.5 KB/s" {
		t.Fatalf("GetThroughput(1500) = %q, want \"1.5 KB/s\"", got)
	}
	if got := GetThroughput("0"); got != "0 B/s" {
		t.Fatalf("GetThroughput(0) = %q, want \"0 B/s\"", got)
	}
}

func TestGetMinuteValue_AppendsM(t *testing.T) {
	if got := GetMinuteValue(5); got != "5m" {
		t.Fatalf("GetMinuteValue(5) = %q, want 5m", got)
	}
	if got := GetMinuteValue(1.5); got != "1.5m" {
		t.Fatalf("GetMinuteValue(1.5) = %q, want 1.5m", got)
	}
}

// TestFormat_GoldenOracle is a regression layer driven by the JS oracle
// (testdata/golden-gen.js) when present. The fixture is committed; CI
// asserts it stays in sync. The "testdata/golden.json" file is optional
// — when missing, the test is skipped (so contributors without Node can
// still run `go test`). When present, every case must match exactly.
func TestFormat_GoldenOracle(t *testing.T) {
	path := filepath.Join("testdata", "golden.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("testdata/golden.json missing; run `node testdata/golden-gen.js` to regenerate")
		}
		t.Fatalf("read golden fixture: %v", err)
	}
	var fixtures struct {
		GetValueByUnit []struct {
			Num       string  `json:"num"`
			Unit      string  `json:"unit"`
			Precision int     `json:"precision"`
			Want      float64 `json:"want"`
		} `json:"getValueByUnit"`
		GetSuitableValue []struct {
			Value    string `json:"value"`
			UnitType string `json:"unitType"`
			Want     string `json:"want"`
		} `json:"getSuitableValue"`
		WorthValue []struct {
			Input string `json:"input"`
			Want  string `json:"want"`
		} `json:"worthValue"`
		FormatFrequency []struct {
			Value float64 `json:"value"`
			From  string  `json:"from"`
			Want  string  `json:"want"`
		} `json:"formatFrequency"`
	}
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatalf("decode golden fixture: %v", err)
	}

	for _, c := range fixtures.GetValueByUnit {
		got := GetValueByUnit(c.Num, c.Unit, c.Precision)
		if got != c.Want {
			t.Errorf("getValueByUnit(%q, %q, %d) = %v, want %v", c.Num, c.Unit, c.Precision, got, c.Want)
		}
	}
	for _, c := range fixtures.GetSuitableValue {
		got := GetSuitableValue(c.Value, UnitType(c.UnitType), "0")
		if got != c.Want {
			t.Errorf("getSuitableValue(%q, %q) = %q, want %q", c.Value, c.UnitType, got, c.Want)
		}
	}
	for _, c := range fixtures.WorthValue {
		got := WorthValue(c.Input)
		if got != c.Want {
			t.Errorf("worthValue(%q) = %q, want %q", c.Input, got, c.Want)
		}
	}
	for _, c := range fixtures.FormatFrequency {
		got := FormatFrequency(c.Value, c.From)
		if got != c.Want {
			t.Errorf("formatFrequency(%v, %q) = %q, want %q", c.Value, c.From, got, c.Want)
		}
	}
}
