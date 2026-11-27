package utils

import "fmt"

func stageBFLLauncherMesh(bfl map[string]interface{}, prepared, reuseValues bool, apply func(reuse bool) error) error {
	bfl["prepareAPIMesh"] = true
	if !prepared {
		bfl["enforceAPIMesh"] = false
		if err := apply(reuseValues); err != nil {
			return fmt.Errorf("prepare BFL API policies: %w", err)
		}
	}
	bfl["enforceAPIMesh"] = true
	if err := apply(true); err != nil {
		return fmt.Errorf("activate BFL API policies: %w", err)
	}
	return nil
}
