package handlers

import (
	"os"
	"time"

	"k8s.io/klog/v2"
)

// OverlayGatewayOpLockTTL is longer than the enable/disable operation timer (60s)
// so a live in-flight lock is not treated as stale mid-operation.
const OverlayGatewayOpLockTTL = 90 * time.Second

const overlayGatewayOpLockStaleMsg = "overlay gateway operation lock expired"

// ClearOverlayGatewayOpLocks removes enable/disable in-flight lock files left
// behind after a crash. Does not touch application settings or cni-dhcp.
func ClearOverlayGatewayOpLocks() {
	clearOverlayGatewayOpLocks(OverlayGatewayEnableLockFile, OverlayGatewayDisableLockFile)
}

func clearOverlayGatewayOpLocks(paths ...string) {
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			klog.Errorf("overlay gateway: clear op lock %s failed: %v", path, err)
		}
	}
}

// consumeOverlayGatewayOpLock reports whether an in-flight op lock is present.
// Locks older than OverlayGatewayOpLockTTL are removed and reported via staleMsg
// so callers can surface an English machine message without treating them as
// still in progress.
func consumeOverlayGatewayOpLock(path string) (inFlight bool, staleMsg string) {
	info, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			klog.Errorf("overlay gateway: stat op lock %s failed: %v", path, err)
		}
		return false, ""
	}
	if time.Since(info.ModTime()) > OverlayGatewayOpLockTTL {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			klog.Errorf("overlay gateway: remove stale op lock %s failed: %v", path, err)
		} else {
			klog.Warningf("overlay gateway: removed stale op lock %s", path)
		}
		return false, overlayGatewayOpLockStaleMsg
	}
	return true, ""
}
