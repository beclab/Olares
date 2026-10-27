//go:build windows

package clusterexec

import (
	"time"

	"golang.org/x/term"
)

// resizePollInterval is how often the Windows variant checks for a terminal
// size change. Windows has no SIGWINCH, so we poll; 500ms is responsive enough
// for a human dragging a window edge while staying effectively free.
const resizePollInterval = 500 * time.Millisecond

// watchTerminalResize invokes onResize whenever the local terminal size
// changes, until the returned stop function is called. Windows has no SIGWINCH,
// so this polls the terminal size on fd and fires onResize only when the
// dimensions actually change (avoiding a flood of redundant resize frames).
func watchTerminalResize(fd int, onResize func()) (stop func()) {
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(resizePollInterval)
		defer ticker.Stop()
		lastW, lastH, _ := term.GetSize(fd)
		for {
			select {
			case <-ticker.C:
				w, h, err := term.GetSize(fd)
				if err == nil && (w != lastW || h != lastH) {
					lastW, lastH = w, h
					onResize()
				}
			case <-done:
				return
			}
		}
	}()
	return func() { close(done) }
}
