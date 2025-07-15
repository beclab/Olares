package client

import (
	"context"
	"errors"
	"strings"

	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

func (c *termipass) validateJWS(_ context.Context) (error, string) {
	if strings.TrimSpace(c.jws) == "" {
		klog.Error("jws is empty")
		return errors.New("invalid jws"), ""
	}

	if ok, olaresID := utils.ValidateJWS(c.jws); ok {
		return nil, olaresID
	} else {
		klog.Error("jws validation failed")
		return errors.New("invalid jws"), ""
	}
}
