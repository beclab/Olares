package umountusb

import (
	"context"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
)

type umountUsb struct {
	commands.Operation
}

var _ commands.Interface = &umountUsb{}

func New() commands.Interface {
	return &umountUsb{
		Operation: commands.Operation{
			Name: commands.UmountUsb,
		},
	}
}

func (i *umountUsb) Execute(ctx context.Context, p any) (res any, err error) {
	param, ok := p.(*Param)
	if !ok {
		err = commands.ErrInvalidParam
		return
	}

	err = utils.UmountUsbDevice(ctx, param.Path)

	return
}
