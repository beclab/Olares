package network

import (
	"fmt"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/pkg/errors"
)

const (
	// OverlayParentAltname is the alternative interface name the underlay
	// NetworkAttachmentDefinition uses as its macvlan master. Every node adds
	// it to the wired NIC that faces the LAN, so one cluster-wide NAD resolves
	// to the right device on each node without renaming anything.
	OverlayParentAltname = "olares-lan"

	// OverlayDesiredStateFile marks that the user has enabled the overlay
	// gateway on this node. olaresd converges the node towards it on boot.
	OverlayDesiredStateFile = "/var/lib/olares/overlay-gateway/enabled"
	// OverlayUdevRuleFile re-adds the alternative name whenever the NIC
	// appears, before NetworkManager and olaresd start. A udev rule is used
	// rather than a systemd.link file because udev applies only the first
	// matching .link file per device and netplan already ships one for every
	// NetworkManager connection.
	OverlayUdevRuleFile = "/etc/udev/rules.d/80-olares-lan.rules"
	// legacyOverlayLinkFile is the earlier persistence file; it is removed
	// whenever the rule is written or the alternative name is dropped.
	legacyOverlayLinkFile = "/etc/systemd/network/10-olares-lan.link"
	// overlayMigratedMarker records that the bridge was active before the
	// upgrade tore it down and names the NIC that left the bridge, so the later
	// upgrade steps know which device to name and that the desired state and
	// the overlay Pods must be restored even when re-run.
	overlayMigratedMarker = "/var/lib/olares/overlay-gateway/.migrated-from-bridge"

	overlayBridgeConnection   = "br-olares"
	overlayBridgeSlavePrefix  = "br-olares-slave-"
	overlayOriginalConnection = "original-connection"

	overlayMigrateVerifyAttempts = 30
	overlayMigrateVerifyInterval = time.Second
)

// commandRunner executes one privileged shell command on the node. It is a
// function so the migration logic can be unit-tested without NetworkManager.
type commandRunner func(cmd string) (string, error)

func sudoRunner(runtime connector.Runtime) commandRunner {
	return func(cmd string) (string, error) {
		return runtime.GetRunner().SudoCmd(cmd, false, false)
	}
}

// overlayUdevRuleContent renders the udev rule that re-adds the alternative
// name when the NIC with this MAC appears. ipPath must be absolute: udev does
// not search PATH. $env{INTERFACE} is the final interface name after any
// rename, which %k is not guaranteed to be.
func overlayUdevRuleContent(mac, ipPath string) string {
	return fmt.Sprintf(`ACTION=="add", SUBSYSTEM=="net", ATTR{address}=="%s", RUN+="%s link property add dev $env{INTERFACE} altname %s"`,
		mac, ipPath, OverlayParentAltname)
}

// overlayUdevRuleCommand writes the rule through the node runner. The runner
// wraps every command in `bash -c "..."`, inside which single quotes do not
// stop parameter expansion, so the `$` of $env{INTERFACE} must be escaped or
// bash would expand it to an empty variable before the rule reaches the file.
func overlayUdevRuleCommand(mac, ipPath string) string {
	rule := strings.ReplaceAll(overlayUdevRuleContent(mac, ipPath), "$", `\$`)
	return fmt.Sprintf("mkdir -p %s && printf '%%s\\n' '%s' > %s && rm -f %s",
		parentDir(OverlayUdevRuleFile), rule, OverlayUdevRuleFile, legacyOverlayLinkFile)
}

// resolveIPCommand returns the absolute path of the ip utility on this node.
func resolveIPCommand(run commandRunner) (string, error) {
	out, err := run("command -v ip")
	path := strings.TrimSpace(out)
	if err != nil || path == "" {
		return "", errors.New("ip command not found")
	}
	return path, nil
}

// linkNameFromIPOutput extracts the primary interface name from
// `ip -o link show <name>` output such as "2: enp3s0: <BROADCAST,...".
func linkNameFromIPOutput(out string) string {
	line := strings.TrimSpace(strings.SplitN(out, "\n", 2)[0])
	parts := strings.SplitN(line, ":", 3)
	if len(parts) < 2 {
		return ""
	}
	name := strings.TrimSpace(parts[1])
	if at := strings.Index(name, "@"); at >= 0 {
		name = name[:at]
	}
	return name
}

// resolveOverlayAltname returns the device currently carrying the alternative
// name, or "" when no device has it.
func resolveOverlayAltname(run commandRunner) string {
	out, err := run("ip -o link show " + OverlayParentAltname)
	if err != nil {
		return ""
	}
	return linkNameFromIPOutput(out)
}

