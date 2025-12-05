package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestWorkQueue(t *testing.T) {
	q := NewQ()

	a1 := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "app1",
			Namespace: "user1",
		},
	}
	a2 := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "app2",
			Namespace: "user1",
		},
	}
	a3 := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "app1",
			Namespace: "user2",
		},
	}
	q.Add(a1)
	q.Add(a2)
	q.Add(a3)
	// user1/app1 --> user2/app1 --> user1/app2
	result := []string{}
	item, _ := q.Get()
	a, _ := item.(reconcile.Request)
	result = append(result, a.String())
	item, _ = q.Get()
	a, _ = item.(reconcile.Request)
	result = append(result, a.String())
	q.Done(a1)
	item, _ = q.Get()
	a, _ = item.(reconcile.Request)
	result = append(result, a.String())
	expected := []string{"user1/app1", "user2/app1", "user1/app2"}
	assert.Equal(t, expected, result)
	q.Done(a2)
	q.Done(a3)

	// user1/app1 --> user1/app2 --> user2/app1
	q.Add(a1)
	q.Add(a2)
	q.Add(a3)
	expected = []string{"user1/app1", "user1/app2", "user2/app1"}
	result = []string{}
	item, _ = q.Get()
	q.Done(item)
	a, _ = item.(reconcile.Request)
	result = append(result, a.String())
	item, _ = q.Get()
	q.Done(item)
	a, _ = item.(reconcile.Request)
	result = append(result, a.String())
	item, _ = q.Get()
	q.Done(item)
	a, _ = item.(reconcile.Request)
	result = append(result, a.String())
	assert.Equal(t, expected, result)

}
