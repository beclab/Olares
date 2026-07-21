//go:build linux
// +build linux

package utils

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"
)

func ConnectWifi(ctx context.Context, ssid, password string) error {
	if ssid == "" {
		return errors.New("ssid is empty")
	}

	nmcli, err := findCommand(ctx, "nmcli")
	if err != nil {
		return err
	}

	args := []string{
		"d",
		"wifi",
		"connect",
		ssid,
	}

	if password != "" {
		args = append(args, "password", password)
	}

	cmd := exec.CommandContext(ctx, nmcli, args...)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	klog.Info(string(output))

	if err != nil {
		klog.Error("exec cmd error, ", err, ", nmcli", " ", strings.Join(args, " "))
		return err
	}

	if strings.Contains(string(output), "Error") {
		err = errors.New(string(output))
		return err
	}

	return nil
}

func EnableWifi(ctx context.Context) error {
	nmcli, err := findCommand(ctx, "nmcli")
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, nmcli, "r", "wifi", "on")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	klog.Info(string(output))

	if err != nil {
		klog.Error("exec cmd error, ", err, ", nmcli r wifi on")
		return err
	}

	return nil
}

func GetWifiDevice(ctx context.Context) (map[string]Device, error) {
	return deviceStatus(ctx, func(d *Device) bool { return d.Type == "wifi" }, true)
}

// managedByOthers reports whether the device is managed by another component
// (CNI, tunnels, tailscale, ...) and should be skipped by NetworkManager logic.
func managedByOthers(name string) bool {
	for _, devPrefix := range []string{"cali", "kube", "tun", "tailscale"} {
		if strings.HasPrefix(name, devPrefix) {
			return true
		}
	}
	return false
}

func GetAllDevice(ctx context.Context) (map[string]Device, error) {
	return deviceStatus(ctx, func(d *Device) bool {
		return !managedByOthers(d.Name)
	}, true)
}

// setDeviceManaged switches a single device to managed via nmcli.
func setDeviceManaged(ctx context.Context, name string) error {
	nmcli, err := findCommand(ctx, "nmcli")
	if err != nil {
		klog.Error("find nmcli error, ", err)
		return err
	}

	cmd := exec.CommandContext(ctx, nmcli, "device", "set", name, "managed", "yes")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Error("exec cmd error, ", err, ", nmcli device set ", name, " managed yes")
		return err
	}
	if strings.Contains(string(output), "Error") {
		err = errors.New(string(output))
		klog.Error("exec cmd error, ", err, ", nmcli device set ", name, " managed yes")
		return err
	}
	return nil
}

func ManagedAllDevices(ctx context.Context) (map[string]Device, error) {
	return deviceStatus(ctx, func(d *Device) bool {
		if managedByOthers(d.Name) {
			return false
		}
		if d.State == "unmanaged" {
			if err := setDeviceManaged(ctx, d.Name); err != nil {
				return false
			}
		}
		return true
	}, true)
}

// ManagedDeviceStatus enumerates network devices with a single
// `nmcli device status` call and, in the same pass, switches any unmanaged
// device to managed. It deliberately skips the per-device `nmcli device show`
// / `nmcli connection show` fan-out (i.e. it does not populate IP/gateway/DNS/
// method fields), because the frequent state-polling path only needs
// Name/Type/State/Connection. This replaces the previous back-to-back
// ManagedAllDevices + GetAllDevice calls, which together spawned ~(4 + 8M)
// bash/nmcli processes per poll (M = device count) and kept NetworkManager
// busy.
func ManagedDeviceStatus(ctx context.Context) (map[string]Device, error) {
	return deviceStatus(ctx, func(d *Device) bool {
		if managedByOthers(d.Name) {
			return false
		}
		if d.State == "unmanaged" {
			if err := setDeviceManaged(ctx, d.Name); err != nil {
				return false
			}
		}
		return true
	}, false)
}

func deviceStatus(ctx context.Context, filter func(d *Device) bool, details bool) (map[string]Device, error) {
	nmcli, err := findCommand(ctx, "nmcli")
	if err != nil {
		return nil, err
	}

	fields := []string{"DEVICE", "TYPE", "STATE", "CONNECTION"}

	cmdArgs := []string{"-g", strings.Join(fields, ",")}
	cmdArgs = append(cmdArgs, "device", "status")

	cmd := exec.CommandContext(ctx, nmcli, cmdArgs...)
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute nmcli with args %+q: %w", cmdArgs, err)
	}

	parsedOutput, err := parseCmdOutput(output, len(fields))
	if err != nil {
		return nil, fmt.Errorf("failed to parse nmcli output: %w", err)
	}

	statuss := make(map[string]Device)
	for _, fields := range parsedOutput {
		d := Device{
			Name:       fields[0],
			Type:       fields[1],
			State:      fields[2],
			Connection: fields[3],
		}

		if filter == nil || filter(&d) {
			if details {
				err = showDeviceByNM(ctx, d.Name, &d)
				if err != nil {
					klog.Error("failed to get device details for ", d.Name, ": ", err)
					continue
				}
			}

			statuss[d.Name] = d
		}
	}

	return statuss, nil
}

