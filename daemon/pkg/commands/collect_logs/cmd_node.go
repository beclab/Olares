package collectlogs

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/fanout"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
)

// nodeCollectLogs is the node-local executor invoked by the master
// orchestrator on every node. It runs synchronously (the orchestrator waits on
// the HTTP response) and collects only this node's logs into the shared staging
// directory under the caller's Home.
type nodeCollectLogs struct {
	commands.Operation
	*commands.BaseCommand
}

var _ commands.Interface = &nodeCollectLogs{}

// NewNode builds the node-local executor. It reuses the CollectLogs operation
// name so the state-machine validators admit it without changes.
func NewNode() commands.Interface {
	return &nodeCollectLogs{
		Operation: commands.Operation{
			Name: commands.CollectLogs,
		},
		BaseCommand: commands.NewBaseCommand(),
	}
}

// NodeResultData is the payload a node returns to the orchestrator.
type NodeResultData struct {
	Node       string `json:"node"`
	StagingDir string `json:"stagingDir"`
}

func (i *nodeCollectLogs) Execute(ctx context.Context, p any) (res any, err error) {
	req, ok := p.(*NodeRequest)
	if !ok || req == nil {
		return nil, fmt.Errorf("collect-logs-node: invalid param type %T", p)
	}
	if req.RunID == "" {
		return nil, fmt.Errorf("collect-logs-node: missing runID")
	}

	kubeClient, err := utils.GetKubeClient()
	if err != nil {
		return nil, fmt.Errorf("create k8s client error: %w", err)
	}

	// Defense in depth: re-authorize under the forwarded identity even though
	// the master already did.
	rs, err := authorize(ctx, kubeClient, &req.Param)
	if err != nil {
		return nil, err
	}

	home, err := utils.GetUserspacePvcHostPath(ctx, rs.username, kubeClient)
	if err != nil {
		return nil, fmt.Errorf("get caller home path error: %w", err)
	}
	nodeName, _, _, err := utils.GetThisNodeName(ctx, kubeClient)
	if err != nil {
		return nil, fmt.Errorf("get this node name error: %w", err)
	}

	stagingDir := nodeStagingDir(home, req.RunID, nodeName)
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir staging dir error: %w", err)
	}

	cmds := translateFlags(rs, stagingDir)
	// Detached from the request context (a dropped dispatch connection must not
	// abort collection), but bounded so the node self-terminates slightly before
	// the master's per-node timeout. This prevents orphaned root collectors and
	// the master's staging cleanup from racing a still-writing process.
	runCtx, cancel := context.WithTimeout(context.Background(), fanout.DefaultTimeout-time.Minute)
	defer cancel()
	if _, err := i.BaseCommand.Run_(runCtx, "olares-cli", cmds...); err != nil {
		return nil, fmt.Errorf("collect logs error: %w", err)
	}

	return &NodeResultData{Node: nodeName, StagingDir: stagingDir}, nil
}

// translateFlags turns the resolved scope into olares-cli logs flags. Wildcards
// are already expanded, so the CLI only ever sees concrete values.
func translateFlags(rs *resolvedScope, outputDir string) []string {
	cmds := []string{
		"logs",
		"--output-dir", outputDir,
		"--ignore-kube-errors", "true",
	}

	if rs.collectSystemd {
		if len(rs.systemdComponents) > 0 {
			cmds = append(cmds, "--components", strings.Join(rs.systemdComponents, ","))
		}
		if rs.since != "" {
			cmds = append(cmds, "--since", rs.since)
		}
		if rs.maxLines > 0 {
			cmds = append(cmds, "--max-lines", strconv.Itoa(rs.maxLines))
		}
	} else {
		cmds = append(cmds, "--skip-systemd")
	}

	if !rs.collectDmesg {
		cmds = append(cmds, "--skip-dmesg")
	}
	if !rs.collectNetwork {
		cmds = append(cmds, "--skip-network")
	}
	if !rs.collectClusterInfo {
		cmds = append(cmds, "--skip-cluster-info")
	}

	if !rs.collectPods {
		cmds = append(cmds, "--skip-pod-logs")
	} else if !rs.allNamespaces {
		cmds = append(cmds, "--pod-namespaces", strings.Join(rs.podNamespaces, ","))
	}

	return cmds
}
