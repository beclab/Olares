package nodestatus

import (
	"context"
	"errors"
	"testing"
	"time"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

func strp(s string) *string { return &s }

func TestBuildCarriesIdentity(t *testing.T) {
	id := Identity{
		NodeName:   "master-1",
		Role:       inventory.RoleMaster,
		Hostname:   "olares-one",
		DeviceType: DeviceTypeOlaresOne,
	}
	st := clistate.State{TerminusState: clistate.TerminusRunning}

	got := Build(id, st, nil, time.Unix(1700000000, 0))

	if got.NodeName != "master-1" || got.Role != inventory.RoleMaster {
		t.Errorf("identity not carried through: %+v", got)
	}
	if got.Hostname != "olares-one" || got.DeviceType != DeviceTypeOlaresOne {
		t.Errorf("hostname/device type not carried through: %+v", got)
	}
	if got.ObservedAt == nil || !got.ObservedAt.Equal(time.Unix(1700000000, 0)) {
		t.Errorf("observedAt = %v, want the time the state was observed", got.ObservedAt)
	}
}

// observedAt describes when the state was last refreshed, not when this
// response was written: a stale answer that timestamps itself "now" is how a
// dead node keeps looking fresh.
func TestBuildWithoutAnObservationSaysSoInsteadOfUsingNow(t *testing.T) {
	got := Build(Identity{}, clistate.State{TerminusState: clistate.Checking}, nil, time.Time{})

	if got.ObservedAt != nil {
		t.Errorf("observedAt = %v, want null when the state has never been refreshed", got.ObservedAt)
	}
}

func TestBuildUnresolvedIdentityIsUnknownNotFabricated(t *testing.T) {
	got := Build(Identity{}, clistate.State{TerminusState: clistate.TerminusRunning}, nil, time.Now())

	if got.Role != inventory.RoleUnknown {
		t.Errorf("role = %q, want %q when this node cannot resolve its own role", got.Role, inventory.RoleUnknown)
	}
	if got.NodeName != "" {
		t.Errorf("nodeName = %q, want empty rather than a guess", got.NodeName)
	}
	if got.DeviceType != DeviceTypeUnknown {
		t.Errorf("deviceType = %q, want %q", got.DeviceType, DeviceTypeUnknown)
	}
}

func TestBuildKeepsRawTerminusState(t *testing.T) {
	for _, s := range []clistate.TerminusState{
		clistate.TerminusRunning,
		clistate.SystemError,
		clistate.Checking,
		clistate.TerminusState("a-state-this-build-never-heard-of"),
	} {
		got := Build(Identity{}, clistate.State{TerminusState: s}, nil, time.Now())
		if got.TerminusState != s {
			t.Errorf("terminusState = %q, want the raw %q", got.TerminusState, s)
		}
	}
}

func TestBuildSelfReportIsOnline(t *testing.T) {
	got := Build(Identity{}, clistate.State{TerminusState: clistate.SystemError}, nil, time.Now())

	if got.Connectivity != ConnectivityOnline {
		t.Errorf("connectivity = %q, want %q: a node answering this call is reachable", got.Connectivity, ConnectivityOnline)
	}
}

func TestBuildHealthAndPhaseAreIndependent(t *testing.T) {
	cases := []struct {
		state  clistate.TerminusState
		health Health
		phase  Phase
	}{
		{clistate.TerminusRunning, HealthHealthy, PhaseRunning},
		// Olares is not up on this node, so "running" would describe a system
		// that is not there.
		{clistate.NotInstalled, HealthUnknown, PhaseUnknown},
		{clistate.Uninitialized, HealthUnknown, PhaseUnknown},
		{clistate.SystemError, HealthDegraded, PhaseRunning},
		{clistate.NetworkNotReady, HealthDegraded, PhaseRunning},
		{clistate.InvalidIpAddress, HealthDegraded, PhaseRunning},
		{clistate.InstallFailed, HealthDegraded, PhaseRunning},
		{clistate.InitializeFailed, HealthDegraded, PhaseRunning},
		{clistate.IPChangeFailed, HealthDegraded, PhaseRunning},
		{clistate.Restarting, HealthUnknown, PhaseRestarting},
		{clistate.Shutdown, HealthUnknown, PhaseShuttingDown},
		{clistate.Upgrading, HealthUnknown, PhaseMaintenance},
		{clistate.Installing, HealthUnknown, PhaseMaintenance},
		{clistate.Initializing, HealthUnknown, PhaseMaintenance},
		{clistate.Uninstalling, HealthUnknown, PhaseMaintenance},
		{clistate.IPChanging, HealthUnknown, PhaseMaintenance},
		{clistate.DiskModifing, HealthUnknown, PhaseMaintenance},
		{clistate.AddingNode, HealthUnknown, PhaseMaintenance},
		{clistate.RemovingNode, HealthUnknown, PhaseMaintenance},
		{clistate.SelfRepairing, HealthUnknown, PhaseMaintenance},
		{clistate.Checking, HealthUnknown, PhaseUnknown},
		{clistate.TerminusState(""), HealthUnknown, PhaseUnknown},
		{clistate.TerminusState("brand-new-state"), HealthUnknown, PhaseUnknown},
	}

	for _, c := range cases {
		got := Build(Identity{}, clistate.State{TerminusState: c.state}, nil, time.Now())
		if got.Health != c.health {
			t.Errorf("%s: health = %q, want %q", c.state, got.Health, c.health)
		}
		if got.Phase != c.phase {
			t.Errorf("%s: phase = %q, want %q", c.state, got.Phase, c.phase)
		}
	}
}

func TestBuildRestartingIsNotUnhealthy(t *testing.T) {
	got := Build(Identity{}, clistate.State{TerminusState: clistate.Restarting}, nil, time.Now())

	if got.Health == HealthDegraded {
		t.Errorf("a restarting node must not be reported as degraded: %+v", got)
	}
}

func TestBuildReportsHardware(t *testing.T) {
	st := clistate.State{
		TerminusState: clistate.TerminusRunning,
		CpuInfo:       "NVIDIA Grace",
		GPUList:       []string{"NVIDIA GB10"},
	}

	got := Build(Identity{}, st, nil, time.Now())

	if got.CPU != "NVIDIA Grace" {
		t.Errorf("cpu = %q, want the reported CPU", got.CPU)
	}
	if len(got.GPUs) != 1 || got.GPUs[0] != "NVIDIA GB10" {
		t.Errorf("gpus = %+v, want the reported GPU list", got.GPUs)
	}
}

func TestBuildGPUsAreAlwaysAList(t *testing.T) {
	got := Build(Identity{}, clistate.State{TerminusState: clistate.TerminusRunning}, nil, time.Now())

	if got.GPUs == nil {
		t.Error("want an empty list rather than null when no GPU is detected")
	}
}

func TestBuildTurnsPressuresIntoConditions(t *testing.T) {
	st := clistate.State{
		TerminusState: clistate.TerminusRunning,
		Pressure: []clistate.NodePressure{
			{Type: "MemoryPressure", Message: "kubelet has insufficient memory"},
		},
	}

	got := Build(Identity{}, st, nil, time.Now())

	var found bool
	for _, c := range got.Conditions {
		if c.Type == "MemoryPressure" {
			found = true
			if !c.Status || c.Message == "" {
				t.Errorf("pressure condition lost detail: %+v", c)
			}
		}
	}
	if !found {
		t.Errorf("MemoryPressure not surfaced as a condition: %+v", got.Conditions)
	}
}

func TestBuildConditionsAreAlwaysAList(t *testing.T) {
	got := Build(Identity{}, clistate.State{TerminusState: clistate.TerminusRunning}, nil, time.Now())

	if got.Conditions == nil {
		t.Error("want an empty list rather than null when there is nothing to report")
	}
}

func TestBuildDegradedStateGetsACondition(t *testing.T) {
	got := Build(Identity{}, clistate.State{TerminusState: clistate.SystemError}, nil, time.Now())

	var found bool
	for _, c := range got.Conditions {
		if c.Type == ConditionOlaresStateAbnormal {
			found = true
			if !c.Status || c.Message == "" {
				t.Errorf("abnormal-state condition lost detail: %+v", c)
			}
		}
	}
	if !found {
		t.Errorf("a degraded node should say why: %+v", got.Conditions)
	}
}

func TestBuildCapabilitiesAreAlwaysAMap(t *testing.T) {
	got := Build(Identity{}, clistate.State{TerminusState: clistate.TerminusRunning}, nil, time.Now())

	if got.Capabilities == nil {
		t.Error("want an empty object rather than null so clients can iterate unconditionally")
	}
}

// withCommands makes the power commands resolvable, or not, without touching
// the machine the test runs on.
func withCommands(t *testing.T, available ...string) {
	t.Helper()
	set := map[string]bool{}
	for _, c := range available {
		set[c] = true
	}
	prev := lookPath
	lookPath = func(file string) (string, error) {
		if set[file] {
			return "/sbin/" + file, nil
		}
		return "", errors.New("executable file not found in $PATH")
	}
	t.Cleanup(func() { lookPath = prev })
}

func workerOn(st clistate.State) ProbeInput {
	return ProbeInput{Identity: Identity{Role: inventory.RoleWorker}, State: st}
}

func TestDetectDeclaresPowerOnAWorkerThatHasTheCommands(t *testing.T) {
	withCommands(t, "shutdown", "reboot", "olares-cli", "chpasswd")

	caps := Detect(context.Background(), workerOn(clistate.State{TerminusState: clistate.TerminusRunning}))

	for _, name := range []string{CapPowerShutdown, CapPowerReboot, CapLogsCollect, CapSetSSHPassword} {
		c, ok := caps[name]
		if !ok {
			t.Errorf("capability %q not declared", name)
			continue
		}
		if !c.Supported {
			t.Errorf("capability %q declared but unsupported", name)
		}
	}
}

// olaresd in a container would run shutdown against the container, not the
// machine. Offering the button there promises a power-off that never happens.
// Log collection shells out like the power commands do, so it is probed the
// same way rather than assumed.
func TestDetectWithoutTheCollectorDeclaresNoLogCollection(t *testing.T) {
	withCommands(t, "shutdown", "reboot")

	caps := Detect(context.Background(), workerOn(clistate.State{TerminusState: clistate.TerminusRunning}))

	if _, ok := caps[CapLogsCollect]; ok {
		t.Error("logs.collect declared with no olares-cli on the machine")
	}
}

func TestDetectInContainerModeDeclaresNoPower(t *testing.T) {
	withCommands(t, "shutdown", "reboot", "olares-cli", "chpasswd")
	container := "docker"

	caps := Detect(context.Background(), workerOn(clistate.State{
		TerminusState: clistate.TerminusRunning,
		ContainerMode: &container,
	}))

	if _, ok := caps[CapPowerShutdown]; ok {
		t.Error("power.shutdown declared while olaresd runs in a container")
	}
	if _, ok := caps[CapPowerReboot]; ok {
		t.Error("power.reboot declared while olaresd runs in a container")
	}
	if _, ok := caps[CapSetSSHPassword]; ok {
		t.Error("ssh.setPassword declared while olaresd runs in a container")
	}
	if _, ok := caps[CapLogsCollect]; !ok {
		t.Error("logs.collect works in a container and should stay declared")
	}
}

func TestDetectWithoutChpasswdDeclaresNoSSHPassword(t *testing.T) {
	withCommands(t, "shutdown", "reboot", "olares-cli")

	caps := Detect(context.Background(), workerOn(clistate.State{TerminusState: clistate.TerminusRunning}))

	if _, ok := caps[CapSetSSHPassword]; ok {
		t.Error("ssh.setPassword declared with no chpasswd on the machine")
	}
}

// Powering the control node off is a cluster operation, not a node one: the
// node list, the orchestration and the caller all live on it.
func TestDetectOnTheMasterOffersNoSingleNodeShutdown(t *testing.T) {
	withCommands(t, "shutdown", "reboot", "olares-cli")

	caps := Detect(context.Background(), ProbeInput{
		Identity: Identity{Role: inventory.RoleMaster},
		State:    clistate.State{TerminusState: clistate.TerminusRunning},
	})

	if _, ok := caps[CapPowerShutdown]; ok {
		t.Error("power.shutdown declared on the control node")
	}
	if _, ok := caps[CapPowerReboot]; !ok {
		t.Error("the control node can still be rebooted on its own")
	}
}

func TestDetectWithAnUnresolvedRoleOffersNoShutdown(t *testing.T) {
	withCommands(t, "shutdown", "reboot")

	caps := Detect(context.Background(), ProbeInput{
		Identity: Identity{Role: inventory.RoleUnknown},
		State:    clistate.State{TerminusState: clistate.TerminusRunning},
	})

	if _, ok := caps[CapPowerShutdown]; ok {
		t.Error("power.shutdown declared on a node that cannot tell whether it is the control node")
	}
}

func TestDetectWithoutTheUnderlyingCommandsDeclaresNoPower(t *testing.T) {
	withCommands(t)

	caps := Detect(context.Background(), workerOn(clistate.State{TerminusState: clistate.TerminusRunning}))

	if _, ok := caps[CapPowerShutdown]; ok {
		t.Error("power.shutdown declared with no shutdown command on the machine")
	}
	if _, ok := caps[CapPowerReboot]; ok {
		t.Error("power.reboot declared with no reboot command on the machine")
	}
}

func TestDetectDeclaresEachPowerCommandOnItsOwn(t *testing.T) {
	withCommands(t, "reboot")

	caps := Detect(context.Background(), workerOn(clistate.State{TerminusState: clistate.TerminusRunning}))

	if _, ok := caps[CapPowerShutdown]; ok {
		t.Error("power.shutdown declared although only reboot resolves")
	}
	if _, ok := caps[CapPowerReboot]; !ok {
		t.Error("power.reboot should be declared when the reboot command resolves")
	}
}

func TestDetectDoesNotGuessDeviceSpecificCapabilities(t *testing.T) {
	withCommands(t, "shutdown", "reboot")

	caps := Detect(context.Background(), workerOn(clistate.State{TerminusState: clistate.TerminusRunning}))

	for _, name := range []string{"gpu_mode", "cpu_frequency_limit", "auto_power_on", "gb10_compute"} {
		if _, ok := caps[name]; ok {
			t.Errorf("capability %q declared without a probe that can confirm it", name)
		}
	}
}

func TestDetectSkipsCapabilitiesThatCannotBeProbed(t *testing.T) {
	unprobeable := func(context.Context, ProbeInput) (string, Capability, bool) {
		return "gb10_compute", Capability{Supported: true}, false
	}
	confirmed := func(context.Context, ProbeInput) (string, Capability, bool) {
		return "power.led", Capability{Supported: true, Config: map[string]any{"colors": []string{"red"}}}, true
	}

	caps := Detect(context.Background(), workerOn(clistate.State{}), unprobeable, confirmed)

	if _, ok := caps["gb10_compute"]; ok {
		t.Error("a probe that cannot confirm a capability must leave it undeclared")
	}
	c, ok := caps["power.led"]
	if !ok {
		t.Fatal("a confirmed capability should be declared")
	}
	if c.Config == nil {
		t.Errorf("capability config dropped: %+v", c)
	}
}

func TestDetectPassesIdentityAndStateToProbes(t *testing.T) {
	var seen ProbeInput
	spy := func(_ context.Context, in ProbeInput) (string, Capability, bool) {
		seen = in
		return "", Capability{}, false
	}
	in := ProbeInput{
		Identity: Identity{Role: inventory.RoleWorker, DeviceType: DeviceTypeOlaresOne},
		State:    clistate.State{TerminusState: clistate.TerminusRunning},
	}

	Detect(context.Background(), in, spy)

	if seen.Identity.DeviceType != DeviceTypeOlaresOne || seen.State.TerminusState != clistate.TerminusRunning {
		t.Errorf("probes cannot decide anything without identity and state, got %+v", seen)
	}
}

func TestDetectExtraProbesDoNotDropTheDefaults(t *testing.T) {
	withCommands(t, "shutdown", "reboot", "olares-cli")
	extra := func(context.Context, ProbeInput) (string, Capability, bool) {
		return "logs.stream", Capability{Supported: true}, true
	}

	caps := Detect(context.Background(), workerOn(clistate.State{TerminusState: clistate.TerminusRunning}), extra)

	if _, ok := caps[CapPowerShutdown]; !ok {
		t.Error("default capabilities must survive an extra probe")
	}
	if _, ok := caps["logs.stream"]; !ok {
		t.Error("extra probe not applied")
	}
}

func TestDeviceTypeFromDeviceName(t *testing.T) {
	cases := []struct {
		name string
		in   *string
		want string
	}{
		{"olares one, spaced", strp("Olares One"), DeviceTypeOlaresOne},
		{"olares one, joined", strp("OlaresOne"), DeviceTypeOlaresOne},
		{"olares one, padded", strp(" olaresone\n"), DeviceTypeOlaresOne},
		{"unbranded default", strp("Selfhosted"), "selfhosted"},
		{"unseen device", strp("DGX Spark"), "dgx-spark"},
		{"missing", nil, DeviceTypeUnknown},
		{"blank", strp("   "), DeviceTypeUnknown},
	}

	for _, c := range cases {
		if got := DeviceType(c.in); got != c.want {
			t.Errorf("%s: DeviceType = %q, want %q", c.name, got, c.want)
		}
	}
}
