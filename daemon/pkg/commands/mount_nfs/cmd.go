package mountnfs

import (
	"context"
	"errors"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

type mountNfs struct {
	commands.Operation
}

var _ commands.Interface = &mountNfs{}

func New() commands.Interface {
	return &mountNfs{
		Operation: commands.Operation{
			Name: commands.MountNfs,
		},
	}
}

func (i *mountNfs) Execute(ctx context.Context, p any) (res any, err error) {
	param, ok := p.(*Param)
	if !ok {
		err = errors.New("invalid param")
		return
	}

	err = utils.MountNfsDriver(ctx, param.MountBaseDir, param.MountPath, param.Server, param.NfsPath)
	if err != nil {
		klog.Error("mount nfs driver error, ", err)
	}

	return
}
