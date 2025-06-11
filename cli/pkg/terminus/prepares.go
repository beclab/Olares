package terminus

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
)

type NotEqualDesiredVersion struct {
	common.KubePrepare
}

func (n *NotEqualDesiredVersion) PreCheck(runtime connector.Runtime) (bool, error) {
	ksVersion, ok := n.PipelineCache.GetMustString(common.KubeSphereVersion)
	if !ok {
		ksVersion = ""
	}

	if n.KubeConf.Cluster.KubeSphere.Version == ksVersion {
		return false, nil
	}
	return true, nil
}
