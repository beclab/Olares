//go:build !windows

package clusterexec

import (
	"os"
	"os/signal"
	"syscall"
)

// watchTerminalResize invokes onResize whenever the local terminal size
// changes, until the returned stop function is called. On Unix this is driven
// by SIGWINCH, which the kernel delivers on every window-size change. The fd is
// unused here (the signal carries no size; onResize reads the current size
// itself), but kept in the signature so the Windows polling variant can use it.
func watchTerminalResize(_ int, onResize func()) (stop func()) {
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-winch:
				onResize()
			case <-done:
				return
			}
		}
	}()
	return func() {
		signal.Stop(winch)
		close(done)
	}
}
