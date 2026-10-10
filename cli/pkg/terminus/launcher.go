package terminus

import (
	"fmt"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
)

type ClearBFLValues struct {
	common.KubeAction
}

func (c *ClearBFLValues) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("cat /dev/null > %s/wizard/config/launcher/values.yaml", runtime.GetInstallerDir()), false, false)

	return nil
}
