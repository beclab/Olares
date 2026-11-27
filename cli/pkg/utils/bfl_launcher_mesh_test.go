package utils

import (
	"errors"
	"reflect"
	"testing"
)

func TestBFLLauncherAppliesPoliciesBeforeEnforcing(t *testing.T) {
	type phase struct{ prepare, enforce, reuse, waitReady bool }
	for _, tc := range []struct {
		name            string
		enforced, reuse bool
		want            []phase
	}{
		{"new install", false, false, []phase{{true, false, false, false}, {true, true, true, true}}},
		{"legacy upgrade", false, true, []phase{{true, false, true, false}, {true, true, true, true}}},
		{"already strict", true, true, []phase{{true, true, true, true}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]interface{}{"image": "beclab/bfl:v0.4.56", "userspace_pvc": "existing"}
			var got []phase
			err := stageBFLLauncherMesh(values, tc.enforced, tc.reuse, func(reuse, waitReady bool) error {
				got = append(got, phase{values["prepareAPIMesh"] == true, values["enforceAPIMesh"] == true, reuse, waitReady})
				if values["image"] != "beclab/bfl:v0.4.56" || values["userspace_pvc"] != "existing" {
					t.Fatal("mesh rollout changed the image or storage")
				}
				return nil
			}, func() error { return nil })
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
	err := stageBFLLauncherMesh(bfl, false, true, func(bool, bool) error { calls++; return errPrepare }, func() error { t.Fatal("waited after failed preparation"); return nil })
	if !errors.Is(err, errPrepare) || calls != 1 || bfl["enforceAPIMesh"] != false {
		t.Fatalf("preparation failure did not prevent activation: calls=%d, err=%v", calls, err)
	}
}

func TestBFLLauncherDoesNotEnforceBeforeLinkerdAccepts(t *testing.T) {
	bfl := map[string]interface{}{}
	notAccepted := errors.New("Linkerd did not accept route")
	calls := 0
	err := stageBFLLauncherMesh(bfl, false, true, func(bool, bool) error {
		calls++
		return nil
	}, func() error { return notAccepted })
	if !errors.Is(err, notAccepted) || calls != 1 || bfl["enforceAPIMesh"] != false {
		t.Fatalf("unaccepted route did not prevent activation: calls=%d, err=%v", calls, err)
	}
}