// ensureOverlayAltname puts the alternative name on phy, the NIC that just
// left the bridge, and persists it. The upgrade is the authoritative
// reconciliation for that node: a name found on another device can only be a
// leftover, so it is moved with a warning instead of failing the upgrade.
func ensureOverlayAltname(run commandRunner, phy string) error {
	if bound := resolveOverlayAltname(run); bound != "" && bound != phy {
		logger.Warnf("overlay-parent: alternative name %s was on %s, moving it to %s", OverlayParentAltname, bound, phy)
		if _, err := run(fmt.Sprintf("ip link property del dev %s altname %s", bound, OverlayParentAltname)); err != nil {
			return errors.Wrapf(err, "remove alternative name %s from %s", OverlayParentAltname, bound)
		}
	} else if bound == phy {
		logger.Infof("overlay-parent: %s already carries alternative name %s", phy, OverlayParentAltname)
	}
	if resolveOverlayAltname(run) != phy {
		if _, err := run(fmt.Sprintf("ip link property add dev %s altname %s", phy, OverlayParentAltname)); err != nil {
			return errors.Wrapf(err, "add alternative name %s to %s", OverlayParentAltname, phy)
		}
	}
	mac, err := run("cat /sys/class/net/" + phy + "/address")
	if err != nil {
		return errors.Wrapf(err, "read MAC of %s", phy)
	}
	mac = strings.TrimSpace(mac)
	ipPath, err := resolveIPCommand(run)
	if err != nil {
		return err
	}
	if _, err := run(overlayUdevRuleCommand(mac, ipPath)); err != nil {
		return errors.Wrap(err, "write udev rule")
	}
	logger.Infof("overlay-parent: %s carries alternative name %s (mac %s), persisted in %s", phy, OverlayParentAltname, mac, OverlayUdevRuleFile)
	return nil
}

func removeOverlayAltname(run commandRunner) {
	_, _ = run("rm -f " + OverlayUdevRuleFile + " " + legacyOverlayLinkFile)
	if dev := resolveOverlayAltname(run); dev != "" {
		_, _ = run(fmt.Sprintf("ip link property del dev %s altname %s", dev, OverlayParentAltname))
	}
}

// overlayBridgeState is what the migration learned about the legacy bridge.
type overlayBridgeState struct {
	Exists bool
	Active bool
	Slaves []string
	Phy    string
}

func parseNMConnectionNames(out string) []string {
	var names []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			names = append(names, line)
		}
	}
	return names
}

func probeOverlayBridge(run commandRunner) (overlayBridgeState, error) {
	var st overlayBridgeState
	out, err := run("nmcli -g NAME connection show")
	if err != nil {
		return st, fmt.Errorf("list network connections: %w", err)
	}
	for _, name := range parseNMConnectionNames(out) {
		if name == overlayBridgeConnection {
			st.Exists = true
		}
		if strings.HasPrefix(name, overlayBridgeSlavePrefix) {
			st.Slaves = append(st.Slaves, name)
			if st.Phy != "" {
				return st, fmt.Errorf("multiple physical bridge profiles require explicit migration")
			}
			st.Phy = strings.TrimPrefix(name, overlayBridgeSlavePrefix)
		}
	}
	if !st.Exists {
		return st, nil
	}
	out, err = run("nmcli -g GENERAL.STATE connection show " + overlayBridgeConnection)
	if err != nil {
		return st, fmt.Errorf("read bridge state: %w", err)
	}
	switch strings.TrimSpace(out) {
	case "activated":
		st.Active = true
	case "", "deactivated":
	default:
		return st, fmt.Errorf("bridge state is not stable: %q", strings.TrimSpace(out))
	}
	if st.Phy == "" {
		out, err = run("ip -o link show master " + overlayBridgeConnection)
		if err != nil {
			return st, fmt.Errorf("read bridge parent: %w", err)
		}
		st.Phy = linkNameFromIPOutput(out)
	}
	return st, nil
}

func hasNMConnection(run commandRunner, name string) (bool, error) {
	out, err := run("nmcli -g NAME connection show")
	if err != nil {
		return false, err
	}
	for _, n := range parseNMConnectionNames(out) {
		if n == name {
			return true, nil
		}
	}
	return false, nil
}

func phyHasAddressAndDefaultRoute(run commandRunner, phy string) bool {
	addr, err := run("ip -4 -o addr show dev " + phy + " scope global")
	if err != nil || strings.TrimSpace(addr) == "" {
		return false
	}
	route, err := run("ip -4 route show default dev " + phy)
	return err == nil && strings.TrimSpace(route) != ""
}

func deleteOverlayBridgeProfiles(run commandRunner, st overlayBridgeState) error {
	names := append([]string(nil), st.Slaves...)
	if st.Exists {
		names = append(names, overlayBridgeConnection)
	}
	for _, name := range names {
		if _, err := run("nmcli connection delete " + name); err != nil {
			return fmt.Errorf("delete bridge profile %s: %w", name, err)
		}
	}
	return nil
}

