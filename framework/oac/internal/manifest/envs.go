package manifest

import (
	"errors"
	"fmt"
	"strings"
)

const olaresUserEnvPrefix = "OLARES_USER"

// validateV3Envs applies apiVersion=v3 rules to envs[]. App-local envName
// values must not use the OLARES_USER prefix; that namespace is reserved for
// user-level variables referenced via valueFrom.envName.
func validateV3Envs(envs []AppEnvVar) error {
	var errs []error
	for i, e := range envs {
		if strings.HasPrefix(e.EnvName, olaresUserEnvPrefix) {
			errs = append(errs, fmt.Errorf(
				"envs[%d].envName: must not start with %q (declare an app-local name and map user variables with valueFrom.envName)",
				i, olaresUserEnvPrefix,
			))
		}
	}
	return errors.Join(errs...)
}

// validateV3Configuration runs checks that only apply to apiVersion=v3 manifests.
func validateV3Configuration(c *AppConfiguration) error {
	if normalizeAPIVersion(c.APIVersion) != APIVersionV3 {
		return nil
	}
	return validateV3Envs(c.Envs)
}
