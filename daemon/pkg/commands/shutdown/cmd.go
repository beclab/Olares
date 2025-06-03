package shutdown

import (
	"context"
	"time"

	"bytetrade.io/web3os/terminusd/pkg/commands"
)

type shutdown struct {
	commands.Operation
	*commands.BaseCommand
}

var _ commands.Interface = &shutdown{}

func New() commands.Interface {
	return &shutdown{
		Operation: commands.Operation{
			Name: commands.Shutdown,
		},
		BaseCommand: commands.NewBaseCommand(),
	}
}

func (s *shutdown) Execute(ctx context.Context, _ any) (res any, err error) {
	delay := time.NewTimer(2 * time.Second)

	go func() {
		<-delay.C
		s.BaseCommand.Run_(ctx, "shutdown", "0")
	}()

	return nil, nil
}
