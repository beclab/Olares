//go:build !linux
// +build !linux

package utils

import (
	"context"
	"errors"
)

func KernelSupportsAltname() bool { return false }

func SelectOverlayParent(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func ResolveOverlayParent(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func EnsureOverlayParentAltname(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func OverlayParentLinkUp(ctx context.Context) (string, bool, error) {
	return "", false, errors.New("not implemented")
}

func ConvergeOverlayGateway(ctx context.Context) {}
