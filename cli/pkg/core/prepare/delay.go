package prepare

import (
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"time"
)

// InitialDelay is a Prepare implementation that simply wait for Duration amount of time
type InitialDelay struct {
	BasePrepare
	Duration time.Duration
}

func (p *InitialDelay) PreCheck(runtime connector.Runtime) (bool, error) {
	time.Sleep(p.Duration)
	return true, nil
}
