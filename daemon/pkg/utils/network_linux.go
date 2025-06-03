//go:build linux
// +build linux

package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

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
