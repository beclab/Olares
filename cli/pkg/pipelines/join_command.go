package pipelines

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/storage"
	"github.com/beclab/Olares/cli/pkg/terminus"
	"github.com/spf13/viper"
)

// JoinCommandOptions holds the choices that change how the command is produced
// rather than what ends up in it.
type JoinCommandOptions struct {
	// AssumeYes takes the detected address and SSH account as they are and
	// skips the SSH check, so nothing is asked and nothing is dialed.
	AssumeYes bool
	// Quiet prints the worker command and nothing else, so the output can be
	// captured verbatim.
	Quiet bool
}

// narrate explains what is happening, on stderr so it can never end up in a
// captured command, and not at all in quiet mode.
func (o JoinCommandOptions) narrate(format string, args ...any) {
	if o.Quiet {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

// workerCommandSpec is what goes into the rendered worker command.
type workerCommandSpec struct {
	Version    string
	CDNService string
	// MasterAuthInfo is the opaque token carrying the master address and SSH
	// credentials the worker needs.
	MasterAuthInfo string
}

// JoinCommand gathers this master's cluster information and prints the
// self-contained command to run on a new worker.
//
// The command itself goes to stdout and everything else to stderr, so
// `olares-cli node join-command > cmd.sh` yields something directly runnable
// even without --quiet.
func JoinCommand(ctx context.Context, opts JoinCommandOptions) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("this command needs root privileges; run 'sudo olares-cli node join-command'")
	}
	return joinCommand(ctx, common.NewArgument(), opts)
}

func joinCommand(ctx context.Context, arg *common.Argument, opts JoinCommandOptions) error {
	if !arg.SystemInfo.IsLinux() {
		return fmt.Errorf("the worker join command can only be generated on a Linux Olares master")
	}
	arg.SetOlaresCDNService(arg.OlaresCDNService)

	olaresVersion, err := phase.GetOlaresVersion()
	if err != nil {
		return fmt.Errorf("this machine is not an active Olares master: %w", err)
	}
	olaresVersion = strings.TrimSpace(olaresVersion)
	if !storage.IsJuiceFSEnabled() {
		return fmt.Errorf("JuiceFS is not enabled on this Olares master; run 'sudo olares-cli node enable-juicefs' and retry")
	}

	masterHost, err := resolveAdvertisedMasterHost(arg)
	if err != nil {
		return err
	}

	cdnService, err := resolveJoinCDNService(ctx, arg, opts)
	if err != nil {
		return err
	}

	// Settle the credentials only once the rest of the command is known to be
	// valid, so nobody types a password just to be told the address was unusable.
	masterConfig := &common.MasterHostConfig{
		MasterHost:    masterHost,
		MasterSSHUser: masterSSHUserSuggestion(arg),
		MasterSSHPort: masterSSHPort(arg),
	}
	if err := resolveMasterSSHAccess(masterConfig, opts); err != nil {
		return err
	}

	command := buildJoinCommand(workerCommandSpec{
		Version:        olaresVersion,
		CDNService:     cdnService,
		MasterAuthInfo: encodeMasterAuthInfo(masterConfig),
	})

	opts.narrate("\nRun the following command on the worker node:\n\n")
	fmt.Println(command)
	if masterConfig.MasterSSHPassword != "" {
		opts.narrate("\n%s is Base64-encoded, not encrypted. Anyone holding this command can recover the master's SSH credentials, so share it only with the intended worker administrator.\n",
			common.ENV_MASTER_AUTH_INFO)
	} else {
		opts.narrate("\nThe command carries no password. The worker will take one from %s or %s, or ask for it.\n",
			common.ENV_MASTER_SSH_PASSWORD, common.ENV_MASTER_SSH_PRIVATE_KEY_PATH)
	}
	return nil
}

// resolveMasterSSHAccess settles on the SSH credentials the worker will use to
// reach this master.
//
// Interactively it confirms the account and proves the password works, so a bad
// credential surfaces here rather than midway through the worker's install.
// Either flag makes the command non-interactive instead, which is what makes it
// usable from a script:
//
//	--master-ssh-password  use it without asking, and still prove it works
//	--yes                  take the detected account and embed no password at
//	                       all, leaving the worker to supply one
//	--yes with a password  use it without asking and without dialing the master
func resolveMasterSSHAccess(cfg *common.MasterHostConfig, opts JoinCommandOptions) error {
	password := viper.GetString(common.FlagMasterSSHPassword)

	if opts.AssumeYes || password != "" {
		cfg.MasterSSHPassword = password
		if opts.AssumeYes || password == "" {
			// Nothing to prove, or the operator asserted it already.
			return nil
		}
		if err := verifyMasterSSH(cfg); err != nil {
			return fmt.Errorf(
				"could not verify SSH login and sudo access for %s@%s:%d: %w\n"+
					"pass --yes as well to embed the password without verifying it",
				cfg.MasterSSHUser, cfg.MasterHost, cfg.MasterSSHPort, err)
		}
		opts.narrate("SSH login and sudo access verified for %s@%s:%d.\n",
			cfg.MasterSSHUser, cfg.MasterHost, cfg.MasterSSHPort)
		return nil
	}

	// The password is needed because a worker joins by talking to the master over
	// SSH, which is not obvious to someone who is already root on this machine,
	// so it is spelled out before asking.
	opts.narrate("A worker joins by connecting back to this master over SSH, so the command\n"+
		"needs an account on this machine that workers can use.\n\n"+
		"  address: %s:%d\n  user:    %s\n\n",
		cfg.MasterHost, cfg.MasterSSHPort, cfg.MasterSSHUser)

	useSuggested, err := promptConfirmation("Use this account? [Y/n]: ", true)
	if err != nil {
		return err
	}
	if !useSuggested {
		cfg.MasterSSHUser = ""
	}
	if err := promptMasterCredentials(cfg); err != nil {
		return err
	}
	opts.narrate("SSH login and sudo access verified for %s@%s:%d.\n",
		cfg.MasterSSHUser, cfg.MasterHost, cfg.MasterSSHPort)
	return nil
}

