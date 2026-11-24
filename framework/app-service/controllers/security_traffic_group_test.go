package controllers

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway/meshinagent"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	netv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func TestAutheliaBFLPolicySurvivesSweepAndRevokesDeletedService(t *testing.T) {
	ctx := context.Background()
	newService := func(namespace, ip string) *corev1.Service {
		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "bfl", Namespace: namespace},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP, ClusterIP: ip, Selector: map[string]string{"tier": "bfl"},
				Ports: []corev1.ServicePort{{Port: 80, TargetPort: intstr.FromInt(8080), Protocol: corev1.ProtocolTCP}},
			},
		}
	}
	alice := newService("user-space-alice", "10.233.1.11")
	bob := newService("user-space-bob", "10.233.1.12")
	foreign := newService("user-system-alice", "10.233.1.13")
	wrongSelector := newService("user-space-other", "10.233.1.14")
	wrongSelector.Spec.Selector["tier"] = "other"
	r := &SecurityReconciler{Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).
		WithObjects(alice, bob, foreign, wrongSelector).Build()}
	namespaces := []string{"os-framework", alice.Namespace, bob.Namespace}
	wantRequests := []reconcile.Request{
		{NamespacedName: client.ObjectKey{Name: alice.Namespace}},
		{NamespacedName: client.ObjectKey{Name: "user-system-alice"}},
		{NamespacedName: client.ObjectKey{Name: "os-framework"}},
	}
	if got := trafficGroupNetworkPolicyRequests(ctx, alice); !reflect.DeepEqual(got, wantRequests) {
		t.Fatalf("BFL event enqueues %v, want %v", got, wantRequests)
	}
	reconcileDependencies := func() {
		t.Helper()
		for _, namespace := range namespaces {
			if err := r.reconcileTrafficGroupDependencies(ctx, namespace); err != nil {
				t.Fatal(err)
			}
		}
	}
	getPolicy := func(namespace, direction string) netv1.NetworkPolicy {
		t.Helper()
		var policy netv1.NetworkPolicy
		key := client.ObjectKey{Namespace: namespace, Name: "olares-platform-authelia-bfl-" + direction}
		if err := r.Get(ctx, key, &policy); err != nil {
			t.Fatal(err)
		}
		return policy
	}
	reconcileDependencies()
	out := getPolicy("os-framework", "egress")
	aliceIn := getPolicy(alice.Namespace, "ingress")
	bobIn := getPolicy(bob.Namespace, "ingress")
	if len(out.Spec.Egress) != 4 || len(aliceIn.Spec.Ingress) != 1 || len(bobIn.Spec.Ingress) != 1 {
		t.Fatalf("unexpected BFL grants: egress=%v alice=%v bob=%v", out.Spec, aliceIn.Spec, bobIn.Spec)
	}
	autheliaSelector := map[string]string{"app": "authelia-backend", "traffic.olares.io/platform-mesh": "true"}
	if !reflect.DeepEqual(out.Spec.PodSelector.MatchLabels, autheliaSelector) {
		t.Fatalf("unexpected BFL caller: %v", out.Spec.PodSelector)
	}
	for i, in := range []netv1.NetworkPolicy{aliceIn, bobIn} {
		rule := in.Spec.Ingress[0]
		to := out.Spec.Egress[2*i]
		if len(rule.From) != 1 || len(rule.Ports) != 1 || rule.Ports[0].Port.IntValue() != 8080 ||
			!reflect.DeepEqual(rule.From[0].NamespaceSelector.MatchLabels, map[string]string{"kubernetes.io/metadata.name": "os-framework"}) ||
			!reflect.DeepEqual(rule.From[0].PodSelector.MatchLabels, autheliaSelector) ||
			len(to.To) != 1 || len(to.Ports) != 1 || to.Ports[0].Port.IntValue() != 8080 ||
			to.To[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != in.Namespace ||
			!reflect.DeepEqual(to.To[0].PodSelector.MatchLabels, map[string]string{"tier": "bfl"}) {
			t.Fatalf("unexpected BFL peer in %s: ingress=%v egress=%v", in.Namespace, rule, to)
		}
	}
	for _, policy := range []netv1.NetworkPolicy{out, aliceIn, bobIn} {
		legacy := &netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "legacy", Namespace: policy.Namespace}}
		if err := r.createOrUpdateNetworkPolicy(ctx, legacy, nil, &security.NetworkPolicies{legacy}); err != nil {
			t.Fatal(err)
		}
		var current netv1.NetworkPolicy
		if err := r.Get(ctx, client.ObjectKeyFromObject(&policy), &current); err != nil || current.ResourceVersion != policy.ResourceVersion {
			t.Fatalf("legacy sweep changed policy %s/%s: %v", policy.Namespace, policy.Name, err)
		}
	}
	if err := r.Delete(ctx, alice); err != nil {
		t.Fatal(err)
	}
	reconcileDependencies()
	if got := getPolicy("os-framework", "egress"); !reflect.DeepEqual(got.Spec.Egress, out.Spec.Egress[2:]) {
		t.Fatalf("unexpected egress after deleting Alice's BFL: %v", got.Spec.Egress)
	}
	if got := getPolicy(alice.Namespace, "ingress"); len(got.Spec.Ingress) != 0 {
		t.Fatalf("Alice still has an ingress grant: %v", got.Spec.Ingress)
	}
	if got := getPolicy(bob.Namespace, "ingress"); got.ResourceVersion != bobIn.ResourceVersion {
		t.Fatalf("Bob's policy changed after deleting Alice's BFL: %v", got.Spec)
	}
}

