// Package portal forwards NetworkManager connectivity changes to the desktop.
package portal

import (
	"context"
	"errors"
	"net/url"
	"runtime"
	"time"

	"github.com/godbus/dbus/v5"
	"k8s.io/klog/v2"
)

const DefaultService = "com.olares.desktop.PortalHelper"
const helperInterface = DefaultService
const helperPath dbus.ObjectPath = "/com/olares/desktop/PortalHelper"
const nmService = "org.freedesktop.NetworkManager"
const nmPath dbus.ObjectPath = "/org/freedesktop/NetworkManager"
const propertiesInterface = "org.freedesktop.DBus.Properties"

type target struct{ UUID, Interface, URL string }
type helper interface {
	open(context.Context, target) error
	close(context.Context, string) error
}

// attempted is retained even after a timeout: the helper may have accepted a
// call whose response was lost. It must then be closed on a network change.
type delivery struct {
	attempted target
	delivered bool
}

func (d *delivery) reconcile(ctx context.Context, h helper, want target) error {
	if d.attempted.UUID != "" && d.attempted != want {
		if err := h.close(ctx, d.attempted.UUID); err != nil {
			return err
		}
		d.attempted = target{}
		d.delivered = false
	}
	if want.UUID == "" || d.delivered {
		return nil
	}
	if !validURL(want.URL) {
		return errors.New("invalid networkmanager connectivity check uri")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	d.attempted = want
	if err := h.open(ctx, want); err != nil {
		return err
	}
	d.delivered = true
	return nil
}

type busHelper struct {
	conn    *dbus.Conn
	service string
}

func (h busHelper) open(ctx context.Context, t target) error {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return h.conn.Object(h.service, helperPath).CallWithContext(c, helperInterface+".OpenPortal", 0, t.UUID, t.Interface, t.URL).Err
}
func (h busHelper) close(ctx context.Context, id string) error {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return h.conn.Object(h.service, helperPath).CallWithContext(c, helperInterface+".ClosePortal", 0, id).Err
}

func get(ctx context.Context, conn *dbus.Conn, path dbus.ObjectPath, iface, name string, out any) error {
	var v dbus.Variant
	if err := conn.Object(nmService, path).CallWithContext(ctx, propertiesInterface+".Get", 0, iface, name).Store(&v); err != nil {
		return err
	}
	return v.Store(out)
}
func storeProperty(props map[string]dbus.Variant, key string, out any) error {
	v, ok := props[key]
	if !ok || v.Value() == nil {
		return errors.New("networkmanager property is unavailable")
	}
	return v.Store(out)
}
func readTarget(ctx context.Context, conn *dbus.Conn) (target, error) {
	var props map[string]dbus.Variant
	if err := conn.Object(nmService, nmPath).CallWithContext(ctx, propertiesInterface+".GetAll", 0, nmService).Store(&props); err != nil {
		return target{}, err
	}
	var connectivity uint32
	if err := storeProperty(props, "Connectivity", &connectivity); err != nil {
		return target{}, err
	}
	if connectivity != 2 {
		return target{}, nil
	}
	var path dbus.ObjectPath
	if err := storeProperty(props, "PrimaryConnection", &path); err != nil {
		return target{}, err
	}
	if path == "/" || path == "" {
		return target{}, nil
	}
	var t target
	if err := storeProperty(props, "ConnectivityCheckUri", &t.URL); err != nil {
		return target{}, err
	}
	const activeInterface = nmService + ".Connection.Active"
	if err := get(ctx, conn, path, activeInterface, "Uuid", &t.UUID); err != nil {
		return target{}, err
	}
	var devices []dbus.ObjectPath
	if err := get(ctx, conn, path, activeInterface, "Devices", &devices); err != nil {
		return target{}, err
	}
	if t.UUID == "" || len(devices) == 0 {
		return target{}, errors.New("portal connection has no uuid or device")
	}
	if err := get(ctx, conn, devices[0], nmService+".Device", "Interface", &t.Interface); err != nil {
		return target{}, err
	}
	return t, nil
}
func validURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != "" && u.User == nil
}
func nextBackoff(current time.Duration) time.Duration {
	if current < time.Second {
		return time.Second
	}
	if current >= 15*time.Second {
		return 30 * time.Second
	}
	return current * 2
}

// Run is independent of Bluetooth and cluster readiness. The returned channel
// lets main wait for subscriptions and in-flight calls to be released.
func Start(ctx context.Context, service string) <-chan struct{} {
	done := make(chan struct{})
	go func() { defer close(done); run(ctx, service) }()
	return done
}
func run(ctx context.Context, service string) {
	if runtime.GOOS != "linux" {
		return
	}
	if service == "" {
		service = DefaultService
	}
	var state delivery
	delay := time.Second
	for ctx.Err() == nil {
		conn, err := dbus.ConnectSystemBus()
		if err == nil {
			state.delivered = false
			err = watch(ctx, conn, service, &state)
			conn.Close()
		}
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			klog.Warning("Portal monitor disconnected; retrying system bus subscription")
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		delay = nextBackoff(delay)
	}
}