// resolveJoinCDNService determines which CDN endpoint the worker should download
// from.
//
// The cluster's own OLARES_SYSTEM_CDN_SERVICE setting is authoritative: it is
// region-specific (a .cn deployment uses a different endpoint), user-editable
// in Settings, and stored only in the cluster, so it cannot be recovered from
// the local process environment. Handing the worker the compiled-in default
// instead would send it to an endpoint it may not be able to reach at all.
//
// Anything passed to this command explicitly still wins, for staging or mirror
// setups.
func resolveJoinCDNService(ctx context.Context, arg *common.Argument, opts JoinCommandOptions) (string, error) {
	if arg.OlaresCDNService != cc.DefaultOlaresCDNService {
		return arg.OlaresCDNService, validateCDNService(arg.OlaresCDNService)
	}
	clusterCDN, err := terminus.ClusterCDNService(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read the cluster's %s setting: %w\n"+
			"pass --cdn-service explicitly to bypass this lookup",
			common.ENV_OLARES_CDN_SERVICE, err)
	}
	if clusterCDN == "" {
		opts.narrate("This cluster does not declare %s; the worker will use the default CDN endpoint.\n",
			common.ENV_OLARES_CDN_SERVICE)
		return arg.OlaresCDNService, nil
	}
	return clusterCDN, validateCDNService(clusterCDN)
}

// resolveAdvertisedMasterHost picks the address workers will use to reach this
// master. Auto-detection covers the common single-NIC case; --master-host is for
// the machines where it guesses a network the workers cannot reach.
func resolveAdvertisedMasterHost(arg *common.Argument) (string, error) {
	if explicit := strings.TrimSpace(arg.MasterHost); explicit != "" {
		if err := validateMasterHost(explicit); err != nil {
			return "", err
		}
		return explicit, nil
	}
	detected := strings.TrimSpace(arg.SystemInfo.GetLocalIp())
	if err := validateMasterHost(detected); err != nil {
		return "", err
	}
	return detected, nil
}

// masterSSHUserSuggestion proposes the account the worker should use to reach
// this master, preferring the human who invoked sudo over root.
func masterSSHUserSuggestion(arg *common.Argument) string {
	if user := strings.TrimSpace(arg.MasterSSHUser); user != "" {
		return user
	}
	if sudoUser := strings.TrimSpace(os.Getenv("SUDO_USER")); sudoUser != "" && sudoUser != "root" {
		return sudoUser
	}
	if user := strings.TrimSpace(arg.SystemInfo.GetUsername()); user != "" {
		return user
	}
	return "root"
}

func masterSSHPort(arg *common.Argument) int {
	if arg.MasterSSHPort != 0 {
		return arg.MasterSSHPort
	}
	return 22
}

func validateMasterHost(masterHost string) error {
	if ip := net.ParseIP(masterHost); ip == nil || ip.To4() == nil || ip.IsLoopback() {
		return fmt.Errorf("the detected master address %q is not a reachable IPv4 address; set OS_LOCALIP to the LAN IPv4 address workers can reach", masterHost)
	}
	return nil
}

func validateCDNService(cdnService string) error {
	if cdnService == "" {
		return nil
	}
	cdnURL, err := url.ParseRequestURI(cdnService)
	if err != nil || (cdnURL.Scheme != "http" && cdnURL.Scheme != "https") || cdnURL.Host == "" {
		return fmt.Errorf("CDN service %q is invalid; use an http:// or https:// URL", cdnService)
	}
	return nil
}

// buildJoinCommand renders the one-liner a worker administrator pastes into a
// terminal.
//
// The version travels in the script's URL rather than in the environment: the
// published script is per-version and carries it already. The CDN endpoint has
// to be passed, though, and is always included rather than only when it differs
// from the default: one script serves every region, it decides where every
// subsequent download comes from, and relying on the worker to guess the right
// regional endpoint is exactly the failure this command exists to prevent.
func buildJoinCommand(spec workerCommandSpec) string {
	cdn := strings.TrimRight(strings.TrimSpace(spec.CDNService), "/")
	env := []string{
		common.ENV_MASTER_AUTH_INFO + "=" + shellQuote(spec.MasterAuthInfo),
		common.ENV_OLARES_CDN_SERVICE + "=" + shellQuote(cdn),
	}
	return fmt.Sprintf(
		"export %s && curl -fsSL %s | bash",
		strings.Join(env, " "),
		shellQuote(terminus.JoinClusterScriptURL(cdn, spec.Version)),
	)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}
