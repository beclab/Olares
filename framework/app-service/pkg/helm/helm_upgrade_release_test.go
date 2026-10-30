package helm

import (
	"context"
	"strings"
	"testing"
)

func TestUpgradeReleaseChartsRejectsNilChart(t *testing.T) {
	err := UpgradeReleaseCharts(context.Background(), nil, "demo", "demo-ns", nil, nil)
	if err == nil {
		t.Fatal("expected nil release chart error")
	}
	if !strings.Contains(err.Error(), "no chart") {
		t.Fatalf("error = %q, want missing chart detail", err)
	}
}
