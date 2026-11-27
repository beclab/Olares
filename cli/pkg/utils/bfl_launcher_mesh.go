package utils

import "fmt"

func stageBFLLauncherMesh(bfl map[string]interface{}, enforced, reuseValues bool, apply func(reuse, waitReady bool) error, waitPolicies func() error) error {
	bfl["prepareAPIMesh"] = true
	bfl["enforceAPIMesh"] = enforced
	if err := apply(reuseValues, enforced); err != nil {
		return fmt.Errorf("prepare BFL API policies: %w", err)
	}
	if err := waitPolicies(); err != nil {
		return err
	}
	if enforced {
		return nil
	}
	bfl["enforceAPIMesh"] = true
	if err := apply(true, true); err != nil {
		return fmt.Errorf("activate BFL API policies: %w", err)
	}
	return nil
}
