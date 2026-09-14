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
	// OverlayLinkFile persists the alternative name across reboots through
	// systemd-udevd, independently of olaresd.
	OverlayLinkFile = "/etc/systemd/network/10-olares-lan.link"
	// overlayMigratedMarker records that the bridge was active before the
	// upgrade tore it down, so the later upgrade steps know they must write the
	// desired state and recreate the overlay Pods even when re-run.
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

// nmDevice is one row of `nmcli -t -f DEVICE,TYPE,STATE,CONNECTION device`.
type nmDevice struct {
	Name       string
	Type       string
	State      string
	Connection string
}

func parseNMDevices(out string) []nmDevice {
	var devs []nmDevice
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, ":", 4)
		if len(fields) < 3 {
			continue
		}
		d := nmDevice{Name: fields[0], Type: fields[1], State: fields[2]}
		if len(fields) == 4 {
			d.Connection = fields[3]
		}
		devs = append(devs, d)
	}
	return devs
}

// selectWiredInterface picks the wired NIC that carries the LAN: an ethernet
// device NetworkManager reports as connected and that is not enslaved to the
// legacy overlay bridge. The default-route interface is deliberately not used
// as a fallback: a host with Wi-Fi as a second default route would pick the
// wrong device.
func selectWiredInterface(devs []nmDevice) (string, bool) {
	for _, d := range devs {
		if d.Type != "ethernet" || !strings.HasPrefix(d.State, "connected") {
			continue
		}
		if strings.HasPrefix(d.Connection, overlayBridgeSlavePrefix) {
			continue
		}
		return d.Name, true
	}
	return "", false
}

// overlayLinkFileContent renders the systemd.link unit that re-adds the
// alternative name whenever the NIC appears. Matching on MAC plus type keeps
// the rule stable across kernel interface renames.
func overlayLinkFileContent(mac string) string {
	return fmt.Sprintf("[Match]\nMACAddress=%s\nType=ether\n\n[Link]\nAlternativeName=%s\n", mac, OverlayParentAltname)
}

func overlayLinkFileCommand(mac string) string {
	return fmt.Sprintf("mkdir -p /etc/systemd/network && printf '%%s\\n' '[Match]' 'MACAddress=%s' 'Type=ether' '' '[Link]' 'AlternativeName=%s' > %s",
		mac, OverlayParentAltname, OverlayLinkFile)
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

// ensureOverlayAltname adds the alternative name to the selected wired NIC and
// persists it. It refuses to move the name to a different device: a silent
// re-bind would move every overlay Pod to another network.
func ensureOverlayAltname(run commandRunner) (string, error) {
	out, err := run("nmcli -t -f DEVICE,TYPE,STATE,CONNECTION device")
	if err != nil {
		return "", errors.Wrap(err, "list network devices")
	}
	dev, ok := selectWiredInterface(parseNMDevices(out))
	if !ok {
		logger.Warnf("overlay-parent: no connected wired interface, alternative name %s not added", OverlayParentAltname)
		return "", nil
	}
	if bound := resolveOverlayAltname(run); bound != "" && bound != dev {
		return "", fmt.Errorf("alternative name %s is bound to %s, expected %s", OverlayParentAltname, bound, dev)
	} else if bound == "" {
		if _, err := run(fmt.Sprintf("ip link property add dev %s altname %s", dev, OverlayParentAltname)); err != nil {
			return "", errors.Wrapf(err, "add alternative name %s to %s", OverlayParentAltname, dev)
		}
	}
	mac, err := run("cat /sys/class/net/" + dev + "/address")
	if err != nil {
		return "", errors.Wrapf(err, "read MAC of %s", dev)
	}
	mac = strings.TrimSpace(mac)
	if _, err := run(overlayLinkFileCommand(mac)); err != nil {
		return "", errors.Wrap(err, "write systemd link file")
	}
	logger.Infof("overlay-parent: %s carries alternative name %s (mac %s)", dev, OverlayParentAltname, mac)
	return dev, nil
}

func removeOverlayAltname(run commandRunner) {
	_, _ = run("rm -f " + OverlayLinkFile)
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
	if _, err := run(fmt.Sprintf("mkdir -p %s && touch %s", parentDir(overlayMigratedMarker), overlayMigratedMarker)); err != nil {
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

func overlayMigratedFromBridge(run commandRunner) bool {
	_, err := run("test -f " + overlayMigratedMarker)
	return err == nil
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

// EnsureOverlayAltname adds the alternative name to the wired NIC and persists it.
type EnsureOverlayAltname struct {
	common.KubeAction
}

func (e *EnsureOverlayAltname) Execute(runtime connector.Runtime) error {
	_, err := ensureOverlayAltname(sudoRunner(runtime))
	return err
}

// WriteOverlayDesiredIfMigrated records the overlay gateway as enabled when the
// upgrade tore down an active bridge, so the node ends the upgrade in the same
// user-visible state it started in.
type WriteOverlayDesiredIfMigrated struct {
	common.KubeAction
}

func (w *WriteOverlayDesiredIfMigrated) Execute(runtime connector.Runtime) error {
	run := sudoRunner(runtime)
	if !overlayMigratedFromBridge(run) {
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
	return overlayMigratedFromBridge(sudoRunner(runtime))
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
