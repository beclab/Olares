//go:build linux
// +build linux

package utils

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/beclab/Olares/daemon/cmd/terminusd/version"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
)

func tryToLoginSSH(user, password string) bool {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", "localhost:22", config)
	if err != nil {
		klog.Error("ssh dial error: ", err)
		return false
	}
	defer client.Close()
	return true
}

func checkDefaultSSHPassword() bool {
	defaultPasswords := map[string][]string{
		"olares": []string{
			"olares",
		},
	}

	for user, passwords := range defaultPasswords {
		for _, password := range passwords {
			if tryToLoginSSH(user, password) {
				return true
			}
		}
	}

	return false
}

func IsDefaultSSHPassword() bool {
	if version.VENDOR == "main" {
		return false
	}

	return checkDefaultSSHPassword()
}

func SetSSHPassword(user, password string) error {
	if password == "" {
		err := "password is empty"
		klog.Error(err)
		return errors.New(err)
	}

	if user == "" {
		err := "user is empty"
		klog.Error(err)
		return errors.New(err)
	}

	cmd := exec.Command("chpasswd")
	cmd.Stdin = strings.NewReader(user + ":" + password + "\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("set ssh password error: %v, output: %s", err, string(output))
	}

	return err
}
