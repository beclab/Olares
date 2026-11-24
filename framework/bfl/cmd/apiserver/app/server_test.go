package app

import (
	"net"
	"testing"
)

func TestCheckLocalAPIRequiresListeningSocket(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	address := listener.Addr().String()
	if err := checkLocalAPI(address); err != nil {
		t.Fatalf("listening API was unhealthy: %v", err)
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	if err := checkLocalAPI(address); err == nil {
		t.Fatal("stopped local API reported healthy")
	}
}
