package reboot

import (
	"context"
	"time"

	"github.com/beclab/Olares/daemon/pkg/commands"
)

type reboot struct {
	commands.Operation
	*commands.BaseCommand
}

var _ commands.Interface = &reboot{}

func New() commands.Interface {
	return &reboot{
		Operation: commands.Operation{
			Name: commands.Reboot,
		},
		BaseCommand: commands.NewBaseCommand(),
	}
}

func (s *reboot) Execute(ctx context.Context, _ any) (res any, err error) {
	delay := time.NewTimer(2 * time.Second)
	go func() {
		<-delay.C
		s.BaseCommand.Run_(ctx, "reboot")
	}()

	return nil, nil
}
