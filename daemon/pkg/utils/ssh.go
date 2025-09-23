//go:build !linux
// +build !linux

package utils

func IsDefaultSSHPassword() bool {
	return false
}

func SetSSHPassword(password string) error {
	return nil
}
