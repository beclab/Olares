package webhook

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

func TestAdmissionMatchConditionsSupported(t *testing.T) {
	cases := map[string]struct {
		supported bool
		wantErr   bool
	}{
		"v1.33.3+k3s1": {supported: true},
		"v1.28.0":      {supported: true},
		"v1.27.9":      {supported: false},
		"1.26.15":      {supported: false},
		"garbage":      {wantErr: true},
	}
	for gitVersion, want := range cases {
		got, err := admissionMatchConditionsSupported(gitVersion)
		if (err != nil) != want.wantErr {
			t.Errorf("%s: err = %v, wantErr %v", gitVersion, err, want.wantErr)
			continue
		}
		if got != want.supported {
			t.Errorf("%s: supported = %v, want %v", gitVersion, got, want.supported)
		}
	}
}

func TestMacvlanMatchExpressionsCoverLabelAndBothSelectionKeys(t *testing.T) {
	expr := macvlanInitMatchExpression()
	for _, want := range []string{constants.ApplicationMacvlanInitLabel, multusNetworksAnnotation, multusDefaultNetworkAnnotation, "||"} {
		if !strings.Contains(expr, want) {
			t.Fatalf("mutating match expression %q lacks %q", expr, want)
		}
	}
	for _, want := range []string{multusNetworksAnnotation, multusDefaultNetworkAnnotation} {
		if !strings.Contains(macvlanSelectionMatchExpression, want) {
			t.Fatalf("validating match expression %q lacks %q", macvlanSelectionMatchExpression, want)
		}
	}
}
