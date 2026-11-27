// Package exposeports allocates node-facing application ports before Helm runs.
package exposeports

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReservationAnnotation marks an install's allocated ports in Spec.Config.
const ReservationAnnotation = "traffic.olares.io/exposed-ports-reserved"

const firstDynamicPort = 46800
const lastDynamicPort = 49999

type portKey struct {
	port     int32
	protocol string
}

type portSet map[portKey]struct{}

type inventory struct {
	applications []appv1.Application
	managers     []appv1.ApplicationManager
}

type allocator struct {
	gate     chan struct{}
	list     func(context.Context) (inventory, error)
	readNode func() (portSet, error)
	start    func() int32
}

// The app-service StatefulSet has one replica. Both install and upgrade use
// this gate, holding it until the assignment is stored in ApplicationManager.
// A replacement Pod reconstructs reservations from the API before allocating.
var liveAllocator = allocator{
	gate:     make(chan struct{}, 1),
	list:     listInventory,
	readNode: func() (portSet, error) { return readNodeSockets(nodeSocketDir) },
	start:    func() int32 { return firstDynamicPort + rand.Int32N(lastDynamicPort-firstDynamicPort+1) },
}

// Reserve passes a copy with assigned ports to persist. The callback must save
// it in the owning ApplicationManager before returning success.
func Reserve(ctx context.Context, owner *appv1.ApplicationManager, cfg *appcfg.ApplicationConfig, previous []appcfg.ServicePort, persist func(context.Context, *appcfg.ApplicationConfig) error) error {
	if len(cfg.Ports) == 0 {
		return persist(ctx, cfg)
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return liveAllocator.reserve(ctx, owner, cfg, previous, persist)
}

func (a *allocator) reserve(ctx context.Context, owner *appv1.ApplicationManager, cfg *appcfg.ApplicationConfig, previous []appcfg.ServicePort, persist func(context.Context, *appcfg.ApplicationConfig) error) error {
	select {
	case a.gate <- struct{}{}:
		defer func() { <-a.gate }()
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	inv, err := a.list(ctx)
	if err != nil {
		return fmt.Errorf("load exposed port reservations: %w", err)
	}
	busy, own, err := inv.reservations(owner)
	if err != nil {
		return err
	}
	node, err := a.readNode()
	if err != nil {
		return err
	}

	next := *cfg
	next.Ports = slices.Clone(cfg.Ports)
	next.TailScale.ACLs = slices.Clone(cfg.TailScale.ACLs)
	if err := assign(next.Ports, previous, busy, own, node, a.start()); err != nil {
		return err
	}
	for _, port := range next.Ports {
		if port.AddToTailscaleAcl {
			acl := appv1.ACL{Action: "accept", Src: []string{"*"}, Proto: port.Protocol, Dst: []string{fmt.Sprintf("*:%d", port.ExposePort)}}
			if !slices.ContainsFunc(next.TailScale.ACLs, func(existing appv1.ACL) bool {
				return existing.Action == acl.Action && existing.Proto == acl.Proto &&
					slices.Equal(existing.Src, acl.Src) && slices.Equal(existing.Dst, acl.Dst)
			}) {
				next.TailScale.ACLs = append(next.TailScale.ACLs, acl)
			}
		}
	}
	return persist(ctx, &next)
}

func listInventory(ctx context.Context) (inventory, error) {
	client, err := utils.GetClient()
	if err != nil {
		return inventory{}, err
	}
	api := client.AppV1alpha1()
	apps, err := api.Applications().List(ctx, metav1.ListOptions{})
	if err != nil {
		return inventory{}, err
	}
	managers, err := api.ApplicationManagers().List(ctx, metav1.ListOptions{})
	if err != nil {
		return inventory{}, err
	}
	return inventory{apps.Items, managers.Items}, nil
}

func protocols(protocol string) ([]string, error) {
	switch p := strings.ToLower(strings.TrimSpace(protocol)); p {
	case "", "tcp/udp":
		return []string{"tcp", "udp"}, nil
	case "tcp", "udp":
		return []string{p}, nil
	default:
		return nil, fmt.Errorf("unsupported exposed port protocol %q", protocol)
	}
}

func addPorts(set portSet, ports []appcfg.ServicePort) {
	for _, port := range ports {
		if port.ExposePort == 0 {
			continue
		}
		protos, err := protocols(port.Protocol)
		if err != nil {
			// Older versions accepted unvalidated fixed ports. Protect the
			// number without blocking unrelated installs.
			protos = []string{"tcp", "udp"}
		}
		mark(set, port.ExposePort, protos)
	}
}

func (inv inventory) reservations(owner *appv1.ApplicationManager) (portSet, portSet, error) {
	busy, own := make(portSet), make(portSet)
	for _, app := range inv.applications {
		set := busy
		if app.Spec.Namespace == owner.Spec.AppNamespace && app.Spec.Name == owner.Spec.AppName && owner.Spec.AppNamespace != "" {
			set = own
		}
		addPorts(set, app.Spec.Ports)
	}
	for _, am := range inv.managers {
		if am.Name == owner.Name || am.Spec.Config == "" {
			continue
		}
		switch am.Status.State {
		case appv1.Uninstalled, appv1.PendingCanceled, appv1.DownloadingCanceled, appv1.DownloadFailed, appv1.InstallingCanceled:
			continue
		}
		if am.Spec.OpType == appv1.InstallOp && am.Annotations[ReservationAnnotation] != "true" {
			// Older Pods could already have patched the assignment before
			// this annotation existed. Keep reservations past allocation.
			phase := am.Status.Payload[handlers.PayloadInstallingPhase]
			switch am.Status.State {
			case "", appv1.Pending, appv1.Downloading:
				continue
			case appv1.Installing:
				if phase == "" || phase == handlers.PhaseInstallingExposePorts {
					continue
				}
			}
		}
		for _, raw := range []string{am.Spec.Config, am.Annotations[api.AppPreUpgradeConfigKey]} {
			if raw == "" {
				continue
			}
			var cfg struct{ Ports []appcfg.ServicePort }
			if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
				return nil, nil, fmt.Errorf("application manager %s has invalid port configuration", am.Name)
			}
			addPorts(busy, cfg.Ports)
		}
	}
	return busy, own, nil
}

func assign(ports, previous []appcfg.ServicePort, busy, own, node portSet, start int32) error {
	priorByName := make(map[string]int32)
	for _, port := range previous {
		if port.Name != "" && port.ExposePort != 0 {
			priorByName[port.Name] = port.ExposePort
		}
	}
	for i := range ports {
		if port, ok := priorByName[ports[i].Name]; ok {
			ports[i].ExposePort = port
		}
	}
	for i := range ports {
		port := &ports[i]
		if port.ExposePort == 0 {
			continue
		}
		if port.ExposePort < 1 || port.ExposePort > 65535 {
			return fmt.Errorf("invalid exposed port %d for %q", port.ExposePort, port.Name)
		}
		protos, err := protocols(port.Protocol)
		if err != nil {
			return err
		}
		if !available(port.ExposePort, protos, busy, own, node) {
			return fmt.Errorf("exposed port %d for %q (%s) is already in use", port.ExposePort, port.Name, strings.Join(protos, "/"))
		}
		mark(busy, port.ExposePort, protos)
	}
	for i := range ports {
		port := &ports[i]
		if port.ExposePort != 0 {
			continue
		}
		protos, err := protocols(port.Protocol)
		if err != nil {
			return err
		}
		for n := int32(0); n <= lastDynamicPort-firstDynamicPort; n++ {
			number := firstDynamicPort + (start-firstDynamicPort+n)%(lastDynamicPort-firstDynamicPort+1)
			if available(number, protos, busy, nil, node) {
				port.ExposePort = number
				mark(busy, number, protos)
				break
			}
		}
		if port.ExposePort == 0 {
			return fmt.Errorf("no exposed port in %d-%d available for %q (%s)", firstDynamicPort, lastDynamicPort, port.Name, strings.Join(protos, "/"))
		}
	}
	return nil
}

func available(number int32, protos []string, busy, own, node portSet) bool {
	for _, proto := range protos {
		key := portKey{number, proto}
		if _, reserved := busy[key]; reserved {
			return false
		}
		if _, occupied := node[key]; occupied {
			if _, alreadyOwned := own[key]; !alreadyOwned {
				return false
			}
		}
	}
	return true
}

func mark(set portSet, number int32, protos []string) {
	for _, proto := range protos {
		set[portKey{number, proto}] = struct{}{}
	}
}
