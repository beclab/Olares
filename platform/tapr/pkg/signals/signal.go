package signals

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

var onlyOneSignalHandler = make(chan struct{})
var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

// SetupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func SetupSignalHandler(ctx context.Context, cancel context.CancelFunc) (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		for {
			select {
			case <-c:
				close(stop)
				cancel()
				<-c
				os.Exit(1) // second signal. Exit directly.
			case <-ctx.Done():
				close(stop)
				os.Exit(1)
			}
		}
	}()

	return stop
}
