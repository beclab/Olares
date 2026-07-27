package market

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func TestMarketCloneHelpMatchesCloneContract(t *testing.T) {
	long := NewCmdMarketClone(&cmdutil.Factory{}).Long
	for _, unwanted := range []string{"Clone an installed application", "cloneTarget", "'cloneable: true'"} {
		if strings.Contains(long, unwanted) {
			t.Errorf("clone help contains stale contract %q", unwanted)
		}
	}
	for _, required := range []string{"catalog", "allowMultipleInstall", "templateOnly", "targetApp"} {
		if !strings.Contains(long, required) {
			t.Errorf("clone help must describe %q", required)
		}
	}
}

func TestMarketGetHelpDescribesComputedCloneability(t *testing.T) {
	long := NewCmdMarketGet(&cmdutil.Factory{}).Long
	for _, unwanted := range []string{"full upstream payload", "jq '.cloneable'"} {
		if strings.Contains(long, unwanted) {
			t.Errorf("get help contains stale cloneability guidance %q", unwanted)
		}
	}
	for _, required := range []string{"Cloneable", "allowMultipleInstall", "templateOnly"} {
		if !strings.Contains(long, required) {
			t.Errorf("get help must describe %q", required)
		}
	}
}
