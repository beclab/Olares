package utils

import (
	"errors"
	"reflect"
	"testing"
)

func TestBFLLauncherAppliesPoliciesBeforeEnforcing(t *testing.T) {
	type phase struct{ prepare, enforce, reuse bool }
	for _, tc := range []struct {
		name            string
		prepared, reuse bool
		want            []phase
	}{
		{"new install", false, false, []phase{{true, false, false}, {true, true, true}}},
		{"legacy upgrade", false, true, []phase{{true, false, true}, {true, true, true}}},
		{"prepared or already strict", true, true, []phase{{true, true, true}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]interface{}{"image": "beclab/bfl:v0.4.56", "userspace_pvc": "existing"}
			var got []phase
			err := stageBFLLauncherMesh(values, tc.prepared, tc.reuse, func(reuse bool) error {
				got = append(got, phase{values["prepareAPIMesh"] == true, values["enforceAPIMesh"] == true, reuse})
				if values["image"] != "beclab/bfl:v0.4.56" || values["userspace_pvc"] != "existing" {
					t.Fatal("mesh rollout changed the image or storage")
				}
				return nil
			})
			if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("launcher rollout phases = %+v, err = %v, want %+v", got, err, tc.want)
			}
		})
	}
}

func TestBFLLauncherLeavesEnforcementOffWhenPreparationFails(t *testing.T) {
	bfl := map[string]interface{}{}
	errPrepare := errors.New("policy write failed")
	calls := 0
	err := stageBFLLauncherMesh(bfl, false, true, func(bool) error { calls++; return errPrepare })
	if !errors.Is(err, errPrepare) || calls != 1 || bfl["enforceAPIMesh"] != false {
		t.Fatalf("preparation failure did not prevent activation: calls=%d, err=%v", calls, err)
	}
}
