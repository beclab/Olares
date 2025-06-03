package umountusb

import (
	"context"
	"errors"

	"bytetrade.io/web3os/terminusd/pkg/commands"
	"bytetrade.io/web3os/terminusd/pkg/utils"
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
		err = errors.New("invalid param")
		return
	}

	err = utils.UmountUsbDevice(ctx, param.Path)

	return
}
