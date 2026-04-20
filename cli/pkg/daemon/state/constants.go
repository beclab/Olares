package state

// TerminusDState is the lifecycle state of the olaresd daemon process.
type TerminusDState string

const (
	// Initialize means olaresd has just started and is still
	// bootstrapping its watchers and configuration.
	Initialize TerminusDState = "initialize"

	// Running means olaresd has finished initialization and is
	// serving requests normally.
	Running TerminusDState = "running"
)

// ProcessingState is the lifecycle of a long-running operation that
// olaresd reports progress for (install, uninstall, upgrade, log
// collection).
type ProcessingState string

const (
	// Completed means the operation finished successfully.
	Completed ProcessingState = "completed"

	// Failed means the operation finished with an error. Inspect the
	// associated *Error field for details.
	Failed ProcessingState = "failed"

	// InProgress means the operation is currently running.
	InProgress ProcessingState = "in-progress"
)

// TerminusState is the high-level state machine value for the Olares
// system as observed from this node. Use Describe() to obtain a
// human-readable, one-line summary suitable for end-user output.
type TerminusState string

const (
	// NotInstalled means Olares is not installed on this node.
	NotInstalled TerminusState = "not-installed"

	// Installing means an installation is currently in progress.
	Installing TerminusState = "installing"

	// InstallFailed means the most recent installation attempt
	// failed. Re-run install or uninstall to recover.
	InstallFailed TerminusState = "install-failed"

	// Uninitialized means Olares is installed but the admin user has
	// not completed initial activation yet.
	Uninitialized TerminusState = "uninitialized"

	// Initializing means the admin user is going through the initial
	// activation flow.
	Initializing TerminusState = "initializing"

	// InitializeFailed means the initial activation failed.
	InitializeFailed TerminusState = "initialize-failed"

	// TerminusRunning means Olares is fully installed, activated,
	// and all key pods are healthy.
	TerminusRunning TerminusState = "terminus-running"

	// InvalidIpAddress means the node's IP has changed since
	// installation; run change-ip to fix it.
	InvalidIpAddress TerminusState = "invalid-ip-address"

	// SystemError means one or more critical pods are not running,
	// or the cluster API is unreachable.
	SystemError TerminusState = "system-error"

	// SelfRepairing means olaresd is automatically attempting to
	// recover from a system error.
	SelfRepairing TerminusState = "self-repairing"

	// IPChanging means a change-ip operation is currently running.
	IPChanging TerminusState = "ip-changing"

	// IPChangeFailed means the most recent change-ip attempt failed.
	IPChangeFailed TerminusState = "ip-change-failed"

	// AddingNode means a worker node is currently being joined.
	AddingNode TerminusState = "adding-node"

	// RemovingNode means a worker node is currently being removed.
	RemovingNode TerminusState = "removing-node"

	// Uninstalling means an uninstall is currently in progress.
	Uninstalling TerminusState = "uninstalling"

	// Upgrading means an upgrade install phase is currently running.
	// The dedicated download phase does not flip TerminusState.
	Upgrading TerminusState = "upgrading"

	// DiskModifing means a storage reconfiguration is in progress.
	DiskModifing TerminusState = "disk-modifing"

	// Shutdown means the system is in the process of shutting down.
	Shutdown TerminusState = "shutdown"

	// Restarting means the node has been up for less than the
	// stabilization window (3 minutes for healthy systems, 10 for
	// degraded ones), so reported pod state may be stale.
	Restarting TerminusState = "restarting"

	// Checking means olaresd has not yet completed the first status
	// probe. This is the default value before WatchStatus runs.
	Checking TerminusState = "checking"

	// NetworkNotReady means no usable internal IPv4 address was
	// detected on this node.
	NetworkNotReady TerminusState = "network-not-ready"
)

// String returns the wire value of the state, allowing TerminusState
// to satisfy fmt.Stringer.
func (s TerminusState) String() string {
	return string(s)
}

// Describe returns a one-line, end-user oriented explanation of the
// state value. Empty values are rendered as "unknown state". Unknown
// values are returned as-is so the CLI can still display them.
func (s TerminusState) Describe() string {
	switch s {
	case NotInstalled:
		return "Olares is not installed on this node"
	case Installing:
		return "Olares is currently being installed"
	case InstallFailed:
		return "the most recent install attempt failed"
	case Uninitialized:
		return "Olares is installed but the admin user has not been activated yet"
	case Initializing:
		return "the admin user activation is in progress"
	case InitializeFailed:
		return "the admin user activation failed"
	case TerminusRunning:
		return "Olares is running normally"
	case InvalidIpAddress:
		return "the node IP changed since install; run change-ip to recover"
	case SystemError:
		return "one or more critical pods are not running"
	case SelfRepairing:
		return "olaresd is attempting automatic recovery"
	case IPChanging:
		return "a change-ip operation is in progress"
	case IPChangeFailed:
		return "the most recent change-ip attempt failed"
	case AddingNode:
		return "a worker node is being joined"
	case RemovingNode:
		return "a worker node is being removed"
	case Uninstalling:
		return "Olares is being uninstalled"
	case Upgrading:
		return "an upgrade is being applied"
	case DiskModifing:
		return "the storage layout is being modified"
	case Shutdown:
		return "the system is shutting down"
	case Restarting:
		return "the node was just restarted, status will stabilize shortly"
	case Checking:
		return "olaresd has not finished the first status probe yet"
	case NetworkNotReady:
		return "no usable internal IPv4 address detected"
	case "":
		return "unknown state"
	default:
		return string(s)
	}
}

// AllTerminusStates returns the full list of TerminusState values in
// a stable, documentation-friendly order. It is used by the CLI's
// long help text and by the docs generator.
func AllTerminusStates() []TerminusState {
	return []TerminusState{
		Checking,
		NetworkNotReady,
		NotInstalled,
		Installing,
		InstallFailed,
		Uninitialized,
		Initializing,
		InitializeFailed,
		TerminusRunning,
		Restarting,
		InvalidIpAddress,
		IPChanging,
		IPChangeFailed,
		SystemError,
		SelfRepairing,
		AddingNode,
		RemovingNode,
		Uninstalling,
		Upgrading,
		DiskModifing,
		Shutdown,
	}
}
