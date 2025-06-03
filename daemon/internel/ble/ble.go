//go:build !linux
// +build !linux

package ble

import (
	"context"
	"errors"
)

func NewBleService(ctx context.Context) (*service, error) {
	return nil, errors.New("not implement")
}
