package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	d, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if d.Namespace != "app-gateway" {
		t.Errorf("Namespace = %q, want app-gateway", d.Namespace)
	}
	if d.Gateway.Name != "app-gateway" {
		t.Errorf("Gateway.Name = %q, want app-gateway", d.Gateway.Name)
	}
	if d.Gateway.GatewayClassName != "olares-app-gateway" {
		t.Errorf("Gateway.GatewayClassName = %q, want olares-app-gateway", d.Gateway.GatewayClassName)
	}
	if !d.EnvoyProxy.Enabled {
		t.Error("EnvoyProxy.Enabled = false, want true")
	}
	if !d.TLS.Enabled {
		t.Error("TLS.Enabled = false, want true")
	}
}

func TestNamespace(t *testing.T) {
	if got := Namespace(); got != "app-gateway" {
		t.Errorf("Namespace() = %q, want app-gateway", got)
	}
}
