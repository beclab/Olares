package mountsmb

import (
	"context"
	"errors"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

type mountSmb struct {
	commands.Operation
}

var _ commands.Interface = &mountSmb{}

func New() commands.Interface {
	return &mountSmb{
		Operation: commands.Operation{
			Name: commands.MountSmb,
		},
	}
}

func (i *mountSmb) Execute(ctx context.Context, p any) (res any, err error) {
	param, ok := p.(*Param)
	if !ok {
		err = errors.New("invalid param")
		return
	}

	err = utils.MountSambaDriver(ctx, param.MountBaseDir, param.SmbPath, param.User, param.Password)
	if err != nil {
		klog.Error("mount samba driver error, ", err)
	}

	return
}
