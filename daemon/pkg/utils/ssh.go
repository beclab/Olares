//go:build !linux
// +build !linux

package utils

func IsDefaultSSHPassword() bool {
	return false
}

func SetSSHPassword(user, password string) error {
	return nil
}
