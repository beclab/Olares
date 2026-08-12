package upgrade

import (
	"strings"
	"testing"
)

func TestUpgrader_1_12_7_20260803_UpdateOlaresVersionOrder(t *testing.T) {
	u := upgrader_1_12_7_20260803{}
	tasks := u.UpdateOlaresVersion()
	if len(tasks) < 2 {
		t.Fatalf("want recreate then version write, got %d", len(tasks))
	}
	if tasks[0].GetName() != "DeenvyEdgeBestEffortRecreate" {
		t.Fatalf("first task=%q", tasks[0].GetName())
	}
	foundVersion := false
	for _, tk := range tasks {
		name := tk.GetName()
		if strings.Contains(strings.ToLower(name), "accept") {
			t.Fatalf("must not gate on %q", name)
		}
		if name == "UpdateOlaresVersion" {
			foundVersion = true
		}
	}
	if !foundVersion {
		t.Fatal("expected UpdateOlaresVersion after recreate")
	}
}
