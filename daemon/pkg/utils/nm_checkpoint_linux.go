//go:build linux
// +build linux

package utils

import (
	"context"
	"time"

	"github.com/godbus/dbus/v5"
	"k8s.io/klog/v2"
)

// NetworkManager D-Bus checkpoint API. Checkpoints are the only NetworkManager
// primitive that provides crash-safe, all-or-nothing rollback of a batch of
// device/connection changes: NetworkManager snapshots the selected devices and
// their connection settings, and if the checkpoint is neither destroyed
// (committed) nor rolled back within rollback_timeout, NetworkManager itself
// restores the snapshot even if the caller (olaresd) has crashed or exited.
// This is not exposed by nmcli, so we talk to the D-Bus API directly.
const (
	nmDBusDest      = "org.freedesktop.NetworkManager"
	nmDBusPath      = dbus.ObjectPath("/org/freedesktop/NetworkManager")
	nmDBusInterface = "org.freedesktop.NetworkManager"

	// NMCheckpointCreateFlags
	// https://networkmanager.dev/docs/api/latest/nm-dbus-types.html#NMCheckpointCreateFlags
	nmCheckpointCreateFlagDestroyAll           uint32 = 0x01
	nmCheckpointCreateFlagDeleteNewConnections uint32 = 0x02
	nmCheckpointCreateFlagDisconnectNewDevices uint32 = 0x04
)

// nmSystemBus opens a private connection to the system bus. The caller owns the
// connection and must Close it; the checkpoint is kept alive by NetworkManager
// (bounded by its rollback timeout), not by this connection.
func nmSystemBus() (*dbus.Conn, error) {
	return dbus.ConnectSystemBus()
}

// nmCheckpointCreate snapshots all devices and returns the checkpoint object
// path. If the checkpoint is not destroyed or rolled back within timeout,
// NetworkManager automatically rolls the whole network state back to this
// snapshot, deleting any connections created afterwards and reconnecting the
// original devices.
func nmCheckpointCreate(ctx context.Context, conn *dbus.Conn, timeout time.Duration) (dbus.ObjectPath, error) {
	obj := conn.Object(nmDBusDest, nmDBusPath)
	// empty device list => all devices
	devices := []dbus.ObjectPath{}
	flags := nmCheckpointCreateFlagDestroyAll |
		nmCheckpointCreateFlagDeleteNewConnections |
		nmCheckpointCreateFlagDisconnectNewDevices
	secs := uint32(timeout / time.Second)

	var cp dbus.ObjectPath
	if err := obj.CallWithContext(ctx, nmDBusInterface+".CheckpointCreate", 0,
		devices, secs, flags).Store(&cp); err != nil {
		return "", err
	}
	return cp, nil
}

// nmCheckpointDestroy commits the changes by discarding the checkpoint. After a
// successful destroy NetworkManager keeps the current (bridged) state.
func nmCheckpointDestroy(ctx context.Context, conn *dbus.Conn, cp dbus.ObjectPath) error {
	if cp == "" {
		return nil
	}
	obj := conn.Object(nmDBusDest, nmDBusPath)
	return obj.CallWithContext(ctx, nmDBusInterface+".CheckpointDestroy", 0, cp).Err
}

// nmCheckpointRollback restores the snapshot captured by the checkpoint,
// undoing every device/connection change made since it was created.
func nmCheckpointRollback(ctx context.Context, conn *dbus.Conn, cp dbus.ObjectPath) error {
	if cp == "" {
		return nil
	}
	obj := conn.Object(nmDBusDest, nmDBusPath)
	// result maps each device path to a per-device rollback result code.
	var result map[dbus.ObjectPath]uint32
	if err := obj.CallWithContext(ctx, nmDBusInterface+".CheckpointRollback", 0, cp).Store(&result); err != nil {
		return err
	}
	for dev, code := range result {
		if code != 0 {
			klog.Warningf("checkpoint rollback for device %s returned code %d", dev, code)
		}
	}
	return nil
}
