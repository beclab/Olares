package umountnfs

import (
	"context"
	"errors"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

type umountNfs struct {
	commands.Operation
}

var _ commands.Interface = &umountNfs{}

func New() commands.Interface {
	return &umountNfs{
		Operation: commands.Operation{
			Name: commands.UmountNfs,
		},
	}
}

func (i *umountNfs) Execute(ctx context.Context, p any) (res any, err error) {
	param, ok := p.(*Param)
	if !ok {
		err = errors.New("invalid param")
		return
	}

	err = utils.UmountNfsDriver(ctx, param.MountPath)
	if err != nil {
		klog.Error("umount nfs driver error, ", err)
	}

	return
}
