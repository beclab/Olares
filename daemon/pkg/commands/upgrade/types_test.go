package upgrade

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestAwaitExecutionDoesNotAcceptLogMarkerBeforeFailedExit(t *testing.T) {
	progress := make(chan int, 1)
	progress <- 100
	completion := make(chan error, 1)
	completion <- errors.New("CLI reported failure")

	err := AwaitExecution(context.Background(), newExecutionResWithCompletion(progress, completion), nil)
	if err == nil || !strings.Contains(err.Error(), "CLI reported failure") {
		t.Fatalf("expected CLI failure after log marker, got %v", err)
	}
}

func TestAwaitExecutionChecksFinalProgressAfterSuccessfulExit(t *testing.T) {
	progress := make(chan int, 1)
	progress <- 100
	close(progress)
	completion := make(chan error, 1)
	completion <- nil
	close(completion)

	var observed int
	if err := AwaitExecution(context.Background(), newExecutionResWithCompletion(progress, completion), func(p int) {
		observed = p
	}); err != nil || observed != 100 {
		t.Fatalf("expected success and 100%% progress, got progress=%d error=%v", observed, err)
	}
}

func TestAwaitExecutionReportsMissingLogMarker(t *testing.T) {
	progress := make(chan int)
	close(progress)
	completion := make(chan error, 1)
	completion <- nil
	close(completion)

	if err := AwaitExecution(context.Background(), newExecutionResWithCompletion(progress, completion), nil); err == nil {
		t.Fatal("successful exit without completion marker was accepted")
	}
}

func TestAwaitExecutionWaitsForLogAfterSuccessfulExit(t *testing.T) {
	progress := make(chan int)
	completion := make(chan error, 1)
	completion <- nil
	close(completion)

	done := make(chan error, 1)
	go func() {
		done <- AwaitExecution(context.Background(), newExecutionResWithCompletion(progress, completion), nil)
	}()
	progress <- 100
	close(progress)
	if err := <-done; err != nil {
		t.Fatalf("final progress after process exit was lost: %v", err)
	}
}
