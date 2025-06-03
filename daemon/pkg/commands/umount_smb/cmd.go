package umountsmb

import (
	"context"
	"errors"

	"bytetrade.io/web3os/terminusd/pkg/commands"
	"bytetrade.io/web3os/terminusd/pkg/utils"
	"k8s.io/klog/v2"
)

type umountSmb struct {
	commands.Operation
}

var _ commands.Interface = &umountSmb{}

func New() commands.Interface {
	return &umountSmb{
		Operation: commands.Operation{
			Name: commands.UmountSmb,
		},
	}
}

func (i *umountSmb) Execute(ctx context.Context, p any) (res any, err error) {
	param, ok := p.(*Param)
	if !ok {
		err = errors.New("invalid param")
		return
	}

	err = utils.UmountSambaDriver(ctx, param.MountPath)
	if err != nil {
		klog.Error("umount samba driver error, ", err)
	}

	return
}