// relevantSignal filters unrelated NM properties so traffic counters cannot
// cancel a pending helper call or reset its retry backoff.
func relevantSignal(s *dbus.Signal, service string) (relevant, restarted bool) {
	if s == nil {
		return false, false
	}
	if s.Name == "org.freedesktop.DBus.NameOwnerChanged" && len(s.Body) == 3 {
		name, ok := s.Body[0].(string)
		if !ok || (name != nmService && name != service) {
			return false, false
		}
		return true, name == service
	}
	if s.Path != nmPath || s.Name != propertiesInterface+".PropertiesChanged" || len(s.Body) != 3 {
		return false, false
	}
	iface, ok := s.Body[0].(string)
	if !ok || iface != nmService {
		return false, false
	}
	changed, ok := s.Body[1].(map[string]dbus.Variant)
	if !ok {
		return false, false
	}
	invalidated, ok := s.Body[2].([]string)
	if !ok {
		return false, false
	}
	for _, key := range []string{"Connectivity", "PrimaryConnection", "ConnectivityCheckUri"} {
		if _, ok := changed[key]; ok {
			return true, false
		}
		for _, name := range invalidated {
			if name == key {
				return true, false
			}
		}
	}
	return false, false
}

func watch(ctx context.Context, conn *dbus.Conn, service string, state *delivery) error {
	signals := make(chan *dbus.Signal, 64)
	conn.Signal(signals)
	defer conn.RemoveSignal(signals)
	rules := [][]dbus.MatchOption{
		{dbus.WithMatchSender(nmService), dbus.WithMatchObjectPath(nmPath), dbus.WithMatchInterface(propertiesInterface), dbus.WithMatchMember("PropertiesChanged")},
		{dbus.WithMatchSender("org.freedesktop.DBus"), dbus.WithMatchInterface("org.freedesktop.DBus"), dbus.WithMatchMember("NameOwnerChanged"), dbus.WithMatchArg(0, nmService)},
		{dbus.WithMatchSender("org.freedesktop.DBus"), dbus.WithMatchInterface("org.freedesktop.DBus"), dbus.WithMatchMember("NameOwnerChanged"), dbus.WithMatchArg(0, service)},
	}
	for _, rule := range rules {
		c, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := conn.AddMatchSignalContext(c, rule...)
		cancel()
		if err != nil {
			return err
		}
	}
	return watchEvents(ctx, conn.Context(), signals, service, state, busHelper{conn, service}, func(ctx context.Context) (target, error) { return readTarget(ctx, conn) })
}

// watchEvents owns delivery state; a new signal cancels the current operation,
// and its worker is joined before the next one starts.
func watchEvents(ctx, busCtx context.Context, signals <-chan *dbus.Signal, service string, state *delivery, h helper, snapshot func(context.Context) (target, error)) error {
	results := make(chan error, 1)
	var workerCancel context.CancelFunc
	running, pending, resetDelivery := false, true, false
	delay := time.Duration(0)
	var retry <-chan time.Time
	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
		if running {
			workerCancel()
			<-results
		}
		if ctx.Err() != nil && state.attempted.UUID != "" {
			cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = state.reconcile(cleanup, h, target{})
		}
	}()
	for {
		if pending && !running {
			if resetDelivery {
				state.delivered = false
				resetDelivery = false
			}
			workerCtx, cancel := context.WithCancel(ctx)
			workerCancel = cancel
			running, pending = true, false
			go func() {
				c, cancel := context.WithTimeout(workerCtx, 5*time.Second)
				want, err := snapshot(c)
				cancel()
				if err == nil {
					err = state.reconcile(workerCtx, h, want)
				}
				results <- err
			}()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-busCtx.Done():
			return errors.New("system bus disconnected")
		case signal, ok := <-signals:
			if !ok {
				return errors.New("system bus signal stream closed")
			}
			relevant, restarted := relevantSignal(signal, service)
			if !relevant {
				continue
			}
			pending = true
			resetDelivery = resetDelivery || restarted
			if running {
				workerCancel()
			}
			if timer != nil {
				timer.Stop()
			}
			retry = nil
			delay = 0
		case err := <-results:
			running = false
			workerCancel()
			if pending {
				continue
			}
			if err != nil {
				// Never log remote text or URL queries, which may contain login tokens.
				klog.Warning("Failed to synchronize captive portal state; retrying")
				delay = nextBackoff(delay)
				timer = time.NewTimer(delay)
				retry = timer.C
			} else {
				delay = 0
			}
		case <-retry:
			retry = nil
			pending = true
		}
	}
}
