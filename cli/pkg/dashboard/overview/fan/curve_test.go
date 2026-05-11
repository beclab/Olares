package fan

import (
	"strconv"
	"testing"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// TestBuildCurveEnvelope_TabletopMatchesUpstream pins the curve
// envelope as a 1:1 of pkgdashboard.FanCurveTable. SKILL.md's
// iteration red-line forbids the cmd / pkg sides from drifting; if
// the upstream table changes, this test must be updated alongside
// — never mutated independently.
func TestBuildCurveEnvelope_TabletopMatchesUpstream(t *testing.T) {
	cf := fixtureFlags(t)
	env := BuildCurveEnvelope(cf, "alice@olares.com", time.Now())

	if env.Kind != pkgdashboard.KindOverviewFanCurve {
		t.Errorf("Kind = %q, want %q", env.Kind, pkgdashboard.KindOverviewFanCurve)
	}
	if got, want := len(env.Items), len(pkgdashboard.FanCurveTable); got != want {
		t.Fatalf("Items len = %d, want %d (FanCurveTable size)", got, want)
	}
	for i, want := range pkgdashboard.FanCurveTable {
		got := env.Items[i]
		if got.Raw["step"] != want.Step {
			t.Errorf("row %d Raw.step = %v, want %d", i, got.Raw["step"], want.Step)
		}
		if got.Raw["cpu_fan_rpm"] != want.CPUFanRPM {
			t.Errorf("row %d Raw.cpu_fan_rpm = %v, want %d", i, got.Raw["cpu_fan_rpm"], want.CPUFanRPM)
		}
		// Display values are stringified — pin both shape and value.
		if got.Display["step"] != strconv.Itoa(want.Step) {
			t.Errorf("row %d Display.step = %v, want %q", i, got.Display["step"], strconv.Itoa(want.Step))
		}
		if got.Display["cpu_temp_range"] != want.CPUTempRange {
			t.Errorf("row %d cpu_temp_range = %v, want %q", i, got.Display["cpu_temp_range"], want.CPUTempRange)
		}
	}
}
