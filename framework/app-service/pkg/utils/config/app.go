package config

import (
	"github.com/Masterminds/semver/v3"
	"k8s.io/klog/v2"
)

const (
	// AppCfgFileName config file name for application.
	AppCfgFileName = "OlaresManifest.yaml"

	MinCfgFileVersion  = ">= 0.7.2"
	NewManifestVersion = "0.12.0"
)

func IsNewManifestVersion(version string) bool {
	if version == "" {
		return false
	}
	c, err := semver.NewConstraint(">= " + NewManifestVersion)
	if err != nil {
		klog.Errorf("invalid new manifest version constraint: %v", err)
		return false
	}
	v, err := semver.NewVersion(version)
	if err != nil {
		klog.Errorf("invalid manifest version %s: %v", version, err)
		return false
	}
	return c.Check(v)
}
