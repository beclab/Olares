package controllers

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func TestTrafficGroupPoliciesSurviveLegacySweep(t *testing.T) {
	ctx := context.WithValue(context.Background(), loggerKey, logr.Discard())
	for _, ns := range []string{"kube-system", "os-framework"} {
		t.Run(ns, func(t *testing.T) {
			namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
			baseline := security.NPGroupDefaultDeny(ns)
			legacy := &netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "legacy", Namespace: ns}}
			obsolete := baseline.DeepCopy()
			obsolete.Name = "obsolete"
			r := &SecurityReconciler{Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).
				WithObjects(namespace, &baseline, obsolete).Build()}
			legacyPolicies := security.NetworkPolicies{legacy}

			reconcile := func() {
				t.Helper()
				if err := r.createOrUpdateNetworkPolicy(ctx, legacy, nil, &legacyPolicies); err != nil {
					t.Fatal(err)
				}
				if err := r.reconcileTrafficGroupNetworkPolicies(ctx, namespace); err != nil {
					t.Fatal(err)
				}
			}
			reconcile()
			var policies netv1.NetworkPolicyList
			if err := r.List(ctx, &policies, client.InNamespace(ns)); err != nil {
				t.Fatal(err)
			}
			versions := make(map[string]string)
			for _, policy := range policies.Items {
				versions[policy.Name] = policy.ResourceVersion
			}
			for _, policy := range append(security.BuildTrafficGroupIsolationBundle(ns),
				security.BuildTrafficGroupDependencyPolicies(ns, nil, nil)...) {
				if _, ok := versions[policy.Name]; !ok {
					t.Fatalf("missing policy %s", policy.Name)
				}
			}
			reconcile()
			for name, version := range versions {
				var policy netv1.NetworkPolicy
				if err := r.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, &policy); err != nil {
					t.Fatal(err)
				}
				if policy.ResourceVersion != version {
					t.Fatalf("unchanged policy %s rewritten or recreated", name)
				}
			}
			if err := r.Get(ctx, client.ObjectKeyFromObject(obsolete), &netv1.NetworkPolicy{}); !apierrors.IsNotFound(err) {
				t.Fatalf("obsolete policy not pruned: %v", err)
			}
		})
	}
}

