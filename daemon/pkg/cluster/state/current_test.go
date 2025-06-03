package state

import (
	"context"
	"testing"

	"bytetrade.io/web3os/terminusd/pkg/utils"
)

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
