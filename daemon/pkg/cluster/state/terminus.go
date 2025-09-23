package state

import (
	"errors"

	"github.com/beclab/Olares/daemon/pkg/commands"
)

type ProcessingState string

const (
	Completed  ProcessingState = "completed"
	Failed     ProcessingState = "failed"
	InProgress ProcessingState = "in-progress"
)

type TerminusState string

const (
	NotInstalled     TerminusState = "not-installed"
	Installing       TerminusState = "installing"
	InstallFailed    TerminusState = "install-failed"
	Uninitialized    TerminusState = "uninitialized"
	Initializing     TerminusState = "initializing"
	InitializeFailed TerminusState = "initialize-failed"
	TerminusRunning  TerminusState = "terminus-running"
	InvalidIpAddress TerminusState = "invalid-ip-address"
	SystemError      TerminusState = "system-error"
	SelfRepairing    TerminusState = "self-repairing"
	IPChanging       TerminusState = "ip-changing"
	IPChangeFailed   TerminusState = "ip-change-failed"
	AddingNode       TerminusState = "adding-node"
	RemovingNode     TerminusState = "removing-node"
	Uninstalling     TerminusState = "uninstalling"
	Upgrading        TerminusState = "upgrading"
	DiskModifing     TerminusState = "disk-modifing"
	Shutdown         TerminusState = "shutdown"
	Restarting       TerminusState = "restarting"
	Checking         TerminusState = "checking"
	NetworkNotReady  TerminusState = "network-not-ready"
)

func (s TerminusState) String() string {
	return string(s)
}

func (s TerminusState) ValidateOp(op commands.Interface) error {
	return s.getValidator().ValidateOp(op)
}

func (s TerminusState) getValidator() Validator {
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
		commands.MountSmb, commands.UmountSmb, commands.SetSSHPassword:
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
		commands.CreateUpgradeTarget, commands.SetSSHPassword:
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
		commands.CollectLogs, commands.MountSmb, commands.UmountSmb, commands.SetSSHPassword:
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
		commands.CreateUpgradeTarget, commands.RemoveUpgradeTarget, commands.SetSSHPassword:
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
		commands.CollectLogs, commands.MountSmb, commands.UmountSmb, commands.SetSSHPassword:
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
		commands.CreateUpgradeTarget, commands.RemoveUpgradeTarget, commands.SetSSHPassword:
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
		commands.MountSmb, commands.UmountSmb, commands.SetSSHPassword:
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
		commands.CreateUpgradeTarget, commands.RemoveUpgradeTarget, commands.SetSSHPassword:
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
	case commands.Reboot, commands.Shutdown, commands.Uninstall, commands.SetSSHPassword:
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
