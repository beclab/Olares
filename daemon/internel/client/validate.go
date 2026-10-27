package client

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

func (c *termipass) validateJWS(_ context.Context) (error, string, any) {
	if strings.TrimSpace(c.jws) == "" {
		klog.Error("jws is empty")
		return errors.New("invalid jws, jws is empty"), "", nil
	}

	signed, err := utils.ValidateJWSDetail(c.jws)
	if err != nil {
		klog.Error("jws validation failed, ", err)
		return fmt.Errorf("invalid jws, %v", err), "", nil
	}
	return nil, signed.OlaresID, signed.Body
}
