package lvm

import (
	"bytes"
	"errors"
	"log"
	"os/exec"
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
		return nil, string(e), err
	}

	result, err := c.format(o)
	if err != nil {
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
	path, err := exec.LookPath(cmd)
	if err != nil {
		log.Printf("failed to find command %s: %v\n", cmd, err)
		return ""
	}
	return path
}
