package state

import (
	"errors"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

// Wire-format types live in the cli module so that olaresd and the
// olares-cli command share a single source of truth. Daemon code keeps
// using the unqualified names below via these type aliases.

type ProcessingState = clistate.ProcessingState

const (
	Completed  = clistate.Completed
	Failed     = clistate.Failed
	InProgress = clistate.InProgress
)

type TerminusState = clistate.TerminusState

const (
	NotInstalled     = clistate.NotInstalled
	Installing       = clistate.Installing
	InstallFailed    = clistate.InstallFailed
	Uninitialized    = clistate.Uninitialized
	Initializing     = clistate.Initializing
	InitializeFailed = clistate.InitializeFailed
	TerminusRunning  = clistate.TerminusRunning
	InvalidIpAddress = clistate.InvalidIpAddress
	SystemError      = clistate.SystemError
	SelfRepairing    = clistate.SelfRepairing
	IPChanging       = clistate.IPChanging
	IPChangeFailed   = clistate.IPChangeFailed
	AddingNode       = clistate.AddingNode
	RemovingNode     = clistate.RemovingNode
	Uninstalling     = clistate.Uninstalling
	Upgrading        = clistate.Upgrading
	DiskModifing     = clistate.DiskModifing
	Shutdown         = clistate.Shutdown
	Restarting       = clistate.Restarting
	Checking         = clistate.Checking
	NetworkNotReady  = clistate.NetworkNotReady
)

// ValidateOp returns nil if the operation is allowed in the given
// state, or a descriptive error otherwise. It replaces the previous
// (TerminusState).ValidateOp method, since methods cannot be defined
// on an alias of a type from another package.
func ValidateOp(s TerminusState, op commands.Interface) error {
	return getValidator(s).ValidateOp(op)
}

func getValidator(s TerminusState) Validator {
	switch s {
	case NotInstalled:
		return &NotInstalledValidator{}
	case Uninitialized:
		return &UninitializedValidator{}
	case Initializing:
		return &InitializingValidator{}
	case InitializeFailed:
		return &InitializeFailedValidator{}
	case Installing:
		return &InstallingValidator{}
	case InstallFailed:
		return &InstallFailedValidator{}
	case TerminusRunning:
		return &RunningValidator{}
	case Upgrading:
		return &UpgradingValidator{}
	case InvalidIpAddress:
		return &InvalidIpValidator{}
	case SystemError:
		return &SystemErrorValidator{}
	case SelfRepairing:
		return &SelfRepairingValidator{}
	case IPChanging:
		return &IpChangingValidator{}
	case IPChangeFailed:
		return &IpChangeFailedValidator{}
	case AddingNode:
		return &AddingNodeValidator{}
	case RemovingNode:
		return &RemovingNodeValidator{}
	case Uninstalling:
		return &UninstallingValidator{}
	case DiskModifing:
		return &DiskModifingValidator{}
	case Shutdown:
		return &ShutdownValidator{}
	case Restarting:
		return &RestartingValidator{}
	default:
		return &UnknownStateValidator{}
	}
}

type Validator interface {
	ValidateOp(op commands.Interface) error
}

// not-installed
type NotInstalledValidator struct{}

func (n NotInstalledValidator) ValidateOp(op commands.Interface) error {
	switch op.OperationName() {
	case commands.Install, commands.ChangeIp, commands.Shutdown,
		commands.Reboot, commands.ConnectWifi, commands.ChangeHost,
		commands.MountSmb, commands.UmountSmb, commands.SetSSHPassword,
		commands.MountNfs, commands.UmountNfs:
		return nil
	}

	return errors.New("olares is not installed, cannot perform the operation")
}

// uninitialized
type UninitializedValidator struct{}

func (u UninitializedValidator) ValidateOp(op commands.Interface) error {
	switch op.OperationName() {
	case commands.Initialize, commands.ChangeIp, commands.Reboot,
		commands.Shutdown, commands.Uninstall, commands.ConnectWifi, commands.ChangeHost,
		commands.CollectLogs, commands.MountSmb, commands.UmountSmb,
		commands.CreateUpgradeTarget, commands.SetSSHPassword,
		commands.MountNfs, commands.UmountNfs:
		return nil
	}

	return errors.New("olares is uninitialized, cannot perform the operation")
}

// initializing
type InitializingValidator struct{}

func (u InitializingValidator) ValidateOp(op commands.Interface) error {
	switch op.OperationName() {
	case commands.ChangeIp, commands.Reboot,
		commands.Shutdown, commands.Uninstall,
		commands.ConnectWifi, commands.ChangeHost,
		commands.CollectLogs, commands.MountSmb, commands.UmountSmb, commands.SetSSHPassword,
		commands.MountNfs, commands.UmountNfs:
		return nil
	}

	return errors.New("olares is initializing, cannot perform the operation")
}

type UpgradingValidator struct{}

func (u UpgradingValidator) ValidateOp(op commands.Interface) error {
	switch op.OperationName() {
	case commands.ChangeIp, commands.Reboot,
		commands.Shutdown, commands.Uninstall,
		commands.ConnectWifi, commands.ChangeHost,
		commands.CollectLogs, commands.MountSmb, commands.UmountSmb,
		commands.CreateUpgradeTarget, commands.RemoveUpgradeTarget, commands.SetSSHPassword,
		commands.MountNfs, commands.UmountNfs:
		return nil
	}

	return errors.New("olares is upgrading, cannot perform the operation")
}

// initializeFailed
type InitializeFailedValidator struct{}

func (u InitializeFailedValidator) ValidateOp(op commands.Interface) error {
	switch op.OperationName() {
	case commands.ChangeIp, commands.Reboot,
		commands.Shutdown, commands.Uninstall,
		commands.ConnectWifi, commands.ChangeHost,
		commands.CollectLogs, commands.MountSmb, commands.UmountSmb, commands.SetSSHPassword,
		commands.MountNfs, commands.UmountNfs:
		return nil
	}

	return errors.New("olares is initialize failed, cannot perform the operation")
}

// Installing
type InstallingValidator struct{}

func (i InstallingValidator) ValidateOp(op commands.Interface) error {
	switch op.OperationName() {
	case commands.ChangeIp, commands.Reboot, commands.Shutdown, commands.SetSSHPassword:
		return nil
	}

	return errors.New("olares is Installing, cannot perform the operation")
}

// Install-failed
type InstallFailedValidator struct{}

func (i InstallFailedValidator) ValidateOp(op commands.Interface) error {
	switch op.OperationName() {
	case commands.Reboot, commands.Shutdown, commands.Uninstall,
		commands.CollectLogs, commands.MountSmb, commands.UmountSmb, commands.SetSSHPassword:
		return nil
	}

	return errors.New("olares installation is failed , cannot perform the operation")
}

// terminus-running
type RunningValidator struct{}

func (r RunningValidator) ValidateOp(op commands.Interface) error {
	switch op.OperationName() {
	case commands.ChangeIp, commands.Reboot, commands.Shutdown,
		commands.Uninstall, commands.ConnectWifi, commands.ChangeHost,
		commands.UmountUsb, commands.CollectLogs, commands.MountSmb, commands.UmountSmb,
		commands.CreateUpgradeTarget, commands.RemoveUpgradeTarget, commands.SetSSHPassword,
		commands.MountNfs, commands.UmountNfs:
		return nil
	}

	return errors.New("olares is running, cannot perform the operation")
}

// invalid-ip-address
type InvalidIpValidator struct{}

func (i InvalidIpValidator) ValidateOp(op commands.Interface) error {
	switch op.OperationName() {
	case commands.ChangeIp, commands.Reboot, commands.Shutdown,
		commands.Uninstall, commands.ConnectWifi, commands.ChangeHost,
		commands.MountSmb, commands.UmountSmb, commands.SetSSHPassword,
		commands.MountNfs, commands.UmountNfs:
		return nil
	}

	return errors.New("olares' ip has been changed, cannot perform the operation")
}

// system-error
type SystemErrorValidator struct{}

func (s SystemErrorValidator) ValidateOp(op commands.Interface) error {
	switch op.OperationName() {
	case commands.ChangeIp, commands.Reboot, commands.Shutdown,
		commands.Uninstall, commands.ConnectWifi, commands.ChangeHost,
		commands.CollectLogs, commands.MountSmb, commands.UmountSmb,
		commands.CreateUpgradeTarget, commands.RemoveUpgradeTarget, commands.SetSSHPassword,
		commands.MountNfs, commands.UmountNfs:
		return nil
	}

	return errors.New("olares is in the abnormal state, cannot perform the operation")
}

// self-repairing
type SelfRepairingValidator struct{}

func (s SelfRepairingValidator) ValidateOp(op commands.Interface) error {
	switch op.OperationName() {
	case commands.Reboot, commands.Shutdown, commands.Uninstall,
		commands.ConnectWifi, commands.ChangeHost, commands.SetSSHPassword:
		return nil
	}

	return errors.New("olares is in the self-repairing state, cannot perform the operation")
}

// ip-changing
type IpChangingValidator struct{}

func (i IpChangingValidator) ValidateOp(op commands.Interface) error {
	switch op.OperationName() {
	case commands.Reboot, commands.Shutdown, commands.SetSSHPassword:
		return nil
	}

	return errors.New("olares is in the ip-changing state, cannot perform the operation")
}

// ip-change-failed
type IpChangeFailedValidator struct{}

func (i IpChangeFailedValidator) ValidateOp(op commands.Interface) error {
	switch op.OperationName() {
	case commands.Reboot, commands.Shutdown, commands.Uninstall, commands.SetSSHPassword:
		return nil
	}

	return errors.New("olares is in the ip-change-failed state, cannot perform the operation")
}

// adding-node
type AddingNodeValidator struct{}

func (i AddingNodeValidator) ValidateOp(op commands.Interface) error {
	return errors.New("olares is adding node, cannot perform the operation")
}

// removing-node
type RemovingNodeValidator struct{}

func (i RemovingNodeValidator) ValidateOp(op commands.Interface) error {
	return errors.New("olares is removing node, cannot perform the operation")
}

// uninstalling
type UninstallingValidator struct{}

func (i UninstallingValidator) ValidateOp(op commands.Interface) error {
	return errors.New("olares is uninstalling, cannot perform the operation")
}

// disk-modifing
type DiskModifingValidator struct{}

func (i DiskModifingValidator) ValidateOp(op commands.Interface) error {
	return errors.New("olares is modifing the disk, cannot perform the operation")
}

// restarting
type RestartingValidator struct{}

func (i RestartingValidator) ValidateOp(op commands.Interface) error {
	return errors.New("olares is restaring, cannot perform the operation")
}

// shutdown
type ShutdownValidator struct{}

func (i ShutdownValidator) ValidateOp(op commands.Interface) error {
	return errors.New("olares is shuting down, cannot perform the operation")
}

type UnknownStateValidator struct{}

func (n UnknownStateValidator) ValidateOp(op commands.Interface) error {
	return errors.New("olares status is unknown, cannot perform the operation")
}
