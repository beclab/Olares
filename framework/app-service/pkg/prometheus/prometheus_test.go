package prometheus

import (
	"context"
	"testing"
	"time"

	"k8s.io/klog/v2"
)

func TestGetClusterMetrics(t *testing.T) {
	p, err := New("http://54.241.136.45:31712")
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	ms := p.GetNamedMetrics(context.TODO(), []string{"namespaces_cpu_usage", "namespaces_memory_usage"}, time.Now(), QueryOptions{Level: LevelCluster})

	for _, m := range ms {
		r := GetSortedNamespaceMetrics(&m)

		klog.Info(r)
	}
}
