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
	return deviceStatus(ctx, func(d *Device) bool { return d.Type == "wifi" })
}

func GetAllDevice(ctx context.Context) (map[string]Device, error) {
	return deviceStatus(ctx, func(d *Device) bool {
		managedByOthers := []string{"cali", "kube", "tun", "tailscale"}
		for _, devPrefix := range managedByOthers {
			if strings.HasPrefix(d.Name, devPrefix) {
				return false
			}
		}

		return true
	})
}

func ManagedAllDevices(ctx context.Context) (map[string]Device, error) {
	return deviceStatus(ctx, func(d *Device) bool {
		managedByOthers := []string{"cali", "kube", "tun", "tailscale"}
		for _, devPrefix := range managedByOthers {
			if strings.HasPrefix(d.Name, devPrefix) {
				return false
			}
		}
		if d.State == "unmanaged" {
			nmcli, err := findCommand(ctx, "nmcli")
			if err != nil {
				klog.Error("find nmcli error, ", err)
				return false
			}

			cmd := exec.CommandContext(ctx, nmcli, "device", "set", d.Name, "managed", "yes")
			cmd.Env = os.Environ()
			output, err := cmd.CombinedOutput()
			if err != nil {
				klog.Error("exec cmd error, ", err, ", nmcli device set ", d.Name, " managed yes")
				return false
			}
			if strings.Contains(string(output), "Error") {
				err = errors.New(string(output))
				klog.Error("exec cmd error, ", err, ", nmcli device set ", d.Name, " managed yes")
				return false
			}
		}
		return true
	})
}

func deviceStatus(ctx context.Context, filter func(d *Device) bool) (map[string]Device, error) {
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
			err = showDeviceByNM(ctx, d.Name, &d)
			if err != nil {
				klog.Error("failed to get device details for ", d.Name, ": ", err)
				continue
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

func ResetBridgeConnection(ctx context.Context) error {
	nmcli, err := findCommand(ctx, "nmcli")
	if err != nil {
		return err
	}

	// find the bridge slave connections
	// it should be only one bridge slave connection
	var slaveConnections []string
	cmd := exec.CommandContext(ctx, nmcli, "-g", "NAME", "connection", "show")
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		klog.Error("failed to execute nmcli: %w", err)
	} else {
		lines := bytes.Split(output, []byte("\n"))
		for _, line := range lines {
			cname := strings.TrimSpace(string(line))
			if cname == bridgeConnectionName {
				continue
			}

			if strings.HasPrefix(cname, bridgeSlavePrefix) {
				slaveConnections = append(slaveConnections, cname)
			}
		}
	}
	if len(slaveConnections) > 1 {
		klog.Warningf("unexpected number of bridge slave connections: %d", len(slaveConnections))
	}

	// shutdown the bridge connection
	cmd = exec.CommandContext(ctx, nmcli, "connection", "down", bridgeConnectionName)
	cmd.Env = os.Environ()
	_, err = cmd.Output()
	if err != nil {
		klog.Error("failed to execute nmcli: %w", err)
	}

	// turn on the original connection
	cmd = exec.CommandContext(ctx, nmcli, "connection", "modify", originalConnectionName, "connection.autoconnect", "yes")
	cmd.Env = os.Environ()
	_, err = cmd.Output()
	if err != nil {
		klog.Error("failed to execute nmcli: %w", err)
	}

	cmd = exec.CommandContext(ctx, nmcli, "connection", "up", originalConnectionName)
	cmd.Env = os.Environ()
	_, err = cmd.Output()
	if err != nil {
		klog.Error("failed to execute nmcli: %w", err)
	}

	// delete the bridge slave connections
	for _, cname := range slaveConnections {
		cmd = exec.CommandContext(ctx, nmcli, "connection", "delete", cname)
		cmd.Env = os.Environ()
		_, err = cmd.Output()
		if err != nil {
			klog.Error("failed to execute nmcli: %w", err)
		}
	}

	// delete the bridge connection
	cmd = exec.CommandContext(ctx, nmcli, "connection", "delete", bridgeConnectionName)
	cmd.Env = os.Environ()
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute nmcli: %w", err)
	}

	return nil
}

