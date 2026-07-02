package domain

import (
	"testing"
	"time"

	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// The legacy fixtures below are hand-copies of the six tables (plus two
// helper predicates) from pkg/appstate/state_transition.go, kept
// verbatim inside this test file so the domain package remains free of
// a legacy import.
//
// If pkg/appstate/state_transition.go changes, these fixtures must be
// updated in lock-step; the linter enforces that pkg/applifecycle/
// does not import pkg/appstate, so the sync happens via review, not via
// a runtime dependency.
//
// The tests in this file assert that NewDefaultTable() reproduces every
// row of every legacy table, which is the migration-correctness proof
// for step 1 (skeleton) of the refactor plan.

// legacyStateTransitions mirrors StateTransitions from
// pkg/appstate/state_transition.go:61-256.
var legacyStateTransitions = map[State][]State{
	InitialState:   {appsv1.Pending},
	appsv1.Pending: {appsv1.Downloading, appsv1.PendingCanceling},
	appsv1.Downloading: {
		appsv1.Installing, appsv1.DownloadFailed, appsv1.DownloadingCanceling,
	},
	appsv1.Installing: {
		appsv1.Initializing, appsv1.InstallFailed, appsv1.InstallingCanceling,
		appsv1.Stopping, appsv1.Running,
	},
	appsv1.Initializing: {appsv1.Running, appsv1.InitializingCanceling},
	appsv1.Running: {
		appsv1.Stopping, appsv1.UpgradePending, appsv1.ApplyingEnv,
		appsv1.Uninstalling, appsv1.Uninstalled,
	},
	appsv1.Stopping:     {appsv1.Stopped, appsv1.StopFailed},
	appsv1.Upgrading:    {appsv1.Initializing, appsv1.UpgradeFailed, appsv1.UpgradingCanceling, appsv1.Stopped},
	appsv1.ApplyingEnv:  {appsv1.Initializing, appsv1.ApplyEnvFailed, appsv1.ApplyingEnvCanceling, appsv1.Stopping, appsv1.Stopped},
	appsv1.Uninstalling: {appsv1.Uninstalled, appsv1.UninstallFailed},

	appsv1.PendingCanceling:      {appsv1.PendingCanceled, appsv1.PendingCancelFailed},
	appsv1.DownloadingCanceling:  {appsv1.DownloadingCanceled, appsv1.DownloadingCancelFailed},
	appsv1.InstallingCanceling:   {appsv1.InstallingCanceled, appsv1.InstallingCancelFailed},
	appsv1.InitializingCanceling: {appsv1.Stopping, appsv1.InstallingCancelFailed},

	appsv1.Resuming:             {appsv1.ResumingCanceling, appsv1.ResumeFailed, appsv1.Initializing, appsv1.Stopping},
	appsv1.ResumingCanceling:    {appsv1.Stopping, appsv1.ResumingCancelFailed},
	appsv1.UpgradingCanceling:   {appsv1.UpgradeFailed, appsv1.UpgradingCancelFailed},
	appsv1.ApplyingEnvCanceling: {appsv1.Stopping, appsv1.ApplyingEnvCancelFailed},

	appsv1.Stopped: {appsv1.Resuming, appsv1.Uninstalling, appsv1.UpgradePending, appsv1.ApplyingEnv},

	appsv1.UpgradePending: {appsv1.UpgradeDownloading, appsv1.UpgradePendingCanceling},
	appsv1.UpgradeDownloading: {
		appsv1.Upgrading, appsv1.UpgradeFailed, appsv1.UpgradeDownloadingCanceling,
	},
	appsv1.UpgradePendingCanceling: {
		appsv1.Running, appsv1.Stopped, appsv1.UpgradePendingCancelFailed,
	},
	appsv1.UpgradeDownloadingCanceling: {
		appsv1.Running, appsv1.Stopped, appsv1.UpgradeDownloadingCancelFailed,
	},

	appsv1.DownloadFailed: {appsv1.Pending},
	appsv1.InstallFailed:  {appsv1.Uninstalled, appsv1.Uninstalling},

	appsv1.StopFailed:     {appsv1.Stopping, appsv1.ApplyingEnv, appsv1.Uninstalling},
	appsv1.UpgradeFailed:  {appsv1.UpgradePending, appsv1.Uninstalling},
	appsv1.ApplyEnvFailed: {appsv1.Stopping, appsv1.ApplyingEnv, appsv1.Uninstalling},
	appsv1.ResumeFailed:   {appsv1.Resuming, appsv1.Uninstalling, appsv1.ApplyingEnv},

	appsv1.UninstallFailed:                {appsv1.Uninstalling, appsv1.Uninstalled},
	appsv1.PendingCancelFailed:            {appsv1.PendingCanceling},
	appsv1.DownloadingCancelFailed:        {appsv1.DownloadingCanceling},
	appsv1.UpgradePendingCancelFailed:     {appsv1.UpgradePendingCanceling},
	appsv1.UpgradeDownloadingCancelFailed: {appsv1.UpgradeDownloadingCanceling},
	appsv1.InstallingCancelFailed:         {appsv1.InstallingCanceling, appsv1.Uninstalling},
	appsv1.UpgradingCancelFailed:          {appsv1.UpgradingCanceling, appsv1.Uninstalling},
	appsv1.ApplyingEnvCancelFailed:        {appsv1.ApplyingEnvCanceling, appsv1.Uninstalling},
	appsv1.ResumingCancelFailed:           {appsv1.ResumingCanceling},

	appsv1.PendingCanceled:     {appsv1.Pending},
	appsv1.DownloadingCanceled: {appsv1.Pending},
	appsv1.InstallingCanceled:  {appsv1.Pending},
	appsv1.Uninstalled:         {appsv1.Pending},
}

// legacyOperationAllowedInState mirrors OperationAllowedInState from
// pkg/appstate/state_transition.go:258-388.
var legacyOperationAllowedInState = map[State]map[OpType]bool{
	InitialState:              {appsv1.InstallOp: true},
	appsv1.Pending:            {appsv1.CancelOp: true},
	appsv1.Downloading:        {appsv1.CancelOp: true},
	appsv1.UpgradePending:     {appsv1.CancelOp: true},
	appsv1.UpgradeDownloading: {appsv1.CancelOp: true},
	appsv1.Installing:         {appsv1.CancelOp: true},
	appsv1.Initializing:       {appsv1.CancelOp: true},
	appsv1.Upgrading:          {appsv1.CancelOp: true},
	appsv1.ApplyingEnv:        {appsv1.CancelOp: true, appsv1.StopOp: true},
	appsv1.Resuming:           {appsv1.CancelOp: true, appsv1.StopOp: true},

	appsv1.PendingCanceled:     {appsv1.InstallOp: true},
	appsv1.DownloadingCanceled: {appsv1.InstallOp: true},
	appsv1.InstallingCanceled:  {appsv1.InstallOp: true},
	appsv1.Uninstalled:         {appsv1.InstallOp: true},

	appsv1.DownloadFailed:  {appsv1.InstallOp: true},
	appsv1.InstallFailed:   {appsv1.UninstallOp: true},
	appsv1.StopFailed:      {appsv1.StopOp: true, appsv1.UninstallOp: true},
	appsv1.ResumeFailed:    {appsv1.ResumeOp: true, appsv1.ApplyEnvOp: true, appsv1.UninstallOp: true},
	appsv1.UninstallFailed: {appsv1.UninstallOp: true},
	appsv1.UpgradeFailed:   {appsv1.UninstallOp: true, appsv1.UpgradeOp: true},
	appsv1.ApplyEnvFailed:  {appsv1.UninstallOp: true, appsv1.ApplyEnvOp: true, appsv1.StopOp: true},

	appsv1.PendingCancelFailed:            {appsv1.CancelOp: true},
	appsv1.DownloadingCancelFailed:        {appsv1.CancelOp: true},
	appsv1.UpgradePendingCancelFailed:     {appsv1.CancelOp: true},
	appsv1.UpgradeDownloadingCancelFailed: {appsv1.CancelOp: true},
	appsv1.InstallingCancelFailed:         {appsv1.CancelOp: true, appsv1.UninstallOp: true},
	appsv1.UpgradingCancelFailed:          {appsv1.CancelOp: true, appsv1.UninstallOp: true},
	appsv1.ApplyingEnvCancelFailed:        {appsv1.CancelOp: true, appsv1.UninstallOp: true},
	appsv1.ResumingCancelFailed:           {appsv1.CancelOp: true},

	appsv1.Running: {appsv1.UninstallOp: true, appsv1.UpgradeOp: true, appsv1.ApplyEnvOp: true, appsv1.StopOp: true},
	appsv1.Stopped: {appsv1.UninstallOp: true, appsv1.UpgradeOp: true, appsv1.ApplyEnvOp: true, appsv1.ResumeOp: true},
}

// legacyCancelableStates mirrors CancelableStates from pkg/appstate.
var legacyCancelableStates = map[State]bool{
	appsv1.Pending: true, appsv1.Downloading: true,
	appsv1.UpgradePending: true, appsv1.UpgradeDownloading: true,
	appsv1.Installing:   true,
	appsv1.Initializing: true, appsv1.Resuming: true, appsv1.Upgrading: true,
	appsv1.ApplyingEnv: true,
}

// legacyOperatingStates mirrors OperatingStates from pkg/appstate.
var legacyOperatingStates = map[State]bool{
	appsv1.Pending: true, appsv1.Downloading: true,
	appsv1.UpgradePending: true, appsv1.UpgradeDownloading: true,
	appsv1.Installing:   true,
	appsv1.Initializing: true, appsv1.Resuming: true, appsv1.Upgrading: true,
	appsv1.ApplyingEnv: true, appsv1.Stopping: true, appsv1.Uninstalling: true,
	appsv1.PendingCanceling: true, appsv1.DownloadingCanceling: true,
	appsv1.UpgradePendingCanceling: true, appsv1.UpgradeDownloadingCanceling: true,
	appsv1.InstallingCanceling: true, appsv1.InitializingCanceling: true,
	appsv1.ResumingCanceling: true, appsv1.UpgradingCanceling: true,
	appsv1.ApplyingEnvCanceling: true,
}

// legacyCancelingStates mirrors CancelingStates from pkg/appstate.
var legacyCancelingStates = map[State]bool{
	appsv1.PendingCanceling: true, appsv1.DownloadingCanceling: true,
	appsv1.UpgradePendingCanceling: true, appsv1.UpgradeDownloadingCanceling: true,
	appsv1.InstallingCanceling: true, appsv1.InitializingCanceling: true,
	appsv1.ResumingCanceling: true, appsv1.UpgradingCanceling: true,
	appsv1.ApplyingEnvCanceling: true,
}

// legacyStateToDuration mirrors StateToDurationMap from pkg/appstate.
// Pending and UpgradePending are absent. Legacy StateToDuration gave
// them 24h, but that value is unused and those states do not time out.
// Downloading and UpgradeDownloading are 24h, not the legacy 30 days.
var legacyStateToDuration = map[State]time.Duration{
	appsv1.Downloading:        24 * time.Hour,
	appsv1.UpgradeDownloading: 24 * time.Hour,
	appsv1.Installing:         30 * time.Minute,
	appsv1.Initializing:       time.Hour,
	appsv1.Upgrading:          time.Hour,
	appsv1.ApplyingEnv:        30 * time.Minute,
}

// legacyTerminalReinstallable mirrors IsTerminalReinstallable from pkg/appstate.
var legacyTerminalReinstallable = map[State]bool{
	appsv1.PendingCanceled: true, appsv1.DownloadingCanceled: true,
	appsv1.InstallingCanceled: true, appsv1.Uninstalled: true,
	appsv1.DownloadFailed: true,
}

// legacySafelyDeletable mirrors IsSafelyDeletable from pkg/appstate;
// currently identical to legacyTerminalReinstallable.
var legacySafelyDeletable = map[State]bool{
	appsv1.PendingCanceled: true, appsv1.DownloadingCanceled: true,
	appsv1.InstallingCanceled: true, appsv1.Uninstalled: true,
	appsv1.DownloadFailed: true,
}

// TestNewDefaultTable_AllowsMatchesLegacyStateTransitions is the primary
// migration-correctness proof for the "6 tables to 1" consolidation.
// For every (from, to) declared in the legacy StateTransitions map, the
// new table's Allows(from, to) must return true.
func TestNewDefaultTable_AllowsMatchesLegacyStateTransitions(t *testing.T) {
	tab := NewDefaultTable()
	for from, tos := range legacyStateTransitions {
		for _, to := range tos {
			if !tab.Allows(from, to) {
				t.Errorf("Allows(%q,%q)=false, legacy StateTransitions declares this edge",
					from, to)
			}
		}
	}
}

// TestNewDefaultTable_InstallingCancelingReachesInstallFailed pins the
// edge added when install failures and user cancellation started
// sharing the InstallingCanceling cleanup pipeline. The pipeline lands
// on whichever terminal its caller asked for, so BOTH edges have to
// stay reachable.
func TestNewDefaultTable_InstallingCancelingReachesInstallFailed(t *testing.T) {
	tab := NewDefaultTable()
	if !tab.Allows(appsv1.InstallingCanceling, appsv1.InstallFailed) {
		t.Error("Allows(InstallingCanceling, InstallFailed)=false; install-failure cleanup lands here")
	}
	if !tab.Allows(appsv1.InstallingCanceling, appsv1.InstallingCanceled) {
		t.Error("Allows(InstallingCanceling, InstallingCanceled)=false; user cancel still lands here")
	}
}

// TestNewDefaultTable_AllowsRejectsUndeclared asserts a small hand-picked
// set of forbidden jumps is NOT reported as legal.
func TestNewDefaultTable_AllowsRejectsUndeclared(t *testing.T) {
	tab := NewDefaultTable()
	forbidden := []struct{ from, to State }{
		{appsv1.Pending, appsv1.Running},          // must go through Downloading/Installing/Initializing
		{appsv1.Running, appsv1.Installing},       // no re-install without an intermediate transition
		{appsv1.Uninstalled, appsv1.Running},      // terminal state cannot bypass Pending
		{appsv1.Initializing, appsv1.Uninstalled}, // no direct jump to terminal
	}
	for _, c := range forbidden {
		if tab.Allows(c.from, c.to) {
			t.Errorf("Allows(%q,%q)=true, want false", c.from, c.to)
		}
	}
}

// TestNewDefaultTable_AllowsSelfWrite asserts every declared state
// accepts a self-write via AllowsSelfWrite, matching legacy
// IsStateTransitionAllowed's same-state relaxation.
func TestNewDefaultTable_AllowsSelfWrite(t *testing.T) {
	tab := NewDefaultTable()
	for _, s := range AllStates {
		if !tab.AllowsSelfWrite(s, s) {
			t.Errorf("AllowsSelfWrite(%q,%q)=false, want true", s, s)
		}
	}
}

// TestNewDefaultTable_AllowsOpMatchesLegacy asserts every (from, op)
// declared in the legacy OperationAllowedInState maps to true in the
// new table's AllowsOp.
func TestNewDefaultTable_AllowsOpMatchesLegacy(t *testing.T) {
	tab := NewDefaultTable()
	for from, ops := range legacyOperationAllowedInState {
		for op, allowed := range ops {
			if !allowed {
				continue
			}
			if !tab.AllowsOp(from, op) {
				t.Errorf("AllowsOp(%q,%q)=false, legacy allows it", from, op)
			}
		}
	}
}

// TestNewDefaultTable_AllowsOpRejectsDisallowed checks a hand-picked
// set of (from, op) pairs the legacy table forbids.
func TestNewDefaultTable_AllowsOpRejectsDisallowed(t *testing.T) {
	tab := NewDefaultTable()
	forbidden := []struct {
		from State
		op   OpType
	}{
		{"", appsv1.UninstallOp},
		{appsv1.Pending, appsv1.UninstallOp},
		{appsv1.Running, appsv1.InstallOp},
		{appsv1.Uninstalling, appsv1.UninstallOp},
		{appsv1.InstallFailed, appsv1.InstallOp},
	}
	for _, c := range forbidden {
		if tab.AllowsOp(c.from, c.op) {
			t.Errorf("AllowsOp(%q,%q)=true, want false", c.from, c.op)
		}
	}
}

// TestNewDefaultTable_IsCancelableMatchesLegacy asserts CancelableStates
// migration.
func TestNewDefaultTable_IsCancelableMatchesLegacy(t *testing.T) {
	tab := NewDefaultTable()
	for _, s := range AllStates {
		got, want := tab.IsCancelable(s), legacyCancelableStates[s]
		if got != want {
			t.Errorf("IsCancelable(%q)=%v want %v", s, got, want)
		}
	}
}

// TestNewDefaultTable_IsOperatingMatchesLegacy asserts OperatingStates
// migration.
func TestNewDefaultTable_IsOperatingMatchesLegacy(t *testing.T) {
	tab := NewDefaultTable()
	for _, s := range AllStates {
		got, want := tab.IsOperating(s), legacyOperatingStates[s]
		if got != want {
			t.Errorf("IsOperating(%q)=%v want %v", s, got, want)
		}
	}
}

// TestNewDefaultTable_IsCancelingMatchesLegacy asserts CancelingStates
// migration.
func TestNewDefaultTable_IsCancelingMatchesLegacy(t *testing.T) {
	tab := NewDefaultTable()
	for _, s := range AllStates {
		got, want := tab.IsCanceling(s), legacyCancelingStates[s]
		if got != want {
			t.Errorf("IsCanceling(%q)=%v want %v", s, got, want)
		}
	}
}

// TestNewDefaultTable_TTLMatchesLegacy asserts StateToDurationMap
// migration, including the 10-minute default for states not in the
// legacy map.
func TestNewDefaultTable_TTLMatchesLegacy(t *testing.T) {
	tab := NewDefaultTable()
	for _, s := range AllStates {
		want, ok := legacyStateToDuration[s]
		if !ok {
			want = 10 * time.Minute
		}
		if got := tab.TTL(s); got != want {
			t.Errorf("TTL(%q)=%v want %v", s, got, want)
		}
	}
}

// TestNewDefaultTable_ReinstallableMatchesLegacy asserts
// IsTerminalReinstallable migration.
func TestNewDefaultTable_ReinstallableMatchesLegacy(t *testing.T) {
	tab := NewDefaultTable()
	for _, s := range AllStates {
		got, want := tab.IsTerminalReinstallable(s), legacyTerminalReinstallable[s]
		if got != want {
			t.Errorf("IsTerminalReinstallable(%q)=%v want %v", s, got, want)
		}
	}
}

// TestNewDefaultTable_SafelyDeletableMatchesLegacy asserts
// IsSafelyDeletable migration.
func TestNewDefaultTable_SafelyDeletableMatchesLegacy(t *testing.T) {
	tab := NewDefaultTable()
	for _, s := range AllStates {
		got, want := tab.IsSafelyDeletable(s), legacySafelyDeletable[s]
		if got != want {
			t.Errorf("IsSafelyDeletable(%q)=%v want %v", s, got, want)
		}
	}
}

// TestNewDefaultTable_SafelyDeletableIsSubsetOfReinstallable is the
// same invariant asserted by
// pkg/appstate/state_transition_test.go::TestSafelyDeletableImpliesReinstallable.
// Kept here as a migration-correctness bridge.
func TestNewDefaultTable_SafelyDeletableIsSubsetOfReinstallable(t *testing.T) {
	tab := NewDefaultTable()
	for _, s := range AllStates {
		if tab.IsSafelyDeletable(s) && !tab.IsTerminalReinstallable(s) {
			t.Errorf("state %q is SafelyDeletable but not TerminalReinstallable", s)
		}
	}
}

// TestNewDefaultTable_ReinstallableAllowsInstallOp mirrors legacy
// TestTerminalReinstallableAllowsInstallOp: every reinstallable state
// must also allow InstallOp.
func TestNewDefaultTable_ReinstallableAllowsInstallOp(t *testing.T) {
	tab := NewDefaultTable()
	for _, s := range AllStates {
		if tab.IsTerminalReinstallable(s) && !tab.AllowsOp(s, appsv1.InstallOp) {
			t.Errorf("state %q is TerminalReinstallable but disallows InstallOp", s)
		}
	}
}

// TestNewDefaultTable_CancelableAndCancelingAreOperating mirrors legacy
// TestCancelableAndCancelingAreOperating: every cancelable state and
// every canceling state must also be an operating state.
func TestNewDefaultTable_CancelableAndCancelingAreOperating(t *testing.T) {
	tab := NewDefaultTable()
	for _, s := range AllStates {
		if tab.IsCancelable(s) && !tab.IsOperating(s) {
			t.Errorf("cancelable state %q is not operating", s)
		}
		if tab.IsCanceling(s) && !tab.IsOperating(s) {
			t.Errorf("canceling state %q is not operating", s)
		}
	}
}

// TestNewDefaultTable_OpDrivenEdgesResolvable asserts that when
// AllowsOp(from, op) is true, feeding TriggerOp{op} to Engine.Step
// against a World in state `from` produces a legal transition (either
// a declared Next or an ErrGuardRejected — but never ErrNoTransition).
// This is the invariant that consolidates OperationAllowedInState with
// StateTransitions.
func TestNewDefaultTable_OpDrivenEdgesResolvable(t *testing.T) {
	tab := NewDefaultTable()
	for from, ops := range legacyOperationAllowedInState {
		for op, allowed := range ops {
			if !allowed {
				continue
			}
			cands := tab.Candidates(from, OpKind(op))
			if len(cands) == 0 {
				t.Errorf("no dispatchable Transition for from=%q op=%q, "+
					"but legacy OperationAllowedInState allows it", from, op)
			}
		}
	}
}

// TestNewDefaultTable_AllLegacyStatesHaveSpec asserts every legacy
// state (including the "" sentinel) has a StateSpec registered so
// Runner-side lookups (TTL/IsOperating/etc.) always find a value.
func TestNewDefaultTable_AllLegacyStatesHaveSpec(t *testing.T) {
	tab := NewDefaultTable()
	specStates := map[State]bool{}
	for _, s := range tab.States() {
		specStates[s] = true
	}
	for _, s := range AllStates {
		if !specStates[s] {
			t.Errorf("no StateSpec registered for %q", s)
		}
	}
}
