package pipelines

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/beclab/Olares/cli/version"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/validation"
)

// joinNodeLogFile collects the whole join in one file. Every pipeline the flow
// drives is pinned to it, so a failure anywhere can be diagnosed from a single
// log instead of five.
const joinNodeLogFile = "joinnode.log"

// JoinNodePipeline is the whole worker-side flow behind the single command that
// `olares-cli node join-command` generates on the master:
//
//  1. work out how to reach the master and prove the credentials work
//  2. check the master can accept a worker, and that this node may join it
//  3. give this node a node name the cluster does not already use
//  4. bring its dependencies to the cluster's exact Olares version
//  5. run the regular add-node pipeline
//
// Steps 1 to 3 are the only ones that can ask a question, and they all run
// within seconds, so the operator is never called back to the terminal once the
// long downloads have started. Everything that can be derived or was supplied
// up front is not asked about at all, which is what makes the flow scriptable.
func JoinNodePipeline(ctx context.Context) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("this command needs root privileges; run it with sudo")
	}

	arg := common.NewArgument()
	if !arg.SystemInfo.IsLinux() {
		return fmt.Errorf("only Linux nodes can be added to an Olares cluster")
	}
	arg.SetOlaresVersion(version.VERSION)
	arg.SetOlaresCDNService(arg.OlaresCDNService)
	if arg.OlaresVersion == "" {
		return fmt.Errorf("this olares-cli build does not carry a usable Olares version")
	}
	arg.SetConsoleLog(joinNodeLogFile, true)
	common.SetConsoleLogOverride(joinNodeLogFile)

	// The cluster's version is not known yet, so name the installer's instead of
	// implying this is what the cluster runs.
	fmt.Printf("Joining this machine to an Olares cluster as a worker node, using the Olares %s installer.\n",
		arg.OlaresVersion)

	installedPath := filepath.Join(arg.BaseDir, common.TerminusStateFileInstalled)
	if utils.IsExist(installedPath) {
		installedVersion := utils.StateFileVersion(installedPath)
		if installedVersion == "" {
			installedVersion = "unknown"
		}
		return fmt.Errorf(
			"Olares %s is already installed on this node; run 'sudo olares-cli uninstall' before joining it to another cluster",
			installedVersion)
	}

	if err := resolveMasterConnection(arg); err != nil {
		return err
	}

	info, err := probeMasterInfo(ctx, arg, true)
	if err != nil {
		return err
	}
	arg.SetKubeVersion(info.KubernetesType)
	// The CDN endpoint is region-specific and lives only in the cluster, so a
	// worker that has nothing but the compiled-in default adopts the master's.
	// Anything supplied for this node specifically -- via --cdn-service or
	// OLARES_SYSTEM_CDN_SERVICE, which is what the generated join command sets
	// -- is left alone, so a local mirror still wins.
	if info.CDNService != "" && arg.OlaresCDNService == cc.DefaultOlaresCDNService {
		logger.Infof("using the CDN endpoint configured on the master: %s", info.CDNService)
		arg.SetOlaresCDNService(info.CDNService)
	}

	if err := ensureJoinableHostname(ctx, arg, info.AllNodes); err != nil {
		return err
	}

	// From here on the downstream pipelines build their own Argument from viper,
	// so everything resolved above has to be published there first. This also
	// has to happen before the cleanup below, which wipes master.conf.
	applyJoinConfigToViper(arg)

	if err := ensurePreparedForVersion(ctx, arg); err != nil {
		return err
	}

	logger.Infof("joining the cluster at %s ...", arg.MasterHost)
	if err := AddNodePipeline(ctx); err != nil {
		return fmt.Errorf("join worker to cluster: %w", err)
	}

	fmt.Printf("\nThis node joined the Olares cluster at %s as %q.\n",
		arg.MasterHost, arg.SystemInfo.GetHostname())
	fmt.Printf("Verify it on the master with: sudo /usr/local/bin/kubectl get nodes\n")
	return nil
}

// resolveMasterConnection works out how this worker reaches the master, asking
// for whatever it cannot determine and proving the result actually works before
// anything is downloaded.
func resolveMasterConnection(arg *common.Argument) error {
	if err := resolveMasterConnectionSources(arg); err != nil {
		return err
	}

	host, err := resolveMasterHost(arg.MasterHost)
	if err != nil {
		return err
	}
	arg.MasterHost = host

	if !hasMasterCredential(arg.MasterHostConfig) {
		arg.MasterSSHPrivateKeyPath = defaultPrivateKeyPath()
	}
	if hasMasterCredential(arg.MasterHostConfig) {
		verifyErr := verifyMasterSSH(arg.MasterHostConfig)
		if verifyErr == nil {
			logger.Infof("reaching the master at %s@%s:%d",
				arg.MasterSSHUser, arg.MasterHost, arg.MasterSSHPort)
			return nil
		}
		fmt.Fprintf(os.Stderr, "Could not reach the master as %s@%s with the credentials at hand: %v\n",
			arg.MasterSSHUser, arg.MasterHost, verifyErr)
	}
	return promptMasterCredentials(arg.MasterHostConfig)
}

