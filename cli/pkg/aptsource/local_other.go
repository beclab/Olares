//go:build !linux

package aptsource

import "fmt"

func EnsureDebianComponents() error {
	return fmt.Errorf("Debian APT preparation requires Linux")
}
