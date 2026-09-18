package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// writeChart lays down a one-Deployment chart whose pod template is filled in
// from tmpl, so a test can drive the real helm render that install performs.
func writeChart(t *testing.T, podTemplate string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	files := map[string]string{
		"Chart.yaml": "apiVersion: v2\nname: gpumem\nversion: 0.1.0\n",
		"templates/deploy.yaml": `apiVersion: apps/v1
kind: Deployment
metadata:
  name: engine
spec:
  replicas: 1
  selector:
    matchLabels:
      app: engine
  template:
` + podTemplate,
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

// TestGetWorkloadResourcesFromChartGPUMemory drives the whole install-time path
// — helm render, workload walk, GPU-memory read — the way resolveAutoResources
// does, rather than the podGPUMemoryBytes helper on its own. It is the check
// that a chart written the way a discrete AMD/Intel app has to write it comes
// back with a usable quota instead of the zero that left its card unbindable.
func TestGetWorkloadResourcesFromChartGPUMemory(t *testing.T) {
	const nvidiaStylePod = `    metadata:
      labels:
        app: engine
    spec:
      containers:
      - name: engine
        image: busybox
        resources:
          requests:
            cpu: 200m
            memory: 2Gi
            nvidia.com/gpumem: 23552
          limits:
            cpu: "6"
            memory: 35Gi
            nvidia.com/gpumem: 23552
`

	// What amd-gpu / intel-gpu must render instead: no memory extended
	// resource anywhere, the quota on the pod template annotation.
	const annotatedPod = `    metadata:
      labels:
        app: engine
      annotations:
        gpu.bytetrade.io/required-gpu-memory: "23552Mi"
        gpu.bytetrade.io/limited-gpu-memory: "32Gi"
    spec:
      containers:
      - name: engine
        image: busybox
        resources:
          requests:
            cpu: 200m
            memory: 2Gi
          limits:
            cpu: "6"
            memory: 35Gi
`

	// The pre-fix shape for a discrete card: the mode has no way to declare a
	// quota, so nothing reaches the compute layer.
	const barePod = `    metadata:
      labels:
        app: engine
    spec:
      containers:
      - name: engine
        image: busybox
        resources:
          requests:
            cpu: 200m
            memory: 2Gi
          limits:
            cpu: "6"
            memory: 35Gi
`

	tests := []struct {
		name    string
		pod     string
		wantReq int64
		wantLim int64
	}{
		{"nvidia gpumem resource", nvidiaStylePod, 23 * gib, 23 * gib},
		{"discrete card annotation", annotatedPod, 23 * gib, 32 * gib},
		{"nothing declared", barePod, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totals, err := GetWorkloadResourcesFromChart(writeChart(t, tt.pod), map[string]interface{}{})
			if err != nil {
				t.Fatalf("render chart: %v", err)
			}
			if got := totals.RequestsGPUMemBytes.Value(); got != tt.wantReq {
				t.Errorf("requests gpu memory = %d bytes, want %d", got, tt.wantReq)
			}
			if got := totals.LimitsGPUMemBytes.Value(); got != tt.wantLim {
				t.Errorf("limits gpu memory = %d bytes, want %d", got, tt.wantLim)
			}
			// The ordinary quantities must survive the same walk untouched.
			if got := totals.RequestsMemory.String(); got != "2Gi" {
				t.Errorf("requests memory = %s, want 2Gi", got)
			}
			if got := totals.LimitsCPU.String(); got != "6" {
				t.Errorf("limits cpu = %s, want 6", got)
			}
		})
	}
}
