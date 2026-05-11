package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/beclab/Olares/cli/cmd/ctl"
)

func main() {
	// Install a SIGINT/SIGTERM handler that cancels ctx on the first
	// signal and hard-exits on the second. This is what the dead
	// pkg/signals.SetupSignalHandler used to do, now expressed with
	// stdlib primitives.
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		// Wait for a second signal -- if it arrives, exit immediately
		// instead of letting goroutines hang on cleanup.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Fprintln(os.Stderr, "received second signal, force exiting")
		os.Exit(1)
	}()

	cmd := ctl.NewDefaultCommand()

	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
