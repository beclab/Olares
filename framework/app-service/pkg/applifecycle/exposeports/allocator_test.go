package exposeports

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAssignUsesProtocolSpecificOccupancy(t *testing.T) {
	const start = firstDynamicPort
	node := portSet{{start, "tcp"}: {}, {start + 1, "udp"}: {}}
	reserved := portSet{{start + 2, "udp"}: {}}
	ports := []appcfg.ServicePort{
		{Name: "rdp-tcp-udp"},
		{Name: "tcp", Protocol: "tcp"},
		{Name: "udp", Protocol: "udp"},
		{Name: "explicit", Protocol: "udp", ExposePort: start + 3},
	}
	if err := assign(ports, nil, reserved, nil, node, start); err != nil {
		t.Fatal(err)
	}
	for i, want := range []int32{start + 4, start + 1, start, start + 3} {
		if ports[i].ExposePort != want {
			t.Errorf("%s got %d, want %d", ports[i].Name, ports[i].ExposePort, want)
		}
	}
}

func TestAssignPreservesExistingPortAndChecksNewProtocol(t *testing.T) {
	previous := []appcfg.ServicePort{{Name: "rdp", Protocol: "tcp", ExposePort: 47000}}
	own := portSet{{47000, "tcp"}: {}}
	node := portSet{{47000, "tcp"}: {}, {47000, "udp"}: {}}
	if err := assign([]appcfg.ServicePort{{Name: "rdp", Protocol: "tcp"}}, previous, portSet{}, own, node, firstDynamicPort); err != nil {
		t.Fatalf("existing TCP listener should retain its number: %v", err)
	}
	if err := assign([]appcfg.ServicePort{{Name: "rdp"}}, previous, portSet{}, own, node, firstDynamicPort); err == nil {
		t.Fatal("a new UDP listener would collide with the node")
	}
}

func TestAssignRejectsFixedConflicts(t *testing.T) {
	tests := []struct {
		name  string
		ports []appcfg.ServicePort
		node  portSet
		busy  portSet
	}{
		{"node listener", []appcfg.ServicePort{{Name: "rdp", ExposePort: 47000}}, portSet{{47000, "udp"}: {}}, nil},
		{"platform reservation", []appcfg.ServicePort{{Name: "rdp", ExposePort: 47000}}, nil, portSet{{47000, "tcp"}: {}}},
		{"same application", []appcfg.ServicePort{{Name: "a", ExposePort: 47000}, {Name: "b", ExposePort: 47000, Protocol: "udp"}}, nil, nil},
		{"invalid protocol", []appcfg.ServicePort{{Name: "x", Protocol: "sctp"}}, nil, nil},
		{"invalid fixed port", []appcfg.ServicePort{{Name: "x", ExposePort: 65536}}, nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			busy := tt.busy
			if busy == nil {
				busy = make(portSet)
			}
			if err := assign(tt.ports, nil, busy, nil, tt.node, firstDynamicPort); err == nil {
				t.Fatal("expected an allocation error")
			}
		})
	}
}

func TestAssignWrapsAndReportsExhaustion(t *testing.T) {
	node := make(portSet)
	node[portKey{lastDynamicPort, "udp"}] = struct{}{}
	ports := []appcfg.ServicePort{{Name: "rdp"}}
	if err := assign(ports, nil, make(portSet), nil, node, lastDynamicPort); err != nil || ports[0].ExposePort != firstDynamicPort {
		t.Fatalf("wraparound allocation: %+v, %v", ports, err)
	}
	for n := int32(firstDynamicPort); n <= lastDynamicPort; n++ {
		node[portKey{n, "udp"}] = struct{}{}
	}
	if err := assign([]appcfg.ServicePort{{Name: "rdp"}}, nil, make(portSet), nil, node, lastDynamicPort); err == nil {
		t.Fatal("an exhausted UDP range must block dual protocol allocation")
	}
}

func TestReservationsRespectApplicationLifecycle(t *testing.T) {
	owner := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "current"},
		Spec:       appv1.ApplicationManagerSpec{AppName: "windows", AppNamespace: "windows-alice"},
	}
	inv := inventory{applications: []appv1.Application{
		{Spec: appv1.ApplicationSpec{Name: "windows", Namespace: "windows-alice", Ports: []appcfg.ServicePort{{ExposePort: 46800, Protocol: "tcp"}}}},
		{Spec: appv1.ApplicationSpec{Name: "windows", Namespace: "windows-bob", Ports: []appcfg.ServicePort{{ExposePort: 46801, Protocol: "legacy-unknown"}}}},
	}, managers: []appv1.ApplicationManager{
		*owner,
		reservation("provisioning", 46802, appv1.Installing),
		reservation("cleaned-up", 46803, appv1.Uninstalled),
		reservation("download-failed", 46804, appv1.DownloadFailed),
	}}
	busy, own, err := inv.reservations(owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(own) != 1 || len(busy) != 4 || !available(46803, []string{"tcp", "udp"}, busy, own, nil) {
		t.Fatalf("busy=%v own=%v", busy, own)
	}
	inv.managers[1].Spec.Config = "{broken"
	if _, _, err := inv.reservations(owner); err == nil {
		t.Fatal("unreadable active reservation should fail closed")
	}
}

