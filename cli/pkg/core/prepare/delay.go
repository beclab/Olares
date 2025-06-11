package prepare

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/core/connector"
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
