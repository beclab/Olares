package phase

import (
	"github.com/beclab/Olares/cli/pkg/kubernetes"
	"github.com/beclab/Olares/cli/pkg/terminus"
)

func GetOlaresVersion() (string, error) {
	var terminusTask = &terminus.GetOlaresVersion{}
	return terminusTask.Execute()
}

func GetKubeType() string {
	var kubeTypeTask = &kubernetes.GetKubeType{}
	return kubeTypeTask.Execute()
}

func GetKubeVersion() (string, string, error) {
	var kubeTask = &kubernetes.GetKubeVersion{}
	return kubeTask.Execute()
}
