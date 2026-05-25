package umountsmb

import (
	"context"
	"errors"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
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
		return
	}

	if removeErr := utils.RemoveMountRecord(commands.MOUNT_RECORDS_FILE, param.MountPath); removeErr != nil {
		klog.Warning("remove mount record error, ", removeErr)
	}

	return
}
