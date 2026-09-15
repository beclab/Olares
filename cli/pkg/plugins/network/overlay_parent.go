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

func overlayUdevRuleCommand(mac, ipPath string) string {
	return fmt.Sprintf("mkdir -p %s && printf '%%s\\n' '%s' > %s && rm -f %s",
		parentDir(OverlayUdevRuleFile), overlayUdevRuleContent(mac, ipPath), OverlayUdevRuleFile, legacyOverlayLinkFile)
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

func probeOverlayBridge(run commandRunner) overlayBridgeState {
	var st overlayBridgeState
	if _, err := run("nmcli -g connection.type connection show " + overlayBridgeConnection); err != nil {
		return st
	}
	st.Exists = true
	if out, err := run("nmcli -g GENERAL.STATE connection show " + overlayBridgeConnection); err == nil && strings.Contains(out, "activated") {
		st.Active = true
	}
	if out, err := run("nmcli -g NAME connection show"); err == nil {
		for _, name := range parseNMConnectionNames(out) {
			if strings.HasPrefix(name, overlayBridgeSlavePrefix) {
				st.Slaves = append(st.Slaves, name)
				if st.Phy == "" {
					st.Phy = strings.TrimPrefix(name, overlayBridgeSlavePrefix)
				}
			}
		}
	}
	if st.Phy == "" {
		if out, err := run("ip -o link show master " + overlayBridgeConnection); err == nil {
			st.Phy = linkNameFromIPOutput(out)
		}
	}
	return st
}

func hasNMConnection(run commandRunner, name string) bool {
	out, err := run("nmcli -g NAME connection show")
	if err != nil {
		return false
	}
	for _, n := range parseNMConnectionNames(out) {
		if n == name {
			return true
		}
	}
	return false
}

func phyHasAddressAndDefaultRoute(run commandRunner, phy string) bool {
	addr, err := run("ip -4 -o addr show dev " + phy + " scope global")
	if err != nil || strings.TrimSpace(addr) == "" {
		return false
	}
	route, err := run("ip -4 route show default dev " + phy)
	return err == nil && strings.TrimSpace(route) != ""
}

func deleteOverlayBridgeProfiles(run commandRunner, st overlayBridgeState) {
	for _, s := range st.Slaves {
		_, _ = run("nmcli connection delete " + s)
	}
	_, _ = run("nmcli connection delete " + overlayBridgeConnection)
}

// migrateBridgeToDirect moves the node from the legacy bridge topology to the
// wired NIC. It reports whether an active bridge was torn down. When the bridge
// exists but is inactive, only its stale profiles are removed and the active
// physical connection is left untouched. When the switch does not leave the
// NIC with an address and a default route, the bridge is brought back and an
// error is returned so the upgrade stops before the node is changed further.
func migrateBridgeToDirect(run commandRunner, sleep func(time.Duration)) (bool, error) {
	st := probeOverlayBridge(run)
	if !st.Exists {
		return false, nil
	}
	if !st.Active {
		logger.Infof("overlay-parent: bridge %s exists but is inactive, removing stale profiles", overlayBridgeConnection)
		deleteOverlayBridgeProfiles(run, st)
		return false, nil
	}
	if st.Phy == "" {
		return false, fmt.Errorf("bridge %s is active but its physical slave could not be determined", overlayBridgeConnection)
	}
	if !hasNMConnection(run, overlayOriginalConnection) {
		if _, err := run(fmt.Sprintf("nmcli connection add type ethernet ifname %s con-name %s ipv4.method auto ipv6.method auto connection.autoconnect yes",
			st.Phy, overlayOriginalConnection)); err != nil {
			return false, errors.Wrap(err, "create the physical connection profile")
		}
	}
	if _, err := run(fmt.Sprintf("nmcli connection modify %s connection.autoconnect yes", overlayOriginalConnection)); err != nil {
		return false, errors.Wrap(err, "enable autoconnect on the physical connection")
	}
	if _, err := run(fmt.Sprintf("mkdir -p %s && printf '%%s' '%s' > %s", parentDir(overlayMigratedMarker), st.Phy, overlayMigratedMarker)); err != nil {
		return false, errors.Wrap(err, "write migration marker")
	}
	if _, err := run("nmcli connection down " + overlayBridgeConnection); err != nil {
		clearOverlayMigratedMarker(run)
		return false, errors.Wrap(err, "deactivate the bridge")
	}
	if _, err := run("nmcli connection up " + overlayOriginalConnection); err != nil {
		logger.Errorf("overlay-parent: activating %s failed: %v; restoring the bridge", overlayOriginalConnection, err)
		_, _ = run("nmcli connection up " + overlayBridgeConnection)
		clearOverlayMigratedMarker(run)
		return false, errors.Wrap(err, "activate the physical connection")
	}
	for i := 0; i < overlayMigrateVerifyAttempts; i++ {
		if phyHasAddressAndDefaultRoute(run, st.Phy) {
			deleteOverlayBridgeProfiles(run, st)
			logger.Infof("overlay-parent: bridge %s torn down, %s carries the host address again", overlayBridgeConnection, st.Phy)
			return true, nil
		}
		sleep(overlayMigrateVerifyInterval)
	}
	logger.Errorf("overlay-parent: %s did not obtain an address and default route; restoring the bridge", st.Phy)
	_, _ = run("nmcli connection down " + overlayOriginalConnection)
	_, _ = run("nmcli connection up " + overlayBridgeConnection)
	clearOverlayMigratedMarker(run)
	return false, fmt.Errorf("%s did not come up with an IPv4 address and default route after leaving the bridge", st.Phy)
}

// clearOverlayMigratedMarker removes the marker after a failed migration so it
// only ever means "the bridge was torn down", never "a migration was attempted".
func clearOverlayMigratedMarker(run commandRunner) {
	if _, err := run("rm -f " + overlayMigratedMarker); err != nil {
		logger.Warnf("overlay-parent: remove migration marker failed: %v", err)
	}
}

func parentDir(path string) string {
	if i := strings.LastIndex(path, "/"); i > 0 {
		return path[:i]
	}
	return "/"
}

// overlayMigratedFromBridge reports whether this upgrade tore down an active
// bridge and, if so, which NIC left it.
func overlayMigratedFromBridge(run commandRunner) (phy string, migrated bool) {
	out, err := run("cat " + overlayMigratedMarker)
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(out), true
}

// MigrateBridgeToDirect is the upgrade task that retires the legacy overlay
// bridge on this node.
type MigrateBridgeToDirect struct {
	common.KubeAction
}

func (m *MigrateBridgeToDirect) Execute(runtime connector.Runtime) error {
	migrated, err := migrateBridgeToDirect(sudoRunner(runtime), time.Sleep)
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
	phy, migrated := overlayMigratedFromBridge(run)
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
	if _, migrated := overlayMigratedFromBridge(run); !migrated {
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
func OverlayMigratedFromBridge(runtime connector.Runtime) bool {
	_, migrated := overlayMigratedFromBridge(sudoRunner(runtime))
	return migrated
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
