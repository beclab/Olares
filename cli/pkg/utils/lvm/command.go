package lvm

import (
	"bytes"
	"errors"
	"os/exec"

	"k8s.io/klog/v2"
)

type command[T any] struct {
	cmd         string
	defaultArgs []string
	format      func(data []byte) (T, error)
}

func (c *command[T]) Run(args ...string) (*T, string, error) {
	if c.cmd == "" {
		return nil, "", errors.ErrUnsupported
	}

	allArgs := append(c.defaultArgs, args...)
	o, e, err := runCommandSplit(c.cmd, allArgs...)
	if err != nil {
		klog.Errorf("command %s failed: %s", c.cmd, string(e))
		return nil, string(e), err
	}

	result, err := c.format(o)
	if err != nil {
		klog.Errorf("command %s format failed: %v", c.cmd, err)
		return nil, "", err
	}

	return &result, "", nil
}

func runCommandSplit(command string, args ...string) ([]byte, []byte, error) {
	var cmdStdout bytes.Buffer
	var cmdStderr bytes.Buffer

	cmd := exec.Command(command, args...)
	cmd.Stdout = &cmdStdout
	cmd.Stderr = &cmdStderr
	err := cmd.Run()

	output := cmdStdout.Bytes()
	error_output := cmdStderr.Bytes()

	return output, error_output, err
}

func findCmd(cmd string) string {
	c := exec.Command("command", "-v", cmd)
	path, err := c.Output()
	if err != nil {
		return ""
	}
	return string(bytes.TrimSpace(path))
}