func TestTrafficGroupAPIDestinationsRotateWhileMeshIsUnready(t *testing.T) {
	ctx := context.WithValue(context.Background(), loggerKey, logr.Discard())
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "os-framework"}}
	baseline := security.NPGroupDefaultDeny(namespace.Name)
	meta := metav1.ObjectMeta{Namespace: "default", Name: "kubernetes"}
	service := &corev1.Service{ObjectMeta: meta, Spec: corev1.ServiceSpec{
		Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.233.0.1",
		Ports: []corev1.ServicePort{{Name: "https", Protocol: corev1.ProtocolTCP, Port: 443}},
	}}
	endpoints := &corev1.Endpoints{ObjectMeta: meta, Subsets: []corev1.EndpointSubset{{
		Addresses: []corev1.EndpointAddress{{IP: "192.168.128.102"}},
		Ports:     []corev1.EndpointPort{{Name: "https", Protocol: corev1.ProtocolTCP, Port: 6443}},
	}}}
	readError := errors.New("API source unavailable")
	var failSource client.Object
	r := &SecurityReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).
			WithObjects(namespace, service, endpoints).
			WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if key == kubernetesAPIKey && reflect.TypeOf(obj) == reflect.TypeOf(failSource) {
						return readError
					}
					return c.Get(ctx, key, obj, opts...)
				},
			}).Build(),
		KubeClient: kubefake.NewSimpleClientset(),
	}
	key := client.ObjectKey{Namespace: namespace.Name, Name: "olares-traffic-group-kube-api-egress"}
	if err := r.reconcileTrafficGroupNetworkPolicies(ctx, namespace); err != nil {
		t.Fatal(err)
	}
	if err := r.Get(ctx, key, &netv1.NetworkPolicy{}); !apierrors.IsNotFound(err) {
		t.Fatalf("new isolation policy created while mesh is unready: %v", err)
	}
	foreign := baseline.DeepCopy()
	foreign.Labels = nil
	if err := r.Create(ctx, foreign); err != nil {
		t.Fatal(err)
	}
	if err := r.reconcileTrafficGroupNetworkPolicies(ctx, namespace); err != nil {
		t.Fatal(err)
	}
	if err := r.Get(ctx, key, &netv1.NetworkPolicy{}); !apierrors.IsNotFound(err) {
		t.Fatalf("unowned baseline activated dependency policies: %v", err)
	}
	if err := r.Delete(ctx, foreign); err != nil {
		t.Fatal(err)
	}
	if err := r.Create(ctx, &baseline); err != nil {
		t.Fatal(err)
	}
	check := func(want ...string) {
		t.Helper()
		if err := r.reconcileTrafficGroupNetworkPolicies(ctx, namespace); err != nil {
			t.Fatal(err)
		}
		var policy netv1.NetworkPolicy
		if err := r.Get(ctx, key, &policy); err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, rule := range policy.Spec.Egress {
			got = append(got, rule.To[0].IPBlock.CIDR)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("API destinations = %v, want %v", got, want)
		}
	}
	check("10.233.0.1/32", "192.168.128.102/32")
	for _, obj := range []client.Object{service, endpoints} {
		var before, after netv1.NetworkPolicy
		if err := r.Get(ctx, key, &before); err != nil {
			t.Fatal(err)
		}
		failSource = obj
		if err := r.reconcileTrafficGroupNetworkPolicies(ctx, namespace); !errors.Is(err, readError) {
			t.Fatalf("failed source read was not returned: %v", err)
		}
		failSource = nil
		if err := r.Get(ctx, key, &after); err != nil || after.ResourceVersion != before.ResourceVersion {
			t.Fatalf("failed source read rewrote the policy: %v", err)
		}
	}
	if err := r.Get(ctx, kubernetesAPIKey, endpoints); err != nil {
		t.Fatal(err)
	}
	endpoints.Subsets[0].Addresses[0].IP = "192.168.128.103"
	if err := r.Update(ctx, endpoints); err != nil {
		t.Fatal(err)
	}
	check("10.233.0.1/32", "192.168.128.103/32")
	if err := r.Delete(ctx, endpoints); err != nil {
		t.Fatal(err)
	}
	check("10.233.0.1/32")
	if err := r.Delete(ctx, service); err != nil {
		t.Fatal(err)
	}
	check()
	for _, obj := range []client.Object{service, endpoints} {
		requests := trafficGroupNetworkPolicyRequests(ctx, obj)
		if len(requests) != 5 || requests[0].Name != "kube-system" || requests[1].Name != "os-framework" || requests[2].Name != "os-gateway" || requests[3].Name != "os-mesh" || requests[4].Name != "os-platform" {
			t.Fatalf("API source event did not enqueue both policy namespaces: %+v", requests)
		}
	}
	nats := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "os-platform", Name: "nats"}}
	if got := trafficGroupNetworkPolicyRequests(ctx, nats); len(got) != 2 ||
		got[0].Name != "os-framework" || got[1].Name != "os-platform" {
		t.Fatalf("NATS change did not enqueue both peers: %+v", got)
	}
	endpoints.Namespace = "user-space-test"
	if got := trafficGroupNetworkPolicyRequests(ctx, endpoints); len(got) != 0 {
		t.Fatalf("unrelated API source enqueued reconciles: %+v", got)
	}
}

type trafficGroupWatchRecorder struct {
	controller.Controller
	watches int
	err     error
}

func (c *trafficGroupWatchRecorder) Watch(source.Source) error {
	c.watches++
	return c.err
}

func TestTrafficGroupWatchesRegisterBothAPISources(t *testing.T) {
	c := &trafficGroupWatchRecorder{}
	if err := watchTrafficGroupNetworkPolicySources(c, nil); err != nil || c.watches != 3 {
		t.Fatalf("registered %d watches, want 3: %v", c.watches, err)
	}
	c.err = errors.New("watch registration failed")
	if err := watchTrafficGroupNetworkPolicySources(c, nil); !errors.Is(err, c.err) || c.watches != 4 {
		t.Fatalf("watch registration continued after failure: %v", err)
	}
}
