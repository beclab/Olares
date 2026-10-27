package portal

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

type fakeHelper struct {
	mu     sync.Mutex
	calls  []string
	fail   bool
	events chan string
	block  bool
}

func (f *fakeHelper) record(ctx context.Context, call string) error {
	f.mu.Lock()
	f.calls = append(f.calls, call)
	fail, block := f.fail, f.block
	f.mu.Unlock()
	if f.events != nil {
		f.events <- call
	}
	if block {
		<-ctx.Done()
		return ctx.Err()
	}
	if fail {
		return errors.New("helper unavailable")
	}
	return nil
}
func (f *fakeHelper) open(ctx context.Context, t target) error   { return f.record(ctx, "open:"+t.UUID) }
func (f *fakeHelper) close(ctx context.Context, id string) error { return f.record(ctx, "close:"+id) }
func TestDeliveryTransitions(t *testing.T) {
	ctx := context.Background()
	h := &fakeHelper{}
	d := delivery{}
	a := target{"a", "wlan0", "http://check.example.com"}
	b := target{"b", "eth0", "http://check.example.com"}
	for _, want := range []target{a, a, b, {}, {}} {
		if err := d.reconcile(ctx, h, want); err != nil {
			t.Fatal(err)
		}
	}
	expected := []string{"open:a", "close:a", "open:b", "close:b"}
	if len(h.calls) != len(expected) {
		t.Fatalf("unexpected calls: %v", h.calls)
	}
	for i, v := range expected {
		if h.calls[i] != v {
			t.Fatalf("unexpected calls: %v", h.calls)
		}
	}
}
func TestFailedOpenStillClosed(t *testing.T) {
	h := &fakeHelper{fail: true}
	d := delivery{}
	a := target{UUID: "a", URL: "http://check.example.com"}
	if d.reconcile(context.Background(), h, a) == nil {
		t.Fatal("expected failure")
	}
	if d.delivered {
		t.Fatal("failed call recorded as success")
	}
	h.fail = false
	if err := d.reconcile(context.Background(), h, target{}); err != nil {
		t.Fatal(err)
	}
	if len(h.calls) != 2 || h.calls[1] != "close:a" {
		t.Fatal("possibly opened window not closed")
	}
}
func TestFailedCloseRetriedBeforeOpen(t *testing.T) {
	h := &fakeHelper{}
	d := delivery{}
	ctx := context.Background()
	if err := d.reconcile(ctx, h, target{UUID: "a", URL: "http://check.example.com"}); err != nil {
		t.Fatal(err)
	}
	h.fail = true
	if d.reconcile(ctx, h, target{UUID: "b", URL: "http://check.example.com"}) == nil {
		t.Fatal("expected close failure")
	}
	h.fail = false
	if err := d.reconcile(ctx, h, target{UUID: "b", URL: "http://check.example.com"}); err != nil {
		t.Fatal(err)
	}
	if len(h.calls) != 4 || h.calls[2] != "close:a" || h.calls[3] != "open:b" {
		t.Fatal(h.calls)
	}
}
func changedSignal() *dbus.Signal {
	return &dbus.Signal{Path: nmPath, Name: propertiesInterface + ".PropertiesChanged", Body: []any{nmService, map[string]dbus.Variant{"Connectivity": dbus.MakeVariant(uint32(2))}, []string{}}}
}
func TestSignalFiltering(t *testing.T) {
	if ok, _ := relevantSignal(changedSignal(), DefaultService); !ok {
		t.Fatal("connectivity ignored")
	}
	signal := changedSignal()
	signal.Body[1] = map[string]dbus.Variant{}
	signal.Body[2] = []string{"PrimaryConnection"}
	if ok, _ := relevantSignal(signal, DefaultService); !ok {
		t.Fatal("invalidated property ignored")
	}
	signal.Body[2] = []string{"Version"}
	if ok, _ := relevantSignal(signal, DefaultService); ok {
		t.Fatal("unrelated property accepted")
	}
	signal = &dbus.Signal{Name: "org.freedesktop.DBus.NameOwnerChanged", Body: []any{DefaultService, ":1.1", ":1.2"}}
	if ok, restarted := relevantSignal(signal, DefaultService); !ok || !restarted {
		t.Fatal("helper restart ignored")
	}
	for _, s := range []*dbus.Signal{nil, {Name: propertiesInterface + ".PropertiesChanged"}, {Path: nmPath, Name: propertiesInterface + ".PropertiesChanged", Body: []any{nmService, 42, []string{}}}} {
		if ok, _ := relevantSignal(s, DefaultService); ok {
			t.Fatal("malformed signal accepted")
		}
	}
}
func awaitEvent(t *testing.T, events <-chan string, want string) {
	t.Helper()
	select {
	case got := <-events:
		if got != want {
			t.Fatalf("got %s want %s", got, want)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("waiting for %s", want)
	}
}
func TestWatcherInitialPortalRestartAndShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signals := make(chan *dbus.Signal, 10)
	h := &fakeHelper{events: make(chan string, 10)}
	done := make(chan error, 1)
	go func() {
		done <- watchEvents(ctx, context.Background(), signals, DefaultService, &delivery{}, h, func(context.Context) (target, error) { return target{UUID: "a", URL: "http://check.example.com"}, nil })
	}()
	awaitEvent(t, h.events, "open:a")
	signals <- &dbus.Signal{Name: "org.freedesktop.DBus.NameOwnerChanged", Body: []any{DefaultService, ":1.1", ":1.2"}}
	awaitEvent(t, h.events, "open:a")
	cancel()
	awaitEvent(t, h.events, "close:a")
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watcher did not stop")
	}
}
func TestWatcherCancelsStaleOpenAndRetries(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signals := make(chan *dbus.Signal, 10)
	h := &fakeHelper{events: make(chan string, 10), block: true}
	done := make(chan error, 1)
	var mu sync.Mutex
	want := target{UUID: "a", URL: "http://check.example.com"}
	go func() {
		done <- watchEvents(ctx, context.Background(), signals, DefaultService, &delivery{}, h, func(context.Context) (target, error) { mu.Lock(); defer mu.Unlock(); return want, nil })
	}()
	awaitEvent(t, h.events, "open:a")
	h.mu.Lock()
	h.block = false
	h.fail = true
	h.mu.Unlock()
	mu.Lock()
	want = target{UUID: "b", URL: "http://check.example.com"}
	mu.Unlock()
	signals <- changedSignal()
	awaitEvent(t, h.events, "close:a") // first close fails
	h.mu.Lock()
	h.fail = false
	h.mu.Unlock()
	awaitEvent(t, h.events, "close:a") // retry succeeds
	awaitEvent(t, h.events, "open:b")
	cancel()
	awaitEvent(t, h.events, "close:b")
	<-done
}
func TestWatcherBusDisconnectStopsWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus, cancelBus := context.WithCancel(context.Background())
	defer cancelBus()
	h := &fakeHelper{events: make(chan string, 2), block: true}
	done := make(chan error, 1)
	go func() {
		done <- watchEvents(ctx, bus, make(chan *dbus.Signal), DefaultService, &delivery{}, h, func(context.Context) (target, error) { return target{UUID: "a", URL: "http://check.example.com"}, nil })
	}()
	awaitEvent(t, h.events, "open:a")
	cancelBus()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("disconnect missing")
		}
	case <-time.After(time.Second):
		t.Fatal("worker leaked")
	}
}
func TestURLAndBackoff(t *testing.T) {
	for _, u := range []string{"http://example.com/check", "https://example.com/check"} {
		if !validURL(u) {
			t.Fatal(u)
		}
	}
	for _, u := range []string{"", "file:///etc/passwd", "javascript:alert(1)", "http://", "http://user:secret@example.com"} {
		if validURL(u) {
			t.Fatal(u)
		}
	}
	delay := time.Duration(0)
	for _, want := range []time.Duration{1, 2, 4, 8, 16, 30, 30} {
		delay = nextBackoff(delay)
		if delay != want*time.Second {
			t.Fatal(delay)
		}
	}
}

func TestMissingNMPropertiesReturnError(t *testing.T) {
	var state uint32
	if err := storeProperty(map[string]dbus.Variant{}, "Connectivity", &state); err == nil {
		t.Fatal("missing property accepted")
	}
	if err := storeProperty(map[string]dbus.Variant{"Connectivity": dbus.MakeVariant("bad")}, "Connectivity", &state); err == nil {
		t.Fatal("wrong property type accepted")
	}
}
func TestInvalidURLClosesPreviousWindow(t *testing.T) {
	d := delivery{}
	h := &fakeHelper{}
	ctx := context.Background()
	if err := d.reconcile(ctx, h, target{UUID: "a", URL: "http://check.example.com"}); err != nil {
		t.Fatal(err)
	}
	if err := d.reconcile(ctx, h, target{UUID: "b", URL: "file:///bad"}); err == nil {
		t.Fatal("invalid URL accepted")
	}
	if len(h.calls) != 2 || h.calls[1] != "close:a" {
		t.Fatal("stale window retained")
	}
}
