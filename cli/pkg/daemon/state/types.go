// Package state defines the data types exposed by the local olaresd
// daemon's GET /system/status endpoint.
//
// These types are intentionally placed in the cli module so that both
// the olaresd daemon and the olares-cli command line tool can share
// the same wire format. The daemon (which already imports this module
// via its go.mod) re-exports these types as aliases, while the CLI
// uses them directly to unmarshal the HTTP response.
//
// Only data types belong here. Business logic that depends on
// daemon-internal packages (state validators, status probing, etc.)
// must remain in the daemon module.
package state

import "time"

// State is the full system status snapshot maintained by olaresd.
// It is refreshed every 5s by the daemon's status watcher and served
// as the `data` field of the GET /system/status response.
//
// All fields use JSON tags that match the wire format byte for byte;
// do not rename JSON keys without updating every consumer (CLI,
// frontend, mDNS clients, etc.).
type State struct {
	// TerminusdState is the lifecycle state of the olaresd daemon
	// itself. Possible values: "initialize" (just started, still
	// bootstrapping) or "running" (fully initialized).
	TerminusdState TerminusDState `json:"terminusdState"`

	// TerminusState is the high-level state of the Olares system.
	// It drives both UI display and command admission control. See
	// the TerminusState constants below for the full enumeration
	// and call Describe() to obtain a one-line explanation.
	TerminusState TerminusState `json:"terminusState"`

	// TerminusName is the Olares ID of the admin user, e.g.
	// "alice@olares.cn". It is read from the local release file when
	// available and refreshed from the cluster once Olares is up.
	TerminusName *string `json:"terminusName,omitempty"`

	// TerminusVersion is the installed Olares version (semver), e.g.
	// "1.12.0".
	TerminusVersion *string `json:"terminusVersion,omitempty"`

	// InstalledTime is the Unix epoch (seconds) at which Olares
	// finished installing on this node. Nil before install completes.
	InstalledTime *int64 `json:"installedTime,omitempty"`

	// InitializedTime is the Unix epoch (seconds) at which the admin
	// user finished the initial activation. Nil before activation.
	InitializedTime *int64 `json:"initializedTime,omitempty"`

	// OlaresdVersion is the running olaresd binary version. Useful
	// for diagnosing version drift between olaresd and the rest of
	// Olares after a partial upgrade.
	OlaresdVersion *string `json:"olaresdVersion,omitempty"`

	// InstallFinishedTime is daemon-internal: the wall clock time at
	// which the most recent install finished. Used to derive
	// InstalledTime when the cluster is not reachable yet. Excluded
	// from the wire format.
	InstallFinishedTime *time.Time `json:"-"`

	// DeviceName is the user-friendly device name (model / chassis
	// name) detected from the host.
	DeviceName *string `json:"device_name,omitempty"`

	// HostName is the kernel hostname of the node.
	HostName *string `json:"host_name,omitempty"`

	// OsType is the OS family, e.g. "linux" or "darwin".
	OsType string `json:"os_type"`

	// OsArch is the CPU architecture, e.g. "amd64" or "arm64".
	OsArch string `json:"os_arch"`

	// OsInfo is a human-readable OS distribution string, e.g.
	// "Ubuntu 22.04".
	OsInfo string `json:"os_info"`

	// OsVersion is the OS version string, e.g. "22.04".
	OsVersion string `json:"os_version"`

	// CpuInfo is the CPU model name as reported by the OS.
	CpuInfo string `json:"cpu_info"`

	// GpuInfo is the GPU model name when one is detected.
	GpuInfo *string `json:"gpu_info,omitempty"`

	// Memory is the total physical memory, formatted as "<n> G".
	Memory string `json:"memory"`

	// Disk is the total filesystem size of the node's data partition,
	// formatted as "<n> G".
	Disk string `json:"disk"`

	// WifiConnected is true when the active default route is over
	// Wi-Fi. The JSON key is "wifiConnected".
	WifiConnected bool `json:"wifiConnected"`

	// WifiSSID is the SSID of the connected Wi-Fi network, when
	// WifiConnected is true.
	WifiSSID *string `json:"wifiSSID,omitempty"`

	// WiredConnected is true when the node has an active Ethernet
	// connection.
	WiredConnected bool `json:"wiredConnected"`

	// HostIP is the internal LAN IPv4 address that Olares uses to
	// register itself in /etc/hosts and to reach other nodes.
	HostIP string `json:"hostIp"`

	// ExternalIP is the public IPv4 address as observed by an
	// external probe. Refreshed at most once per minute.
	ExternalIP string `json:"externalIp"`

	// ExternalIPProbeTime is daemon-internal: when the external IP
	// probe last ran. Excluded from the wire format.
	ExternalIPProbeTime time.Time `json:"-"`

	// InstallingState reports the progress of an installation in
	// flight: "in-progress", "completed", "failed", or empty.
	InstallingState ProcessingState `json:"installingState"`

	// InstallingProgress is a free-form human-readable description
	// of the current installation step.
	InstallingProgress string `json:"installingProgress"`

	// InstallingProgressNum is daemon-internal: the latest numeric
	// install progress percentage. Excluded from the wire format.
	InstallingProgressNum int `json:"-"`

	// UninstallingState mirrors InstallingState for the uninstall
	// flow.
	UninstallingState ProcessingState `json:"uninstallingState"`

	// UninstallingProgress is a free-form description of the
	// current uninstall step.
	UninstallingProgress string `json:"uninstallingProgress"`

	// UninstallingProgressNum is daemon-internal: the latest numeric
	// uninstall progress percentage. Excluded from the wire format.
	UninstallingProgressNum int `json:"-"`

	// UpgradingTarget is the target version of the in-flight
	// upgrade, e.g. "1.13.0". Empty when no upgrade is queued.
	UpgradingTarget string `json:"upgradingTarget"`

	// UpgradingRetryNum is the number of times the upgrader has
	// retried after a transient failure.
	UpgradingRetryNum int `json:"upgradingRetryNum"`

	// UpgradingNextRetryAt is the wall-clock time at which the next
	// retry will fire, when retries are pending.
	UpgradingNextRetryAt *time.Time `json:"upgradingNextRetryAt,omitempty"`

	// UpgradingState is the lifecycle of the install phase of the
	// upgrade ("in-progress", "completed", "failed", or empty).
	UpgradingState ProcessingState `json:"upgradingState"`

	// UpgradingStep is the name of the current upgrade step.
	UpgradingStep string `json:"upgradingStep"`

	// UpgradingProgress is the free-form progress message for the
	// current upgrade step.
	UpgradingProgress string `json:"upgradingProgress"`

	// UpgradingProgressNum is daemon-internal: the latest numeric
	// upgrade progress percentage. Excluded from the wire format.
	UpgradingProgressNum int `json:"-"`

	// UpgradingError is the most recent error seen during upgrade.
	// Empty when no error has occurred.
	UpgradingError string `json:"upgradingError"`

	// UpgradingDownloadState is the lifecycle of the download phase
	// of the upgrade. Olares splits download and install into two
	// phases so that downloads can complete in the background
	// without changing TerminusState to "upgrading".
	UpgradingDownloadState ProcessingState `json:"upgradingDownloadState"`

	// UpgradingDownloadStep is the name of the current download step.
	UpgradingDownloadStep string `json:"upgradingDownloadStep"`

	// UpgradingDownloadProgress is the free-form progress message
	// for the current download step.
	UpgradingDownloadProgress string `json:"upgradingDownloadProgress"`

	// UpgradingDownloadProgressNum is daemon-internal: the latest
	// numeric download progress percentage. Excluded from the wire
	// format.
	UpgradingDownloadProgressNum int `json:"-"`

	// UpgradingDownloadError is the most recent error from the
	// download phase. Empty when no error has occurred.
	UpgradingDownloadError string `json:"upgradingDownloadError"`

	// CollectingLogsState is the lifecycle of the most recent log
	// collection job triggered through olaresd.
	CollectingLogsState ProcessingState `json:"collectingLogsState"`

	// CollectingLogsError is the error from the most recent log
	// collection job, when it failed.
	CollectingLogsError string `json:"collectingLogsError"`

	// DefaultFRPServer is the FRP server address used when frp is
	// enabled. Sourced from the FRP_SERVER env var.
	DefaultFRPServer string `json:"defaultFrpServer"`

	// FRPEnable indicates whether the FRP-based reverse tunnel is
	// turned on. Sourced from the FRP_ENABLE env var.
	FRPEnable string `json:"frpEnable"`

	// ContainerMode is set when olaresd is running inside a
	// container, mirroring the CONTAINER_MODE env var.
	ContainerMode *string `json:"containerMode,omitempty"`

	// Pressure lists the kubernetes node-condition pressures
	// currently active on this node (memory pressure, disk pressure,
	// PID pressure, etc.). Empty when the node is healthy.
	Pressure []NodePressure `json:"pressures,omitempty"`
}

// NodePressure represents a non-Ready kubernetes node condition that
// is currently true on this node, e.g. MemoryPressure, DiskPressure,
// PIDPressure, NetworkUnavailable.
type NodePressure struct {
	// Type is the kubernetes node condition type, e.g.
	// "MemoryPressure".
	Type string `json:"type"`

	// Message is the human-readable explanation provided by kubelet.
	Message string `json:"message"`
}