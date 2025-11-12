//go:build !linux
// +build !linux

package intranet

import (
	"errors"
)

func (d *DSRProxy) regonfigure() error {
	return errors.New("unsupported operation")
}
