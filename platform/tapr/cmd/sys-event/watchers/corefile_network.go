package watchers

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/coredns/corefile-migration/migration/corefile"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var terminusGVR = schema.GroupVersionResource{Group: "sys.bytetrade.io", Version: "v1alpha1", Resource: "terminus"}

const sharedMeshZone = "shared.olares.com"

type machineDNS struct {
	domain    string
	localZone string
	localIP   string
}

func readMachineDNS(ctx context.Context, kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) (machineDNS, error) {
	dns := machineDNS{localZone: "olares.local"}
	terminus, err := dynamicClient.Resource(terminusGVR).Get(ctx, "terminus", metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return dns, err
	}
	if err == nil {
		domain, _, err := unstructured.NestedString(terminus.Object, "spec", "settings", "domainName")
		if err != nil {
			return dns, err
		}
		domain = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(domain), "."))
		if domain != "" && domain != "olares.local" {
			if errs := validation.IsDNS1123Subdomain(domain); len(errs) > 0 {
				return dns, fmt.Errorf("invalid default domain: %s", strings.Join(errs, "; "))
			}
			dns.domain = domain
		}
		localName, _, err := unstructured.NestedString(terminus.Object, "spec", "localName")
		if err != nil {
			return dns, err
		}
		if localName != "" {
			if errs := validation.IsDNS1123Label(localName); len(errs) > 0 {
				return dns, fmt.Errorf("invalid localName: %s", strings.Join(errs, "; "))
			}
			dns.localZone = localName + ".olares.local"
		}
	}

	cm, err := kubeClient.CoreV1().ConfigMaps("os-network").Get(ctx, "default-network", metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return dns, nil
	}
	if err != nil {
		return dns, err
	}
	if value := cm.Data["local-dns-ip"]; value != "" {
		ip := net.ParseIP(value)
		if ip == nil || ip.To4() == nil {
			return dns, fmt.Errorf("default-network local-dns-ip is not an IPv4 address")
		}
		dns.localIP = ip.String()
	}
	return dns, nil
}

func machineDNSTemplates(zone, ip string) []*corefile.Plugin {
	if zone == "" || ip == "" {
		return nil
	}
	return []*corefile.Plugin{
		{
			Name: "template", Args: []string{"IN", "A", zone},
			Options: []*corefile.Option{
				{Name: "match", Args: []string{`"^([^.]+\.)*` + regexp.QuoteMeta(zone) + `\.$"`}},
				{Name: "answer", Args: []string{fmt.Sprintf(`"{{ .Name }} 60 IN A %s"`, ip)}},
				{Name: "fallthrough"},
			},
		},
		{Name: "template", Args: []string{"IN", "ANY", zone}, Options: []*corefile.Option{{Name: "rcode", Args: []string{"NOERROR"}}}},
	}
}

// CorefileNetworkSubscriber keeps machine DNS in sync without depending on a
// User update or on a personal BFL pod being present.
type CorefileNetworkSubscriber struct {
	*Subscriber
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface
}

func machineDNSFields(obj interface{}) [3]string {
	var fields [3]string
	switch resource := obj.(type) {
	case *unstructured.Unstructured:
		fields[0], _, _ = unstructured.NestedString(resource.Object, "spec", "settings", "domainName")
		fields[1], _, _ = unstructured.NestedString(resource.Object, "spec", "localName")
	case *corev1.ConfigMap:
		fields[2] = resource.Data["local-dns-ip"]
	}
	return fields
}

func (s *CorefileNetworkSubscriber) HandleEvent() cache.ResourceEventHandler {
	enqueue := func(obj interface{}) {
		s.Watchers.Enqueue(EnqueueObj{Subscribe: s, Obj: obj, Action: UPDATE})
	}
	return cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch resource := obj.(type) {
			case *unstructured.Unstructured:
				return resource.GetName() == "terminus"
			case *corev1.ConfigMap:
				return resource.Namespace == "os-network" && resource.Name == "default-network"
			}
			return false
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: enqueue,
			UpdateFunc: func(oldObj, newObj interface{}) {
				if machineDNSFields(oldObj) != machineDNSFields(newObj) {
					enqueue(newObj)
				}
			},
			DeleteFunc: enqueue,
		},
	}
}

func (s *CorefileNetworkSubscriber) Do(ctx context.Context, _ interface{}, _ Action) error {
	return RegenerateCorefile(ctx, s.kubeClient, s.dynamicClient)
}

func RegisterCorefileNetworkWatcher(w *Watchers, kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) error {
	sub := &CorefileNetworkSubscriber{Subscriber: NewSubscriber(w), kubeClient: kubeClient, dynamicClient: dynamicClient}
	if err := AddToWatchers[unstructured.Unstructured](w, terminusGVR, sub.HandleEvent()); err != nil {
		return err
	}
	return AddToWatchers[corev1.ConfigMap](w, corev1.SchemeGroupVersion.WithResource("configmaps"), sub.HandleEvent())
}
