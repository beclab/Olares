package node

import (
	"bytes"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
)

func TestRenderListTableIncludesArchitecture(t *testing.T) {
	items := []Node{{}, {}, {}}
	items[0].Metadata.Name = "amd-node"
	items[0].Status.NodeInfo.Architecture = "amd64"
	items[1].Metadata.Name = "arm-node"
	items[1].Status.NodeInfo.Architecture = "arm64"
	items[2].Metadata.Name = "unknown-node"

	var out bytes.Buffer
	p := clusteropts.NewPaginationOptions()
	if err := renderListTable(&out, items, false, p, len(items)); err != nil {
		t.Fatalf("render node list: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"ARCHITECTURE",
		"amd-node",
		"amd64",
		"arm-node",
		"arm64",
		"unknown-node",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("node list missing %q:\n%s", want, got)
		}
	}

	for _, line := range strings.Split(got, "\n") {
		if strings.HasPrefix(line, "unknown-node ") && !strings.HasSuffix(strings.TrimSpace(line), "-") {
			t.Errorf("empty architecture should render as '-':\n%s", got)
		}
	}
}