// resolveMasterConnectionSources layers the non-interactive sources of the
// master connection onto the Argument.
//
// Sources are layered whole-config rather than merged per field across all of
// them: taking the host from one source and the credentials from another
// produces a connection that cannot work and an error naming the wrong machine.
//  1. master.conf, if this node already knows a master (lowest, so that a bare
//     `node join` resumes an interrupted attempt without asking again)
//  2. the MASTER_AUTH_INFO payload from `node join-command`
//  3. explicit flags / environment variables
//
// The ordering is what matters: an operator pointing this node at a cluster
// always outranks whatever the node remembers, so a master.conf left behind by
// an earlier membership can never silently redirect the join.
func resolveMasterConnectionSources(arg *common.Argument) error {
	explicit := common.MasterHostConfig{
		MasterHost:              strings.TrimSpace(viper.GetString(common.FlagMasterHost)),
		MasterSSHUser:           strings.TrimSpace(viper.GetString(common.FlagMasterSSHUser)),
		MasterSSHPassword:       viper.GetString(common.FlagMasterSSHPassword),
		MasterSSHPrivateKeyPath: strings.TrimSpace(viper.GetString(common.FlagMasterSSHPrivateKeyPath)),
		MasterSSHPort:           viper.GetInt(common.FlagMasterSSHPort),
	}

	if payload := strings.TrimSpace(viper.GetString(common.FlagMasterAuthInfo)); payload != "" {
		decoded, err := decodeMasterAuthInfo(payload)
		if err != nil {
			return err
		}
		// A payload names a specific cluster, so it replaces anything remembered
		// rather than being merged into it.
		arg.ClearMasterHostConfig()
		arg.SetMasterHostOverride(*decoded)
	}
	if explicit.MasterHost != "" && explicit.MasterHost != arg.MasterHost {
		// Likewise for an explicitly named, different master.
		arg.ClearMasterHostConfig()
	}
	arg.SetMasterHostOverride(explicit)

	if arg.MasterSSHPort == 0 {
		arg.MasterSSHPort = 22
	}
	if arg.MasterSSHPort < 1 || arg.MasterSSHPort > 65535 {
		return fmt.Errorf("master SSH port must be between 1 and 65535, got %d", arg.MasterSSHPort)
	}
	if arg.MasterSSHUser == "" {
		arg.MasterSSHUser = "root"
	}
	return nil
}

// resolveMasterHost validates the master address, prompting for a replacement
// while it is unusable. Only a routable IPv4 address works: the joining node
// reaches the master over the LAN, and both Kubernetes and the SSH connection
// need an address that resolves the same way from here.
func resolveMasterHost(host string) (string, error) {
	host = strings.TrimSpace(host)
	for {
		if host != "" {
			if ip := net.ParseIP(host); ip != nil && ip.To4() != nil && !ip.IsLoopback() {
				return host, nil
			}
			fmt.Fprintf(os.Stderr, "%q is not a reachable LAN IPv4 address.\n", host)
		}
		var err error
		host, err = promptNonEmptyLine("Master LAN IPv4 address: ", "")
		if err != nil {
			return "", err
		}
	}
}

// ensureJoinableHostname makes sure this machine's hostname can serve as a
// Kubernetes node name in the target cluster, renaming it when it cannot.
// Olares One units all ship with the same hostname, so a collision is the
// normal case here rather than an edge case; --node-name exists so that this
// step, too, can be answered up front by an unattended setup.
func ensureJoinableHostname(ctx context.Context, arg *common.Argument, existingNodeNames []string) error {
	current := strings.ToLower(strings.TrimSpace(arg.SystemInfo.GetHostname()))

	if requested := strings.ToLower(strings.TrimSpace(viper.GetString(common.FlagNodeName))); requested != "" {
		if reason := joinHostnameProblem(requested, existingNodeNames); reason != "" {
			return fmt.Errorf("the requested node name %q cannot be used: %s", requested, reason)
		}
		if requested == current {
			logger.Infof("using worker node name %q", current)
			return nil
		}
		return applyJoinHostname(ctx, arg, requested)
	}

	reason := joinHostnameProblem(current, existingNodeNames)
	if reason == "" {
		logger.Infof("using worker node name %q", current)
		return nil
	}
	fmt.Fprintf(os.Stderr, "\nThis machine's hostname %q cannot be used as its node name: %s\n",
		current, reason)

	for {
		candidate, err := promptNonEmptyLine("New hostname for this worker: ", "")
		if err != nil {
			return err
		}
		candidate = strings.ToLower(strings.TrimSpace(candidate))
		if reason := joinHostnameProblem(candidate, existingNodeNames); reason != "" {
			fmt.Fprintf(os.Stderr, "%q cannot be used: %s\n", candidate, reason)
			continue
		}
		return applyJoinHostname(ctx, arg, candidate)
	}
}