func CreateBridgeConnection(ctx context.Context) error {
	nmcli, err := findCommand(ctx, "nmcli")
	if err != nil {
		return err
	}

	// backup the original connection
	iface, ifUUID, _, err := GetEthernetConnection(ctx)
	if err != nil {
		klog.Error("get ethernet connection error, ", err)
		return err
	}

	// modify the original connection
	cmd := exec.CommandContext(ctx, nmcli, "connection", "modify", ifUUID, "con-name", originalConnectionName)
	cmd.Env = os.Environ()
	_, err = cmd.Output()
	if err != nil {
		klog.Error("failed to execute nmcli: %w", err)
		return fmt.Errorf("failed to execute nmcli: %w", err)
	}

	// find the interface MAC address
	mdata, err := os.ReadFile(fmt.Sprintf("/sys/class/net/%s/address", iface))
	if err != nil {
		klog.Error("read interface MAC address error, ", err)
		return err
	}

	mac := strings.TrimSpace(string(mdata))
	if mac == "" {
		klog.Error("interface MAC address is empty")
		return err
	}

	// create the bridge connection
	/*
		sudo nmcli connection add type bridge con-name br-olares ifname br-olares \
		connection.autoconnect yes bridge.stp no \
		ethernet.cloned-mac-address "$PHY_MAC" \
		ipv4.method auto ipv6.method ignore
	*/
	cmd = exec.CommandContext(ctx, nmcli,
		"connection", "add", "type", "bridge", "con-name", bridgeConnectionName, "ifname", bridgeConnectionName,
		"connection.autoconnect", "yes", "bridge.stp", "no",
		"ethernet.cloned-mac-address", mac,
		"ipv4.method", "auto",
		"ipv6.method", "ignore",
	)
	cmd.Env = os.Environ()
	_, err = cmd.Output()
	if err != nil {
		klog.Error("failed to execute nmcli: %w", err)
		return fmt.Errorf("failed to execute nmcli: %w", err)
	}

	// create the bridge slave connections
	/*
				sudo nmcli connection add type ethernet con-name "$SLAVE_CON" ifname "$PHY" \
		  		master br-olares slave-type bridge connection.autoconnect yes
	*/
	slaveConnectionName := fmt.Sprintf("%s%s", bridgeSlavePrefix, iface)
	cmd = exec.CommandContext(ctx, nmcli,
		"connection", "add", "type", "ethernet", "con-name", slaveConnectionName, "ifname", iface,
		"master", bridgeConnectionName, "slave-type", "bridge", "connection.autoconnect", "yes",
	)
	cmd.Env = os.Environ()
	_, err = cmd.Output()
	if err != nil {
		klog.Error("failed to execute nmcli: %w", err)
		// clean up the connection
		klog.Info("clean up the connection")
		cmd = exec.CommandContext(ctx, nmcli, "connection", "delete", bridgeConnectionName)
		cmd.Env = os.Environ()
		_, err = cmd.Output()
		if err != nil {
			klog.Error("failed to execute nmcli: %w", err)
		}

		cmd = exec.CommandContext(ctx, nmcli, "connection", "delete", slaveConnectionName)
		cmd.Env = os.Environ()
		_, err = cmd.Output()
		if err != nil {
			klog.Error("failed to execute nmcli: %w", err)
		}

		return fmt.Errorf("failed to execute nmcli: %w", err)
	}

	// turn on the bridge connection
	cmd = exec.CommandContext(ctx, "sh", "-c",
		fmt.Sprintf("%s connection modify %s connection.autoconnect no && %s connection down %s && %s connection up %s",
			nmcli, originalConnectionName, nmcli, originalConnectionName, nmcli, bridgeConnectionName))
	cmd.Env = os.Environ()
	_, err = cmd.Output()
	if err != nil {
		klog.Error("failed to execute nmcli: %w", err)
		// clean up the connection
		klog.Info("failed to turn on the bridge connection, reset the bridge connection")
		err = ResetBridgeConnection(ctx)
		if err != nil {
			klog.Error("failed to reset bridge connection: %w", err)
		}
	}

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
