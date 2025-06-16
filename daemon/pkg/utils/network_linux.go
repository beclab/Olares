//go:build linux
// +build linux

package utils

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
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
	return deviceStatus(ctx, func(d *Device) bool { return true })
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
			klog.Info(string(output))
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
			if len(ipAndMask) > 2 {
				device.Ipv4Address = ipAndMask[0]
				cidr, err := strconv.Atoi(ipAndMask[1])
				if err != nil {
					klog.Error("convert cidr error, ", err)
					return err
				}
				mask, err := MaskFromCIDR(cidr)
				if err != nil {
					klog.Error("get mask from cidr error, ", err)
					return err
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
				klog.Error("get connection method error, ", err)
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
