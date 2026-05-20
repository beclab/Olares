package config

import "testing"

func TestDemoMeshDebugEnabled(t *testing.T) {
	d := Defaults{}
	d.Demo.Enabled = true
	d.Demo.MeshDebug = true
	if !d.DemoMeshDebugEnabled() {
		t.Fatal("expected true when demo enabled and meshDebug true")
	}

	d.Demo.MeshDebug = false
	if d.DemoMeshDebugEnabled() {
		t.Fatal("expected false when meshDebug false")
	}

	d.Demo.MeshDebug = true
	d.Demo.Enabled = false
	if d.DemoMeshDebugEnabled() {
		t.Fatal("expected false when demo disabled")
	}
}