func TestReservationsKeepOnlyAssignedAndRollbackPorts(t *testing.T) {
	proposed := reservation("proposed", 46800, appv1.Installing)
	delete(proposed.Annotations, ReservationAnnotation)
	proposed.Status.Payload = map[string]string{handlers.PayloadInstallingPhase: handlers.PhaseInstallingExposePorts}
	legacy := reservation("legacy-assigned", 46801, appv1.Installing)
	delete(legacy.Annotations, ReservationAnnotation)
	legacy.Status.Payload = map[string]string{handlers.PayloadInstallingPhase: handlers.PhaseInstallingHelmInstall}
	upgrade := reservation("upgrade", 46802, appv1.UpgradePending)
	upgrade.Spec.OpType = appv1.UpgradeOp
	upgrade.Annotations[api.AppPreUpgradeConfigKey] = reservation("rollback", 46803, appv1.Running).Spec.Config

	inv := inventory{managers: []appv1.ApplicationManager{proposed, legacy, upgrade}}
	busy, _, err := inv.reservations(&appv1.ApplicationManager{})
	if err != nil {
		t.Fatal(err)
	}
	if !available(46800, []string{"tcp", "udp"}, busy, nil, nil) {
		t.Fatal("an unallocated manifest entry blocked allocation")
	}
	for _, port := range []int32{46801, 46802, 46803} {
		if available(port, []string{"tcp", "udp"}, busy, nil, nil) {
			t.Errorf("active or rollback port %d was not reserved", port)
		}
	}
}

func reservation(name string, port int32, state appv1.ApplicationManagerState) appv1.ApplicationManager {
	return appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name, Annotations: map[string]string{ReservationAnnotation: "true"}},
		Spec:       appv1.ApplicationManagerSpec{OpType: appv1.InstallOp, Config: fmt.Sprintf(`{"Ports":[{"exposePort":%d}]}`, port)},
		Status:     appv1.ApplicationManagerStatus{State: state},
	}
}

func TestConcurrentReservationsPersistBeforeNextAllocation(t *testing.T) {
	const total = 24
	inv := inventory{}
	a := allocator{
		gate:     make(chan struct{}, 1),
		list:     func(context.Context) (inventory, error) { return inv, nil },
		readNode: func() (portSet, error) { return nil, nil },
		start:    func() int32 { return firstDynamicPort },
	}
	ports := make(chan int32, total)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < total; i++ {
		wg.Go(func() {
			<-start
			name := fmt.Sprintf("manager-%d", i)
			am := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: name}}
			cfg := &appcfg.ApplicationConfig{Ports: []appcfg.ServicePort{{Name: "rdp", AddToTailscaleAcl: true}}}
			err := a.reserve(context.Background(), am, cfg, nil, func(_ context.Context, assigned *appcfg.ApplicationConfig) error {
				number := assigned.Ports[0].ExposePort
				raw, err := json.Marshal(assigned)
				if err != nil {
					return err
				}
				if got := assigned.TailScale.ACLs; len(got) != 1 || len(got[0].Dst) != 1 || got[0].Dst[0] != fmt.Sprintf("*:%d", number) {
					return fmt.Errorf("missing assigned port in ACL: %v", got)
				}
				record := reservation(name, number, appv1.Installing)
				record.Spec.Config = string(raw)
				inv.managers = append(inv.managers, record)
				ports <- number
				return nil
			})
			if err != nil || cfg.Ports[0].ExposePort != 0 || len(cfg.TailScale.ACLs) != 0 {
				t.Errorf("allocation=%v mutated input=%+v", err, cfg.Ports)
			}
		})
	}
	close(start)
	wg.Wait()
	close(ports)
	seen := make(map[int32]bool)
	for number := range ports {
		if seen[number] {
			t.Errorf("port %d was reserved twice", number)
		}
		seen[number] = true
	}
	if len(seen) != total {
		t.Fatalf("got %d assignments, want %d", len(seen), total)
	}
}

func TestReservationFailureAndCancellationDoNotConsumePort(t *testing.T) {
	a := allocator{
		gate:     make(chan struct{}, 1),
		list:     func(context.Context) (inventory, error) { return inventory{}, nil },
		readNode: func() (portSet, error) { return nil, nil },
		start:    func() int32 { return firstDynamicPort },
	}
	am := &appv1.ApplicationManager{}
	cfg := &appcfg.ApplicationConfig{Ports: []appcfg.ServicePort{{Name: "rdp"}}}
	failure := errors.New("persist failed")
	if err := a.reserve(context.Background(), am, cfg, nil, func(context.Context, *appcfg.ApplicationConfig) error { return failure }); !errors.Is(err, failure) {
		t.Fatalf("persist error lost: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	a.gate <- struct{}{}
	err := a.reserve(ctx, am, cfg, nil, func(context.Context, *appcfg.ApplicationConfig) error {
		t.Fatal("persist called after cancellation")
		return nil
	})
	<-a.gate
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("waiting for reservation: %v", err)
	}
	if err := a.reserve(context.Background(), am, cfg, nil, func(_ context.Context, assigned *appcfg.ApplicationConfig) error {
		if assigned.Ports[0].ExposePort != firstDynamicPort {
			t.Fatalf("failed reservation consumed port: %+v", assigned.Ports)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	a.readNode = func() (portSet, error) { return nil, failure }
	if err := a.reserve(context.Background(), am, cfg, nil, func(context.Context, *appcfg.ApplicationConfig) error {
		t.Fatal("persist called without node inventory")
		return nil
	}); !errors.Is(err, failure) {
		t.Fatalf("node inventory error lost: %v", err)
	}
}

func TestReserveWithoutPorts(t *testing.T) {
	cfg := &appcfg.ApplicationConfig{}
	if err := Reserve(context.Background(), nil, cfg, nil, func(_ context.Context, got *appcfg.ApplicationConfig) error {
		if got != cfg {
			t.Fatal("no-port install changed configuration")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
