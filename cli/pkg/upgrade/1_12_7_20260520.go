package upgrade

import (
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_7_20260520 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260520) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260520")
}

func (u upgrader_1_12_7_20260520) UpgradeSystemComponents() []task.Interface {
	pre := []task.Interface{
		// Apply the GPUBinding CRD schema bump BEFORE the HAMi helm upgrade
		// runs in upgraderBase.UpgradeSystemComponents(); Helm 3 does not
		// update objects under chart `crds/` on upgrade, so the new spec
		// fields (namespace, owner) would otherwise stay unknown to the
		// apiserver and get pruned out of any binding we write.
		&task.LocalTask{
			Name:   "ApplyGPUBindingCRD",
			Action: new(applyGPUBindingCRD),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	}
	pre = append(pre, u.upgraderBase.UpgradeSystemComponents()...)
	return append(pre,
		&task.LocalTask{
			Name:   "MigrateLegacyGPUBindings",
			Action: new(migrateLegacyGPUBindings),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260520{})
}
