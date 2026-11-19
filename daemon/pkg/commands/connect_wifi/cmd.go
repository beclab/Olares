package connectwifi

import (
	"context"
	"errors"

	"github.com/beclab/Olares/daemon/internel/wifi"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

type connectWifi struct {
	commands.Operation
}

var _ commands.Interface = &connectWifi{}

func New() commands.Interface {
	return &connectWifi{
		Operation: commands.Operation{
			Name: commands.ConnectWifi,
		},
	}
}

func (s *connectWifi) Execute(ctx context.Context, p any) (res any, err error) {
	param, ok := p.(*Param)
	if !ok {
		err = errors.New("invalid param")
		return
	}

	err = wifi.Connect(ctx, *param)

	return
}
