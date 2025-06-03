package client

import (
	"context"
	"errors"
	"strings"

	"k8s.io/klog/v2"
)

func (c *termipass) validateJWS(_ context.Context) error {
	if strings.TrimSpace(c.jws) == "" {
		klog.Error("jws is empty")
		return errors.New("invalid jws")
	}

	// if state.CurrentState.TerminusState == state.TerminusRunning {
	// 	client, err := utils.GetDynamicClient()
	// 	if err != nil {
	// 		klog.Error("get k8s client error, ", err)
	// 		return err
	// 	}

	// 	jws, err := utils.GetAdminUserJws(ctx, client)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if c.jws != jws {
	// 		return errors.New("invalid jws of admin user")
	// 	}
	// }

	// TODO: validate jws on blockchain

	return nil
}