func applyJoinHostname(ctx context.Context, arg *common.Argument, hostname string) error {
	if err := setHostname(ctx, hostname); err != nil {
		return err
	}
	// Later stages build their own SystemInfo and re-read the hostname from the
	// kernel, but this one is already loaded and backs the current Argument.
	arg.SystemInfo.SetHostname(hostname)
	logger.Infof("hostname of this worker changed to %q", hostname)
	return nil
}

// setHostname renames the machine, and refreshes the 127.0.1.1 entry of
// /etc/hosts when there is one, the same way ConfigureOS does later in the
// install: on the distributions that use that entry, leaving the old name there
// makes every sudo call in between warn that it cannot resolve the host. Images
// without a 127.0.1.1 line (most Ubuntu cloud images) resolve their hostname
// some other way and need nothing here, so the edit is a no-op for them.
func setHostname(ctx context.Context, hostname string) error {
	script := fmt.Sprintf(
		"hostnamectl set-hostname %s && sed -i '/^127.0.1.1/s/.*/127.0.1.1      %s/g' /etc/hosts",
		hostname, hostname)
	output, err := exec.CommandContext(ctx, "/bin/bash", "-c", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("set hostname to %q: %w (output: %s)", hostname, err, strings.TrimSpace(string(output)))
	}
	return nil
}

// joinHostnameProblem describes why a hostname cannot be used as this cluster's
// node name, or returns an empty string when it can.
func joinHostnameProblem(hostname string, existingNodeNames []string) string {
	if errs := validation.IsDNS1123Subdomain(hostname); len(errs) > 0 {
		return strings.Join(errs, "; ")
	}
	for _, existing := range existingNodeNames {
		if strings.EqualFold(existing, hostname) {
			return fmt.Sprintf("node name %q already exists in the master cluster", hostname)
		}
	}
	return ""
}

// ensurePreparedForVersion leaves this node with its dependencies prepared for
// exactly the cluster's Olares version, which is what the add-node pipeline
// requires. A prepare state from a different version is cleaned out without
// asking: it cannot be reconciled in place, and nothing is at stake because a
// merely prepared node holds no user data and runs no cluster (JoinNodePipeline
// has already refused to touch a node with Olares installed).
func ensurePreparedForVersion(ctx context.Context, arg *common.Argument) error {
	preparedPath := filepath.Join(arg.BaseDir, common.TerminusStateFilePrepared)
	preparedVersion := utils.StateFileVersion(preparedPath)

	if utils.IsExist(preparedPath) && preparedVersion != arg.OlaresVersion {
		stale := preparedVersion
		if stale == "" {
			stale = "an unknown version"
		}
		logger.Warnf("this node is prepared for %s, but the cluster runs Olares %s; "+
			"removing the existing prepare state and downloading the matching version",
			stale, arg.OlaresVersion)
		// A prepare state for another version cannot be reconciled in place, so
		// the cleanup must not be limited to a single phase.
		viper.Set(common.FlagUninstallAll, true)
		if err := UninstallTerminusPipeline(ctx); err != nil {
			return fmt.Errorf("clean the prepare state of %s: %w", stale, err)
		}
		preparedVersion = ""
	}

	if preparedVersion == arg.OlaresVersion {
		logger.Infof("Olares %s is already prepared on this node, skipping download and prepare", preparedVersion)
		return nil
	}

	logger.Infof("preparing this node for Olares %s; the download can take a while", arg.OlaresVersion)
	if err := StartPreCheckPipeline(); err != nil {
		return fmt.Errorf("worker precheck failed: %w", err)
	}
	if err := DownloadInstallationWizard(); err != nil {
		return fmt.Errorf("download installation wizard: %w", err)
	}
	if err := DownloadInstallationPackage(); err != nil {
		return fmt.Errorf("download installation packages: %w", err)
	}
	if err := PrepareSystemPipeline(ctx, nil); err != nil {
		return fmt.Errorf("prepare worker: %w", err)
	}
	return nil
}

// applyJoinConfigToViper publishes the resolved configuration so the pipelines
// invoked afterwards, which each build their own Argument from viper, agree with
// the decisions made here.
func applyJoinConfigToViper(arg *common.Argument) {
	viper.Set(common.FlagVersion, arg.OlaresVersion)
	viper.Set(common.FlagBaseDir, arg.BaseDir)
	viper.Set(common.FlagCDNService, arg.OlaresCDNService)
	viper.Set(common.FlagKubeType, arg.Kubetype)
	viper.Set(common.FlagMasterHost, arg.MasterHost)
	viper.Set(common.FlagMasterSSHUser, arg.MasterSSHUser)
	viper.Set(common.FlagMasterSSHPassword, arg.MasterSSHPassword)
	viper.Set(common.FlagMasterSSHPrivateKeyPath, arg.MasterSSHPrivateKeyPath)
	viper.Set(common.FlagMasterSSHPort, arg.MasterSSHPort)
}
