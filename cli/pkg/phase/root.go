package phase

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
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

func OlaresVersionGreaterThan(version string) (bool, error) {
	currentVersionString, err := GetOlaresVersion()
	if err != nil {
		return false, err
	}

	t, err := semver.StrictNewVersion(strings.TrimPrefix(version, "v"))
	if err != nil {
		return false, fmt.Errorf("invalid version '%s'", version)
	}
	c, err := semver.StrictNewVersion(strings.TrimPrefix(currentVersionString, "v"))
	if err != nil {
		return false, fmt.Errorf("invalid current version '%s'", currentVersionString)
	}
	return c.Compare(t) > 0, nil
}
