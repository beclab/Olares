package phase

import (
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"bytetrade.io/web3os/installer/pkg/terminus"
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
