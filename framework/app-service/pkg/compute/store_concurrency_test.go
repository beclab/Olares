package compute

import (
	"context"
	"encoding/json"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

// TestMutateAllocationsRetriesLostFirstCreate is a regression test for the
// first-allocation race: when the config map does not exist yet, two
// concurrent allocations both observe NotFound and both Create. The loser used
// to receive AlreadyExists, which RetryOnConflict did not retry, so its
// allocation was silently dropped. mutateAllocations now retries AlreadyExists,
// re-reads the winner's config map, and merges via the Update path.
//
// The race is reproduced deterministically: a Create interceptor persists a
// "winner" config map out-of-band on the first Create and returns AlreadyExists
// to the caller, exactly as a concurrent writer would.
func TestMutateAllocationsRetriesLostFirstCreate(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add corev1 to scheme: %v", err)
	}

	winner := Allocation{AppName: "winner", Owner: "alice", NodeName: "nvidia-a", DeviceID: "gpu0"}
	winnerData, err := json.Marshal([]Allocation{winner})
	if err != nil {
		t.Fatalf("marshal winner allocations: %v", err)
	}

	createCalls := 0
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				createCalls++
				if createCalls == 1 {
					// A concurrent writer wins the create with its own data.
					winnerCM := &corev1.ConfigMap{}
					winnerCM.Namespace = allocationConfigMapNamespace
					winnerCM.Name = allocationConfigMapName
					winnerCM.Data = map[string]string{allocationConfigMapKey: string(winnerData)}
					if err := cl.Create(ctx, winnerCM); err != nil {
						return err
					}
					return apierrors.NewAlreadyExists(
						schema.GroupResource{Resource: "configmaps"},
						allocationConfigMapName,
					)
				}
				return cl.Create(ctx, obj, opts...)
			},
		}).
		Build()

	loser := Allocation{AppName: "loser", Owner: "bob", NodeName: "nvidia-a", DeviceID: "gpu0"}
	selected, err := mutateAllocations(context.Background(), c, func(_ []Node, allocations []Allocation) ([]Allocation, *Allocation, error) {
		next := append(append([]Allocation{}, allocations...), loser)
		return next, &loser, nil
	})
	if err != nil {
		t.Fatalf("mutateAllocations must retry the lost create instead of failing: %v", err)
	}
	if selected == nil || selected.AppName != loser.AppName {
		t.Fatalf("expected the loser allocation to be selected, got %#v", selected)
	}

	var cm corev1.ConfigMap
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: allocationConfigMapNamespace,
		Name:      allocationConfigMapName,
	}, &cm); err != nil {
		t.Fatalf("get allocation config map: %v", err)
	}
	var persisted []Allocation
	if err := json.Unmarshal([]byte(cm.Data[allocationConfigMapKey]), &persisted); err != nil {
		t.Fatalf("unmarshal persisted allocations: %v", err)
	}

	var haveWinner, haveLoser bool
	for _, allocation := range persisted {
		switch allocation.AppName {
		case winner.AppName:
			haveWinner = true
		case loser.AppName:
			haveLoser = true
		}
	}
	if !haveWinner || !haveLoser {
		t.Fatalf("both the winner's and the loser's allocations must survive the create race, got %#v", persisted)
	}
}
