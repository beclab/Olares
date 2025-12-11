package metrics

import (
	"context"
	"testing"
	"time"
)

func TestGetClusterMetrics(t *testing.T) {
	p, err := NewPrometheus("http://54.241.136.45:31712")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	opts := QueryOptions{
		Level: LevelCluster,
	}

	ms := p.GetNamedMetrics(context.TODO(), []string{"cluster_cpu_usage", "cluster_cpu_total", "cluster_cpu_utilisation", "cluster_memory_utilisation"}, time.Now(), opts)

	for _, m := range ms {
		t.Logf("%s: %f", m.MetricName, GetValue(&m))
	}
}