// migrateBridgeToDirect persists intent before disconnecting the bridge. Every
// subsequent attempt converges to the physical parent; it never reactivates the bridge.
func migrateBridgeToDirect(run commandRunner, sleep func(time.Duration)) (bool, error) {
	phy, resuming, err := overlayMigratedFromBridge(run)
	if err != nil {
		return false, err
	}
	var st overlayBridgeState
	for attempt := 0; attempt < 3; attempt++ {
		st, err = probeOverlayBridge(run)
		if err == nil {
			break
		}
		if attempt < 2 {
			sleep(time.Second)
		}
	}
	if err != nil {
		return false, err
	}
	if !resuming && !st.Exists {
		return false, nil
	}
	if !resuming && !st.Active {
		if st.Phy == "" || !phyHasAddressAndDefaultRoute(run, st.Phy) {
			return false, fmt.Errorf("cannot clean inactive bridge before verifying physical connectivity")
		}
		return false, deleteOverlayBridgeProfiles(run, st)
	}
	if !resuming {
		phy = st.Phy
	}
	if !validOverlayInterface(phy) {
		return false, fmt.Errorf("invalid or missing migration parent")
	}
	exists, err := hasNMConnection(run, overlayOriginalConnection)
	if err != nil {
		return false, fmt.Errorf("inspect physical connection: %w", err)
	}
	if !exists {
		method, err := run("nmcli -g ipv4.method connection show " + overlayBridgeConnection)
		if err != nil {
			return false, fmt.Errorf("inspect bridge address configuration: %w", err)
		}
		if strings.TrimSpace(method) != "auto" {
			return false, fmt.Errorf("physical connection is missing; cannot migrate a non-DHCP bridge automatically")
		}
		if _, err := run(fmt.Sprintf("nmcli connection add type ethernet ifname %s con-name %s ipv4.method auto ipv6.method auto connection.autoconnect no", phy, overlayOriginalConnection)); err != nil {
			return false, err
		}
	}
	if exists {
		iface, err := run("nmcli -g connection.interface-name connection show " + overlayOriginalConnection)
		if err != nil {
			return false, fmt.Errorf("inspect physical profile interface: %w", err)
		}
		if strings.TrimSpace(iface) != phy {
			return false, fmt.Errorf("physical profile does not match migration parent")
		}
	}
	if !resuming {
		if _, err := run(fmt.Sprintf("mkdir -p %s && printf '%%s' '%s' > %s.tmp && sync -f %s.tmp && mv -f %s.tmp %s && sync -f %s", parentDir(overlayMigratedMarker), phy, overlayMigratedMarker, overlayMigratedMarker, overlayMigratedMarker, overlayMigratedMarker, parentDir(overlayMigratedMarker))); err != nil {
			return false, fmt.Errorf("persist migration intent: %w", err)
		}
	}
	if st.Exists {
		if _, err := run("nmcli connection modify " + overlayBridgeConnection + " connection.autoconnect no"); err != nil {
			return false, err
		}
	}
	for _, slave := range st.Slaves {
		if _, err := run("nmcli connection modify " + slave + " connection.autoconnect no"); err != nil {
			return false, err
		}
	}
	if _, err := run("nmcli connection modify " + overlayOriginalConnection + " connection.autoconnect yes"); err != nil {
		return false, err
	}
	if st.Active {
		if _, err := run("nmcli connection down " + overlayBridgeConnection); err != nil {
			return false, fmt.Errorf("deactivate bridge; retry to continue migration: %w", err)
		}
	}
	// Do not bounce an already recovered physical connection on a resumed attempt.
	if !phyHasAddressAndDefaultRoute(run, phy) {
		if _, err := run("nmcli connection up " + overlayOriginalConnection); err != nil {
			return false, fmt.Errorf("activate physical parent; retry to continue migration: %w", err)
		}
	}
	for i := 0; i < overlayMigrateVerifyAttempts; i++ {
		if phyHasAddressAndDefaultRoute(run, phy) {
			if st.Exists || len(st.Slaves) > 0 {
				if err := deleteOverlayBridgeProfiles(run, st); err != nil {
					return false, err
				}
			}
			return true, nil
		}
		sleep(overlayMigrateVerifyInterval)
	}
	return false, fmt.Errorf("physical parent has no IPv4 address/default route; migration intent retained for forward recovery")
}

func validOverlayInterface(name string) bool {
	if len(name) == 0 || len(name) > 15 {
		return false
	}
	for _, c := range name {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '.' || c == '-') {
			return false
		}
	}
	return name[0] != '-'
}

