package sshpassword

import (
	"context"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
)

type setSSHPassword struct {
	commands.Operation
}

var _ commands.Interface = &setSSHPassword{}

func New() commands.Interface {
	return &setSSHPassword{
		Operation: commands.Operation{
			Name: commands.SetSSHPassword,
		},
	}
}

func (s *setSSHPassword) Execute(ctx context.Context, p any) (res any, err error) {
	param, ok := p.(*Param)
	if !ok {
		err = commands.ErrInvalidParam
		return
	}

	if param.Username == "" {
		param.Username = "olares"
	}

	err = utils.SetSSHPassword(param.Password)
	return
}
