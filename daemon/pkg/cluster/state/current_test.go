package state

import (
	"context"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/utils"
)

func TestPrimaryGPUUsesFirstListEntry(t *testing.T) {
	got := primaryGPU([]string{"NVIDIA RTX 4070", "Intel Arc"})
	if got == nil || *got != "NVIDIA RTX 4070" {
		t.Fatalf("primaryGPU() = %#v, want first GPU", got)
	}
	if got := primaryGPU(nil); got != nil {
		t.Fatalf("primaryGPU(nil) = %#v, want nil", got)
	}
}

func TestCurrentState(t *testing.T) {
	err := CheckCurrentStatus(context.Background())
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log("state: ", CurrentState)
}

func TestFindProcess(t *testing.T) {
	p, err := utils.ProcessExists(2687)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Logf("process: %v", p)
}