func parentDir(path string) string {
	if i := strings.LastIndex(path, "/"); i > 0 {
		return path[:i]
	}
	return "/"
}

// overlayMigratedFromBridge reports whether this upgrade tore down an active
// bridge and, if so, which NIC left it.
func overlayMigratedFromBridge(run commandRunner) (string, bool, error) {
	out, err := run("if test -e " + overlayMigratedMarker + "; then test -s " + overlayMigratedMarker + " && cat " + overlayMigratedMarker + "; fi")
	if err != nil {
		return "", false, fmt.Errorf("read migration intent: %w", err)
	}
	phy := strings.TrimSpace(out)
	if phy == "" {
		return "", false, nil
	}
	if !validOverlayInterface(phy) {
		return "", false, fmt.Errorf("invalid migration parent in intent")
	}
	return phy, true, nil
}

func migrationRunner(runtime connector.Runtime) commandRunner {
	return func(cmd string) (string, error) {
		if strings.HasPrefix(cmd, "nmcli ") || strings.HasPrefix(cmd, "ip ") {
			cmd = "timeout 30s " + cmd
		}
		return sudoRunner(runtime)(cmd)
	}
}

// ResumeOverlayMigration restores target connectivity before any Kubernetes API
// access. Without durable migration intent this is a read-only no-op.
func ResumeOverlayMigration(runtime connector.Runtime) error {
	run := migrationRunner(runtime)
	_, pending, err := overlayMigratedFromBridge(run)
	if err != nil || !pending {
		return err
	}
	_, err = migrateBridgeToDirect(run, time.Sleep)
	return err
}

// MigrateBridgeToDirect is the upgrade task that retires the legacy overlay
// bridge on this node.
type MigrateBridgeToDirect struct {
	common.KubeAction
}

func (m *MigrateBridgeToDirect) Execute(runtime connector.Runtime) error {
	migrated, err := migrateBridgeToDirect(migrationRunner(runtime), time.Sleep)
	if err != nil {
		return err
	}
	if !migrated {
		logger.Infof("overlay-parent: no active bridge to migrate")
	}
	return nil
}

// EnsureOverlayAltname names the NIC that left the bridge and persists the
// name. Nodes that had no active bridge are left untouched: their alternative
// name is added by olaresd when the user enables the overlay gateway.
type EnsureOverlayAltname struct {
	common.KubeAction
}

func (e *EnsureOverlayAltname) Execute(runtime connector.Runtime) error {
	run := sudoRunner(runtime)
	phy, migrated, err := overlayMigratedFromBridge(run)
	if err != nil {
		return err
	}
	if !migrated {
		logger.Infof("overlay-parent: no bridge was migrated, alternative name left to olaresd")
		return nil
	}
	if phy == "" {
		logger.Warnf("overlay-parent: migration marker names no interface, alternative name left to olaresd")
		return nil
	}
	return ensureOverlayAltname(run, phy)
}

// WriteOverlayDesiredIfMigrated records the overlay gateway as enabled when the
// upgrade tore down an active bridge, so the node ends the upgrade in the same
// user-visible state it started in.
type WriteOverlayDesiredIfMigrated struct {
	common.KubeAction
}

func (w *WriteOverlayDesiredIfMigrated) Execute(runtime connector.Runtime) error {
	run := sudoRunner(runtime)
	_, migrated, err := overlayMigratedFromBridge(run)
	if err != nil {
		return err
	}
	if !migrated {
		return nil
	}
	if _, err := run(fmt.Sprintf("mkdir -p %s && touch %s", parentDir(OverlayDesiredStateFile), OverlayDesiredStateFile)); err != nil {
		return errors.Wrap(err, "write overlay gateway desired state")
	}
	return nil
}

// ClearOverlayMigrationMarker is the last migration step; it runs once the
// overlay Pods have been recreated against the new parent.
type ClearOverlayMigrationMarker struct {
	common.KubeAction
}

func (c *ClearOverlayMigrationMarker) Execute(runtime connector.Runtime) error {
	_, err := sudoRunner(runtime)("rm -f " + overlayMigratedMarker)
	return err
}

// OverlayMigratedFromBridge reports whether the current upgrade tore down an
// active bridge on this node.
func OverlayMigratedFromBridge(runtime connector.Runtime) (bool, error) {
	_, migrated, err := overlayMigratedFromBridge(sudoRunner(runtime))
	return migrated, err
}

// RemoveOverlayAltname is the uninstall task that drops the alternative name
// and its persistence.
type RemoveOverlayAltname struct {
	common.KubeAction
}

func (r *RemoveOverlayAltname) Execute(runtime connector.Runtime) error {
	removeOverlayAltname(sudoRunner(runtime))
	return nil
}