func TestTrafficGroupPoliciesSurviveLegacySweep(t *testing.T) {
	ctx := context.WithValue(context.Background(), loggerKey, logr.Discard())
	for _, ns := range []string{"kube-system", "os-framework", "user-system-alice", "user-space-alice"} {
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
			expected := security.BuildTrafficGroupIsolationBundle(ns)
			expected = append(expected, security.BuildTrafficGroupDependencyPolicies(ns, nil, nil)...)
			expected = append(expected, security.BuildPlatformTrafficPolicies(ns, nil)...)
			for _, policy := range expected {
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
	if err := r.reconcileTrafficGroupNetworkPolicies(ctx, namespace); !errors.Is(err, errTrafficGroupPending) {
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

func TestTrafficGroupAppliesNoEgressBeforeReadinessAndKeepsExistingIsolation(t *testing.T) {
	ctx := context.WithValue(t.Context(), loggerKey, logr.Discard())
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "os-framework", UID: "namespace-uid"}}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: namespace.Name, Name: "native", UID: "pod-uid"},
		Status: corev1.PodStatus{Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}}}
	objs := []client.Object{namespace, pod}
	for _, key := range []client.ObjectKey{kubernetesAPIKey, {Namespace: "kube-system", Name: "coredns"}} {
		objs = append(objs, &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name, UID: "service"}},
			&discoveryv1.EndpointSlice{ObjectMeta: metav1.ObjectMeta{
				Namespace: key.Namespace, Name: key.Name, Labels: map[string]string{discoveryv1.LabelServiceName: key.Name},
				OwnerReferences: []metav1.OwnerReference{{Kind: "Service", Name: key.Name, UID: "service", Controller: ptr.To(true)}},
			}, Ports: []discoveryv1.EndpointPort{{Name: ptr.To("https"), Port: ptr.To(int32(443))}},
				Endpoints: []discoveryv1.Endpoint{{Addresses: []string{"10.0.0.1"}, Conditions: discoveryv1.EndpointConditions{Ready: ptr.To(true)}}}})
	}
	objs = append(objs, &corev1.Endpoints{ObjectMeta: metav1.ObjectMeta{
		Namespace: kubernetesAPIKey.Namespace, Name: kubernetesAPIKey.Name, UID: "api-endpoints",
	}, Subsets: []corev1.EndpointSubset{{
		Addresses: []corev1.EndpointAddress{{IP: "10.0.0.1"}},
		Ports:     []corev1.EndpointPort{{Name: "https", Protocol: corev1.ProtocolTCP, Port: 443}},
	}}})
	var writes []string
	r := &SecurityReconciler{KubeClient: kubefake.NewSimpleClientset(),
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objs...).
			WithInterceptorFuncs(interceptor.Funcs{Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				if _, isNetworkPolicy := obj.(*netv1.NetworkPolicy); isNetworkPolicy {
					writes = append(writes, obj.GetName())
				}
				return c.Create(ctx, obj, opts...)
			}}).Build()}
	meshinagent.SetMeshControlPlaneReadyCheck(func(context.Context, kubernetes.Interface) bool { return true })
	t.Cleanup(func() { meshinagent.SetMeshControlPlaneReadyCheck(nil) })

	if err := r.reconcileTrafficGroupNetworkPolicies(ctx, namespace); !errors.Is(err, errTrafficGroupPending) || len(writes) != 0 {
		t.Fatalf("control plane readiness alone must not write any egress policies: %v, writes %v", err, writes)
	}
	record := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: mesh.EgressReadinessConfig.Namespace, Name: mesh.EgressReadinessConfig.Name},
		Data: map[string]string{namespace.Name: `{"namespaceUID":"namespace-uid","pods":{"native":"pod-uid"},"nativePods":["native"]}`}}
	if err := r.Create(ctx, record); err != nil {
		t.Fatal(err)
	}
	if err := r.reconcileTrafficGroupNetworkPolicies(ctx, namespace); err != nil {
		t.Fatal(err)
	}
	baseline := security.NPGroupDefaultDeny(namespace.Name)
	if len(writes) <= 1 || writes[len(writes)-1] != baseline.Name {
		t.Fatalf("the default deny must follow dependency and ingress policies: %v", writes)
	}
	if err := r.Delete(ctx, record); err != nil {
		t.Fatal(err)
	}
	meshinagent.SetMeshControlPlaneReadyCheck(nil)
	if err := r.reconcileTrafficGroupNetworkPolicies(ctx, namespace); err != nil {
		t.Fatal(err)
	}
	var live netv1.NetworkPolicy
	if err := r.Get(ctx, client.ObjectKeyFromObject(&baseline), &live); err != nil ||
		!reflect.DeepEqual(live.Spec.PolicyTypes, baseline.Spec.PolicyTypes) {
		t.Fatalf("existing egress isolation was removed during an outage: %v, %v", err, live.Spec.PolicyTypes)
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
