package portal

import (
	"bufio"
	"context"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

// Runs against a real private bus when dbus-daemon is available. It does not
// connect to the machine's system bus or modify its NetworkManager state.
func TestDBusPortalLifecycle(t *testing.T) {
	executable, err := exec.LookPath("dbus-daemon")
	if err != nil {
		t.Skip("dbus-daemon is not installed")
	}
	process := exec.Command(executable, "--session", "--nofork", "--print-address=1")
	stdout, err := process.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = process.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = process.Process.Kill(); _ = process.Wait() }()
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		t.Fatal("private bus did not return an address")
	}
	address := scanner.Text()
	service, err := dbus.Connect(address)
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	client, err := dbus.Connect(address)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	for _, name := range []string{nmService, DefaultService} {
		if _, err := service.RequestName(name, dbus.NameFlagDoNotQueue); err != nil {
			t.Fatal(err)
		}
	}
	nm := &mockProperties{values: map[string]dbus.Variant{
		"Connectivity": dbus.MakeVariant(uint32(2)), "PrimaryConnection": dbus.MakeVariant(dbus.ObjectPath("/active")), "ConnectivityCheckUri": dbus.MakeVariant("http://check.example.com"),
	}}
	active := &mockProperties{values: map[string]dbus.Variant{"Uuid": dbus.MakeVariant("connection-a"), "Devices": dbus.MakeVariant([]dbus.ObjectPath{"/device"})}}
	device := &mockProperties{values: map[string]dbus.Variant{"Interface": dbus.MakeVariant("wlan0")}}
	for path, props := range map[dbus.ObjectPath]*mockProperties{nmPath: nm, "/active": active, "/device": device} {
		if err := service.Export(props, path, propertiesInterface); err != nil {
			t.Fatal(err)
		}
	}
	helper := &exportedHelper{events: make(chan string, 10)}
	if err := service.Export(helper, helperPath, helperInterface); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- watch(ctx, client, DefaultService, &delivery{}) }()
	awaitEvent(t, helper.events, "open:connection-a:wlan0:http://check.example.com")
	nm.mu.Lock()
	nm.values["Connectivity"] = dbus.MakeVariant(uint32(4))
	nm.mu.Unlock()
	if err := service.Emit(nmPath, propertiesInterface+".PropertiesChanged", nmService, map[string]dbus.Variant{"Connectivity": dbus.MakeVariant(uint32(4))}, []string{}); err != nil {
		t.Fatal(err)
	}
	awaitEvent(t, helper.events, "close:connection-a")
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("private bus watcher did not stop")
	}
}

type mockProperties struct {
	mu     sync.Mutex
	values map[string]dbus.Variant
}

func (p *mockProperties) GetAll(_ string) (map[string]dbus.Variant, *dbus.Error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := map[string]dbus.Variant{}
	for k, v := range p.values {
		result[k] = v
	}
	return result, nil
}
func (p *mockProperties) Get(_, key string) (dbus.Variant, *dbus.Error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.values[key], nil
}

type exportedHelper struct{ events chan string }

func (h *exportedHelper) OpenPortal(uuid, iface, url string) *dbus.Error {
	h.events <- "open:" + uuid + ":" + iface + ":" + url
	return nil
}
func (h *exportedHelper) ClosePortal(uuid string) *dbus.Error {
	h.events <- "close:" + uuid
	return nil
}