func parseCmdOutput(output []byte, expectedCountOfFields int) ([][]string, error) {
	lines := bytes.FieldsFunc(output, func(c rune) bool { return c == '\n' || c == '\r' })

	var recordLines [][]string
	for i, line := range lines {
		recordLine := splitBySeparator(":", string(line))
		if len(recordLine) != expectedCountOfFields {
			return nil, fmt.Errorf(
				"line %d contains %d fields but should %d",
				i, len(recordLine), expectedCountOfFields,
			)
		}

		recordLines = append(recordLines, recordLine)
	}

	return recordLines, nil
}

func splitBySeparator(separator, line string) []string {
	escape := `\`
	tempEscapedSeparator := "\x00"

	replacedEscape := strings.ReplaceAll(line, escape+separator, tempEscapedSeparator)
	records := strings.Split(replacedEscape, separator)

	for i, record := range records {
		records[i] = strings.ReplaceAll(record, tempEscapedSeparator, separator)
	}

	return records
}

// command: nmcli device show <interface>
// output format:
// GENERAL.DEVICE:                         enp3s0
// GENERAL.TYPE:                           ethernet
// GENERAL.HWADDR:                         34:5A:60:35:69:CC
// GENERAL.MTU:                            1500
// GENERAL.STATE:                          100 (connected)
// GENERAL.CONNECTION:                     Wired connection 1
// GENERAL.CON-PATH:                       /org/freedesktop/NetworkManager/ActiveConnection/1
// WIRED-PROPERTIES.CARRIER:               on
// IP4.ADDRESS[1]:                         192.168.31.145/24
// IP4.GATEWAY:                            192.168.31.1
// IP4.ROUTE[1]:                           dst = 169.254.0.0/16, nh = 0.0.0.0, mt = 1000
// IP4.ROUTE[2]:                           dst = 192.168.31.0/24, nh = 0.0.0.0, mt = 100
// IP4.ROUTE[3]:                           dst = 0.0.0.0/0, nh = 192.168.31.1, mt = 100
// IP4.DNS[1]:                             192.168.31.1
// IP6.ADDRESS[1]:                         2408:8606:1800:1::d4a/128
// IP6.ADDRESS[2]:                         2408:8606:1800:1:16c5:3fa5:ad66:f6d9/64
// IP6.ADDRESS[3]:                         2408:8606:1800:1:16f4:a31b:b33f:26f2/64
// IP6.ADDRESS[4]:                         fe80::7272:12f8:6ef6:2a42/64
// IP6.GATEWAY:                            fe80::5aea:1fff:fe64:b5dc
// IP6.ROUTE[1]:                           dst = fe80::/64, nh = ::, mt = 1024
// IP6.ROUTE[2]:                           dst = 2408:8606:1800:1::/64, nh = ::, mt = 100
// IP6.ROUTE[3]:                           dst = ::/0, nh = fe80::5aea:1fff:fe64:b5dc, mt = 20100
// IP6.ROUTE[4]:                           dst = 2408:8606:1800:1::d4a/128, nh = ::, mt = 100
// IP6.DNS[1]:                             fe80::5aea:1fff:fe64:b5dc
func showDeviceByNM(ctx context.Context, deviceName string, device *Device) error {
	nmcli, err := findCommand(ctx, "nmcli")
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, nmcli, "device", "show", deviceName)
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute nmcli: %w", err)
	}

	lines := bytes.Split(output, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		fields := bytes.SplitN(line, []byte(":"), 2)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(string(fields[0]))
		value := strings.TrimSpace(string(fields[1]))

		switch key {
		case "IP4.ADDRESS[1]":
			ipAndMask := strings.Split(value, "/")
			if len(ipAndMask) > 1 {
				device.Ipv4Address = ipAndMask[0]
				cidr, err := strconv.Atoi(ipAndMask[1])
				if err != nil {
					klog.Error("convert cidr error, ", err)
					continue
				}
				mask, err := MaskFromCIDR(cidr)
				if err != nil {
					klog.Error("get mask from cidr error, ", err)
					continue
				}
				device.Ipv4Mask = mask
			}
		case "IP4.GATEWAY":
			device.Ipv4Gateway = value
		case "IP4.DNS[1]":
			device.Ipv4DNS = value
		case "IP6.ADDRESS[1]":
			device.Ipv6Address = value
		case "IP6.GATEWAY":
			device.Ipv6Gateway = value
		case "IP6.DNS[1]":
			device.Ipv6DNS = value
		case "GENERAL.CONNECTION":
			err := showConnectionByNM(ctx, value, device)
			if err != nil {
				klog.V(8).Info("get connection method error, ", err, ", connection name: ", value)
			}
		default:
			continue
		}
	}

	return nil
}

// nmcli connection show "Wired connection 1"
func showConnectionByNM(ctx context.Context, connectionName string, device *Device) error {
	nmcli, err := findCommand(ctx, "nmcli")
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, nmcli, "connection", "show", connectionName)
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute nmcli: %w", err)
	}

	lines := bytes.Split(output, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		fields := bytes.SplitN(line, []byte(":"), 2)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(string(fields[0]))
		value := strings.TrimSpace(string(fields[1]))

		switch key {
		case "ipv4.method":
			device.Method = value
		}
	}
	return nil
}

type NetworkTraffic struct {
	Interface string
	RxBytes   uint64
	TxBytes   uint64
}

func getInterfaceTraffic() (traffic map[string]*NetworkTraffic, err error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	traffic = make(map[string]*NetworkTraffic)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			return nil, fmt.Errorf("unexpected format for interface %s", parts[0])
		}
		rxBytes, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			return nil, err
		}
		txBytes, err := strconv.ParseUint(fields[8], 10, 64)
		if err != nil {
			return nil, err
		}

		traffic[parts[0]] = &NetworkTraffic{
			Interface: parts[0],
			RxBytes:   rxBytes,
			TxBytes:   txBytes,
		}
	}

	return traffic, nil
}

type NetworkTrafficRate struct {
	Interface  string
	RxBytes    uint64
	TxBytes    uint64
	RxRate     float64
	TxRate     float64
	UpdateTime time.Time
}

var AllNetworkDeviceTraffic = make(map[string]*NetworkTrafficRate)

func UpdateNetworkTraffic(ctx context.Context) {
	traffic, err := getInterfaceTraffic()
	if err != nil {
		klog.Error("get interface traffic error, ", err)
		return
	}

	for name, netTraffic := range traffic {
		rate, ok := AllNetworkDeviceTraffic[name]
		if !ok {
			AllNetworkDeviceTraffic[name] = &NetworkTrafficRate{
				Interface:  name,
				RxBytes:    netTraffic.RxBytes,
				TxBytes:    netTraffic.TxBytes,
				UpdateTime: time.Now(),
			}
			continue
		}

		rate.RxRate = float64(netTraffic.RxBytes-rate.RxBytes) / time.Since(rate.UpdateTime).Seconds()
		rate.TxRate = float64(netTraffic.TxBytes-rate.TxBytes) / time.Since(rate.UpdateTime).Seconds()
		rate.RxBytes = netTraffic.RxBytes
		rate.TxBytes = netTraffic.TxBytes
		rate.UpdateTime = time.Now()
	}
}

func GetInterfaceTraffic(iface string) (rxBytes, txBytes float64, err error) {
	rates, ok := AllNetworkDeviceTraffic[iface]
	if !ok {
		return 0, 0, fmt.Errorf("interface %s not found", iface)
	}
	return rates.RxRate, rates.TxRate, nil
}

// get ethernet connection
func GetEthernetConnection(ctx context.Context) (iface, ifUUID, connection string, err error) {
	var (
		iefs []net.Interface
	)
	if iefs, err = net.Interfaces(); err != nil {
		klog.Error("list network interfaces error, ", err)
		return "", "", "", err
	}

	// get k8s master node ip
	masterIp, err := MasterNodeIp(true)
	if err != nil {
		klog.Error("get master node ip error, ", err)
		// FIXME: if k8s is not installed or not ready, we should return error
	}

	var firstEthernetInterface struct {
		Name       string
		UUID       string
		Connection string
	}
	for _, ief := range iefs {
		if ief.Flags&net.FlagUp == 0 {
			continue
		}

		nmcli, err := findCommand(ctx, "nmcli")
		if err != nil {
			klog.Error("find nmcli command error, ", err)
			return "", "", "", err
		}
		cmd := exec.CommandContext(ctx, nmcli, "-g", "GENERAL.TYPE,IP4.ADDRESS,GENERAL.CON-UUID,GENERAL.CONNECTION", "device", "show", ief.Name)
		cmd.Env = os.Environ()
		output, err := cmd.Output()
		if err != nil {
			klog.Error("failed to execute nmcli: %w", err)
			return "", "", "", err
		}

		lines := bytes.Split(output, []byte("\n"))
		if len(lines) < 4 {
			klog.Errorf("unexpected output from nmcli: %s", string(output))
			return "", "", "", errors.New(fmt.Sprintf("unexpected output from nmcli: %s", string(output)))
		}

		ifType := strings.TrimSpace(string(lines[0]))
		ipv4 := strings.TrimSpace(string(lines[1])) // got 192.168.31.145/24
		ipv4 = strings.Split(ipv4, "/")[0]
		ifUUID := strings.TrimSpace(string(lines[2]))
		connection := strings.TrimSpace(string(lines[3]))
		if ipv4 == "" {
			continue
		}

		if connection == bridgeConnectionName ||
			strings.HasPrefix(connection, bridgeSlavePrefix) {
			// bridge connection
			continue
		}

		// active connection
		switch ifType {
		case "ethernet":
			if masterIp != "" && masterIp == ipv4 {
				// use the ethernet interface of the master node binding in priority
				return ief.Name, ifUUID, connection, nil
			}
			if firstEthernetInterface.Name == "" {
				firstEthernetInterface.Name = ief.Name
				firstEthernetInterface.UUID = ifUUID
				firstEthernetInterface.Connection = connection
			}
		default:
			continue
		}
	}

	if firstEthernetInterface.Name != "" {
		return firstEthernetInterface.Name, firstEthernetInterface.UUID, firstEthernetInterface.Connection, nil
	}

	klog.Error("no ethernet connection found")
	return "", "", "", errors.New("no ethernet connection found")
}

const (
	bridgeConnectionName   = "br-olares"
	bridgeSlavePrefix      = "br-olares-slave-"
	originalConnectionName = "original-connection"
)

// bridgeReadyTimeout bounds how long CreateBridgeConnection waits for the bridge
// to obtain an IPv4 lease before treating the switch as failed and rolling back.
// 90s covers slow home/office DHCP renewals and switch MAC learning while still
// failing closed if a competing profile steals the NIC (no IPv4 will ever appear).
const bridgeReadyTimeout = 90 * time.Second

// checkpointDestroyRetries / checkpointDestroyRetryDelay bound how hard we try
// to commit (CheckpointDestroy) after the bridge is already ready. A failed
// destroy must not be treated as success: NetworkManager would still auto-roll
// back when the checkpoint timeout expires ("UI on, then silently off").
const (
	checkpointDestroyRetries    = 3
	checkpointDestroyRetryDelay = 200 * time.Millisecond
)

// Injectable for unit tests; production defaults to the real D-Bus helpers.
var (
	nmCheckpointDestroyFn  = nmCheckpointDestroy
	nmCheckpointRollbackFn = nmCheckpointRollback
)

// nmcliCombinedOutput runs nmcli (or any argv[0]) and returns combined stdout.
// Tests replace this to assert argv without talking to NetworkManager.
var nmcliCombinedOutput = func(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = os.Environ()
	return cmd.Output()
}

// bridgeAddArgs builds the nmcli argv for creating br-olares with PHY MAC clone
// and DHCP on the bridge (INV-MAC-01).
func bridgeAddArgs(mac string) []string {
	return []string{
		"connection", "add", "type", "bridge", "con-name", bridgeConnectionName, "ifname", bridgeConnectionName,
		"connection.autoconnect", "yes", "bridge.stp", "no",
		"ethernet.cloned-mac-address", mac,
		"ipv4.method", "auto",
		"ipv6.method", "ignore",
	}
}

// classifyBridgeResetUUIDs partitions NM profiles into br-olares / slave /
// original-connection UUID lists used by ResetBridgeConnection. Unrelated
// bridges (e.g. docker0, br0) are ignored.
func classifyBridgeResetUUIDs(conns []nmConnection) (bridgeUUIDs, slaveUUIDs, originalUUIDs []string) {
	for _, c := range conns {
		switch {
		case c.Name == bridgeConnectionName:
			bridgeUUIDs = append(bridgeUUIDs, c.UUID)
		case strings.HasPrefix(c.Name, bridgeSlavePrefix):
			slaveUUIDs = append(slaveUUIDs, c.UUID)
		case c.Name == originalConnectionName:
			originalUUIDs = append(originalUUIDs, c.UUID)
		}
	}
	return bridgeUUIDs, slaveUUIDs, originalUUIDs
}

// activateBridgeSwitch enables the slave, downs the physical profile, then ups
// the bridge — three discrete nmcli invocations (no sh -c).
func activateBridgeSwitch(ctx context.Context, nmcli, slaveUUID, ifUUID, bridgeUUID string) error {
	if _, err := nmcliCombinedOutput(ctx, nmcli, "connection", "modify", "uuid", slaveUUID, "connection.autoconnect", "yes"); err != nil {
		klog.Errorf("overlay bridge activate: enable autoconnect on slave %s failed: %v", slaveUUID, err)
		return fmt.Errorf("failed to enable autoconnect on bridge slave %s: %w", slaveUUID, err)
	}
	if _, err := nmcliCombinedOutput(ctx, nmcli, "connection", "down", "uuid", ifUUID); err != nil {
		klog.Errorf("overlay bridge activate: down original connection %s failed: %v", ifUUID, err)
		return fmt.Errorf("failed to down original connection %s: %w", ifUUID, err)
	}
	if _, err := nmcliCombinedOutput(ctx, nmcli, "connection", "up", "uuid", bridgeUUID); err != nil {
		klog.Errorf("overlay bridge activate: up bridge connection %s failed: %v", bridgeUUID, err)
		return fmt.Errorf("failed to up bridge connection %s: %w", bridgeUUID, err)
	}
	return nil
}

// nmConnection is a minimal view of a NetworkManager connection profile.
type nmConnection struct {
	UUID string
	Type string
	Name string
}

// parseNMConnections parses the terse output of
// `nmcli -t -f UUID,TYPE,NAME connection show`. UUID and TYPE never contain ':',
// so the line can be split into three fields; NetworkManager escapes ':' inside
// the NAME field as '\:'.
func parseNMConnections(output []byte) []nmConnection {
	var conns []nmConnection
	lines := bytes.Split(output, []byte("\n"))
	for _, line := range lines {
		s := strings.TrimSpace(string(line))
		if s == "" {
			continue
		}
		fields := strings.SplitN(s, ":", 3)
		if len(fields) < 3 {
			continue
		}
		conns = append(conns, nmConnection{
			UUID: strings.TrimSpace(fields[0]),
			Type: strings.TrimSpace(fields[1]),
			Name: strings.ReplaceAll(strings.TrimSpace(fields[2]), "\\:", ":"),
		})
	}
	return conns
}

// listNMConnections returns all NetworkManager connection profiles.
func listNMConnections(ctx context.Context, nmcli string) ([]nmConnection, error) {
	cmd := exec.CommandContext(ctx, nmcli, "-t", "-f", "UUID,TYPE,NAME", "connection", "show")
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseNMConnections(output), nil
}

// deleteNMConnection removes a connection profile by UUID (best-effort).
func deleteNMConnection(ctx context.Context, nmcli, uuid string) {
	cmd := exec.CommandContext(ctx, nmcli, "connection", "delete", "uuid", uuid)
	cmd.Env = os.Environ()
	if _, err := cmd.Output(); err != nil {
		klog.Errorf("failed to delete connection %s: %v", uuid, err)
	}
}

// setNMAutoconnect toggles connection.autoconnect for a profile by UUID (best-effort).
func setNMAutoconnect(ctx context.Context, nmcli, uuid string, enabled bool) {
	value := "no"
	if enabled {
		value = "yes"
	}
	cmd := exec.CommandContext(ctx, nmcli, "connection", "modify", "uuid", uuid, "connection.autoconnect", value)
	cmd.Env = os.Environ()
	if _, err := cmd.Output(); err != nil {
		klog.Errorf("failed to set autoconnect=%s for connection %s: %v", value, uuid, err)
	}
}

// cleanupBridgeConnections deletes every leftover br-olares / br-olares-slave-*
// profile by UUID. It is idempotent and prevents stale profiles from accumulating
// across repeated enable/disable cycles.
func cleanupBridgeConnections(ctx context.Context, nmcli string) {
	conns, err := listNMConnections(ctx, nmcli)
	if err != nil {
		klog.Errorf("list connections error: %v", err)
		return
	}
	for _, c := range conns {
		if c.Name == bridgeConnectionName || strings.HasPrefix(c.Name, bridgeSlavePrefix) {
			klog.Infof("clear bridge connection [%s] (%s)", c.Name, c.UUID)
			deleteNMConnection(ctx, nmcli, c.UUID)
		}
	}
}

// removeStaleOriginalConnections deletes duplicate "original-connection" backup
// profiles, keeping only keepUUID. Duplicate backups (same name, different UUID)
// are the root cause of the auto-activation race that steals the physical NIC from
// the bridge slave, so they must be pruned before every switch.
func removeStaleOriginalConnections(ctx context.Context, nmcli, keepUUID string) {
	conns, err := listNMConnections(ctx, nmcli)
	if err != nil {
		klog.Errorf("list connections error: %v", err)
		return
	}
	for _, c := range conns {
		if c.Name == originalConnectionName && c.UUID != keepUUID {
			klog.Infof("remove stale original connection (%s)", c.UUID)
			deleteNMConnection(ctx, nmcli, c.UUID)
		}
	}
}

// connectionUUIDByName resolves a connection profile UUID from its name.
func connectionUUIDByName(ctx context.Context, nmcli, name string) (string, error) {
	cmd := exec.CommandContext(ctx, nmcli, "-g", "connection.uuid", "connection", "show", name)
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		klog.Errorf("overlay bridge: resolve uuid for connection %s failed: %v", name, err)
		return "", err
	}
	uuid := strings.TrimSpace(string(output))
	if uuid == "" {
		klog.Errorf("overlay bridge: connection %s has empty uuid", name)
		return "", fmt.Errorf("connection %s has no uuid", name)
	}
	return uuid, nil
}

// waitBridgeReady blocks until br-olares is active with an IPv4 address, or timeout.
func waitBridgeReady(ctx context.Context, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		if c, err := FindBridgeConnection(ctx); err == nil && c != nil && c.Active && c.Ipv4Address != "" {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(2 * time.Second):
		}
	}
}

// commitNMCheckpoint discards the checkpoint after a successful bridge switch
// (NM "commit"). On persistent Destroy failure it rolls the network back and
// returns an error so callers never report success while NM still holds an
// auto-rollback timer.
func commitNMCheckpoint(ctx context.Context, conn *dbus.Conn, checkpoint dbus.ObjectPath, manualRollback func()) error {
	if checkpoint == "" {
		return nil
	}
	var lastErr error
	for attempt := 1; attempt <= checkpointDestroyRetries; attempt++ {
		if err := nmCheckpointDestroyFn(ctx, conn, checkpoint); err != nil {
			lastErr = err
			klog.Errorf("NM checkpoint destroy failed (attempt %d/%d): %v", attempt, checkpointDestroyRetries, err)
			if attempt < checkpointDestroyRetries {
				select {
				case <-ctx.Done():
					lastErr = ctx.Err()
					klog.Errorf("NM checkpoint destroy aborted by context: %v", lastErr)
					goto rollback
				case <-time.After(checkpointDestroyRetryDelay):
				}
				continue
			}
			break
		}
		return nil
	}

rollback:
	klog.Errorf("NM checkpoint commit failed after retries, rolling back bridged state: %v", lastErr)
	if e := nmCheckpointRollbackFn(ctx, conn, checkpoint); e != nil {
		klog.Errorf("NM checkpoint rollback failed (%v), applying manual rollback", e)
		if manualRollback != nil {
			manualRollback()
		}
	}
	return fmt.Errorf("checkpoint commit failed: %w", lastErr)
}

// ResetBridgeConnection tears down the overlay bridge and restores the original
// physical connection. It operates by UUID so that duplicate/leftover profiles are
// fully cleaned up, leaving no residual br-olares* profiles behind.
func ResetBridgeConnection(ctx context.Context) error {
	nmcli, err := findCommand(ctx, "nmcli")
	if err != nil {
		klog.Errorf("overlay bridge reset: find nmcli failed: %v", err)
		return err
	}

	conns, err := listNMConnections(ctx, nmcli)
	if err != nil {
		// Without a connection list we cannot safely identify br-olares /
		// original-connection UUIDs. Fail closed so disable does not proceed to
		// stop cni-dhcp / clear app settings while the bridge may still be up.
		klog.Errorf("overlay bridge reset: list connections error: %v", err)
		return fmt.Errorf("list NM connections for bridge reset: %w", err)
	}

	var bridgeUUIDs, slaveUUIDs, originalUUIDs []string
	bridgeUUIDs, slaveUUIDs, originalUUIDs = classifyBridgeResetUUIDs(conns)
	if len(slaveUUIDs) > 1 || len(bridgeUUIDs) > 1 {
		klog.Warningf("unexpected leftover bridge profiles: bridges=%d slaves=%d", len(bridgeUUIDs), len(slaveUUIDs))
	}

	// shut down the bridge connection(s)
	for _, u := range bridgeUUIDs {
		klog.Infof("turn off the bridge connection [%s]", u)
		cmd := exec.CommandContext(ctx, nmcli, "connection", "down", "uuid", u)
		cmd.Env = os.Environ()
		if _, err := cmd.Output(); err != nil {
			klog.Errorf("failed to turn off bridge %s: %v", u, err)
		}
	}

	// restore the original physical connection(s)
	for _, u := range originalUUIDs {
		klog.Infof("turn on the original connection [%s]", u)
		setNMAutoconnect(ctx, nmcli, u, true)
		cmd := exec.CommandContext(ctx, nmcli, "connection", "up", "uuid", u)
		cmd.Env = os.Environ()
		if _, err := cmd.Output(); err != nil {
			klog.Errorf("failed to restore original connection %s: %v", u, err)
		}
	}

	// delete slave then bridge profiles by UUID
	for _, u := range slaveUUIDs {
		klog.Infof("delete the bridge slave connection [%s]", u)
		deleteNMConnection(ctx, nmcli, u)
	}
	for _, u := range bridgeUUIDs {
		klog.Infof("delete the bridge connection [%s]", u)
		deleteNMConnection(ctx, nmcli, u)
	}

	return nil
}

// CreateBridgeConnection atomically switches the node's primary networking from
// the physical ethernet interface onto the br-olares bridge. It either fully
// succeeds (bridge active with an IPv4 address) or rolls back to the original
// physical connection and returns an error; it never leaves a half-bridged state
// or leftover/duplicate NetworkManager profiles behind.
func CreateBridgeConnection(ctx context.Context) error {
	nmcli, err := findCommand(ctx, "nmcli")
	if err != nil {
		klog.Errorf("overlay bridge create: find nmcli failed: %v", err)
		return err
	}

	// snapshot the physical ethernet connection currently carrying the node IP
	iface, ifUUID, connName, err := GetEthernetConnection(ctx)
	if err != nil {
		klog.Error("get ethernet connection error, ", err)
		return err
	}

	// read the physical MAC up front so we can abort before mutating anything
	mdata, err := os.ReadFile(fmt.Sprintf("/sys/class/net/%s/address", iface))
	if err != nil {
		klog.Error("read interface MAC address error, ", err)
		return err
	}
	mac := strings.TrimSpace(string(mdata))
	if mac == "" {
		klog.Errorf("overlay bridge create: interface %s MAC address is empty", iface)
		return fmt.Errorf("interface %s MAC address is empty", iface)
	}

	// pre-clean leftover bridge profiles and duplicate backups by UUID (idempotent)
	cleanupBridgeConnections(ctx, nmcli)
	removeStaleOriginalConnections(ctx, nmcli, ifUUID)

	// back up the original connection in place (UUID is stable across the rename);
	// skip when it is already backed up to avoid a rename-to-same-name error.
	if connName != originalConnectionName {
		klog.Infof("backup the original connection [%s] to [%s]", ifUUID, originalConnectionName)
		cmd := exec.CommandContext(ctx, nmcli, "connection", "modify", "uuid", ifUUID, "con-name", originalConnectionName)
		cmd.Env = os.Environ()
		if _, err = cmd.Output(); err != nil {
			klog.Errorf("overlay bridge create: backup original connection %s failed: %v", ifUUID, err)
			return fmt.Errorf("failed to backup original connection: %w", err)
		}
	}

	// Crash-safe rollback net: wrap the switch in a NetworkManager checkpoint so
	// that even if olaresd is killed/panics mid-switch, NetworkManager itself
	// restores the original network after the rollback timeout. Best-effort: on
	// hosts where the checkpoint API is unavailable we fall back to the manual
	// (in-process) rollback below.
	var (
		dbusConn   *dbus.Conn
		checkpoint dbus.ObjectPath
	)
	if c, e := nmSystemBus(); e != nil {
		klog.Warningf("connect system bus for NM checkpoint failed, using manual rollback only: %v", e)
	} else {
		dbusConn = c
		defer dbusConn.Close()
		// timeout must outlast the whole switch (add/slave/up + bridge verify).
		if cp, e := nmCheckpointCreate(ctx, dbusConn, bridgeReadyTimeout+2*time.Minute); e != nil {
			klog.Warningf("NM checkpoint create failed, using manual rollback only: %v", e)
		} else {
			checkpoint = cp
			klog.Infof("created NM checkpoint %s as crash-safe rollback net", cp)
		}
	}

	// transaction state used by the manual (fallback) rollback
	var (
		bridgeUUID    string
		slaveUUID     string
		disabledUUIDs []string
	)
	manualRollback := func() {
		if slaveUUID != "" {
			deleteNMConnection(ctx, nmcli, slaveUUID)
		}
		if bridgeUUID != "" {
			deleteNMConnection(ctx, nmcli, bridgeUUID)
		}
		// re-enable autoconnect for every profile we disabled during the switch
		for _, u := range disabledUUIDs {
			setNMAutoconnect(ctx, nmcli, u, true)
		}
		// bring the original physical connection back online
		setNMAutoconnect(ctx, nmcli, ifUUID, true)
		up := exec.CommandContext(ctx, nmcli, "connection", "up", "uuid", ifUUID)
		up.Env = os.Environ()
		if _, e := up.Output(); e != nil {
			klog.Errorf("failed to restore original connection %s: %v", ifUUID, e)
		}
	}
	// rollback prefers the NM checkpoint (atomic, covers all touched devices and
	// connections); it degrades to the manual rollback if no checkpoint exists
	// or the checkpoint rollback itself fails.
	rollback := func(cause error) error {
		klog.Errorf("create bridge connection failed, rolling back: %v", cause)
		if checkpoint != "" {
			if e := nmCheckpointRollbackFn(ctx, dbusConn, checkpoint); e != nil {
				klog.Errorf("NM checkpoint rollback failed (%v), applying manual rollback", e)
				manualRollback()
			}
			return cause
		}
		manualRollback()
		return cause
	}

	// create the bridge connection; the cloned MAC preserves the DHCP lease/IP
	klog.Infof("create bridge connection [%s]", bridgeConnectionName)
	if _, err = nmcliCombinedOutput(ctx, nmcli, bridgeAddArgs(mac)...); err != nil {
		return rollback(fmt.Errorf("failed to create bridge connection: %w", err))
	}
	if bridgeUUID, err = connectionUUIDByName(ctx, nmcli, bridgeConnectionName); err != nil {
		return rollback(fmt.Errorf("failed to resolve bridge uuid: %w", err))
	}

	// create the bridge slave connection on the physical interface
	slaveConnectionName := fmt.Sprintf("%s%s", bridgeSlavePrefix, iface)
	klog.Infof("create bridge slave connection [%s]", slaveConnectionName)
	if _, err = nmcliCombinedOutput(ctx, nmcli,
		"connection", "add", "type", "ethernet", "con-name", slaveConnectionName, "ifname", iface,
		"master", bridgeConnectionName, "slave-type", "bridge", "connection.autoconnect", "yes",
	); err != nil {
		return rollback(fmt.Errorf("failed to create bridge slave connection: %w", err))
	}
	if slaveUUID, err = connectionUUIDByName(ctx, nmcli, slaveConnectionName); err != nil {
		return rollback(fmt.Errorf("failed to resolve bridge slave uuid: %w", err))
	}

	// disable autoconnect for every other ethernet profile BY UUID so that no
	// competing/duplicate profile can auto-activate and steal the physical NIC
	// from the bridge slave during the switch (root cause of the enable timeout).
	conns, err := listNMConnections(ctx, nmcli)
	if err != nil {
		return rollback(fmt.Errorf("failed to list connections: %w", err))
	}
	for _, c := range conns {
		if c.UUID == slaveUUID || !strings.Contains(c.Type, "ethernet") {
			continue
		}
		klog.Infof("disable auto connect for connection [%s] (%s)", c.Name, c.UUID)
		setNMAutoconnect(ctx, nmcli, c.UUID, false)
		disabledUUIDs = append(disabledUUIDs, c.UUID)
	}

	// Switch the physical NIC onto the bridge with three separate nmcli shots
	// (no shell). Failures are attributed per step; down failure skips up.
	klog.Infof("turn on the bridge connection [%s]", bridgeConnectionName)
	if err := activateBridgeSwitch(ctx, nmcli, slaveUUID, ifUUID, bridgeUUID); err != nil {
		return rollback(err)
	}

	// verify the bridge really came up with an IPv4 address; otherwise roll back
	if !waitBridgeReady(ctx, bridgeReadyTimeout) {
		return rollback(fmt.Errorf("bridge %s did not become active with an IPv4 address within %s", bridgeConnectionName, bridgeReadyTimeout))
	}

	// commit: discard the checkpoint so NetworkManager keeps the bridged state.
	// Destroy failure must not succeed the call — NM would still auto-roll back.
	if err := commitNMCheckpoint(ctx, dbusConn, checkpoint, manualRollback); err != nil {
		return err
	}

	klog.Infof("bridge connection [%s] is active", bridgeConnectionName)
	return nil
}

func CheckOverlayGatewayStatus(ctx context.Context) error {
	// TODO: implement check  overlay gateway enabling status
	/*
		bridge link
		ip -4 addr show dev br-olares
	*/
	return nil
}

func FindBridgeConnection(ctx context.Context) (*BridgeConnection, error) {
	nmcli, err := findCommand(ctx, "nmcli")
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, nmcli, "-g", "name,type,active", "connection", "show")
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(output, []byte("\n"))
	if len(lines) < 1 {
		return nil, nil
	}

	var slaveConnectionName string
	bridgeConnection := &BridgeConnection{}
	bridgeConnectionActive := true
	for _, line := range lines {
		fields := bytes.Split(line, []byte(":"))
		if len(fields) < 3 {
			continue
		}
		name := strings.TrimSpace(string(fields[0]))
		ctype := strings.TrimSpace(string(fields[1]))
		active := strings.TrimSpace(string(fields[2]))

		switch {
		case name == bridgeConnectionName:
			bridgeConnection.BridgeName = name
			if active == "no" {
				bridgeConnectionActive = false
			}

			// check the bridge connection binding an IP address
			cmd = exec.CommandContext(ctx, nmcli, "-g", "IP4.ADDRESS", "connection", "show", name)
			cmd.Env = os.Environ()
			output, err := cmd.Output()
			if err != nil {
				return nil, err
			}

			lines := bytes.Split(output, []byte("\n"))
			if len(lines) < 1 {
				bridgeConnectionActive = false
				continue
			}

			ipv4 := strings.TrimSpace(string(lines[0]))
			if ipv4 == "" {
				bridgeConnectionActive = false
				continue
			}
			ipv4 = strings.Split(ipv4, "/")[0]
			bridgeConnection.Ipv4Address = ipv4
		case strings.Contains(ctype, "ethernet") && strings.HasPrefix(name, bridgeSlavePrefix):
			// should be only one slave connection
			slaveConnectionName = name
			if active == "no" {
				bridgeConnectionActive = false
			}
		default:
		}
	}

	if bridgeConnection.BridgeName != "" {
		// found the bridge connection
		bridgeConnection.SlaveName = slaveConnectionName
		bridgeConnection.Active = bridgeConnectionActive
		return bridgeConnection, nil
	}

	return nil, nil
}

func ListenNetworkCarrierChanges(ctx context.Context, downCallback func()) error {
	updates := make(chan netlink.LinkUpdate)

	if err := netlink.LinkSubscribe(updates, ctx.Done()); err != nil {
		klog.Error("subscribe network changes error, ", err)
		return err
	}

	for {
		select {
		case update := <-updates:
			if update.Attrs().Name == bridgeConnectionName {
				isLowerUp := (update.Flags & 0x10000) != 0
				isUp := (update.Flags & 1) != 0

				if !isLowerUp || !isUp {
					klog.Infof("network change detected: %s, state: %s", update.Attrs().Name, update.Link.Attrs().OperState.String())
					downCallback()
				}
			}
		case <-ctx.Done():
			klog.Info("stop listening network changes")
			return nil
		}
	}
}
