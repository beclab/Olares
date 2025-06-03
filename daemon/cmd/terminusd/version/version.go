package version

import (
	"fmt"

	semver "github.com/Masterminds/semver/v3"
	"k8s.io/klog/v2"
)

var version = "debug"

func Version() string {
	return fmt.Sprintf("olaresd version: %s", version)
}

func RawVersion() *string {
	return &version
}

func VersionUpgrade(newVersion string) bool {
	if version == "debug" || newVersion == "" {
		return false
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		klog.Error("wrong version format, ", err, ", ", version)
		return false
	}

	nv, err := semver.NewVersion(newVersion)
	if err != nil {
		klog.Error("wrong new version format, ", err, ", ", newVersion)
		return false
	}

	return nv.GreaterThan(v)
}
