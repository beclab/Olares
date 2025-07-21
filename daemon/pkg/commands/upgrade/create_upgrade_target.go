package upgrade

import (
	"context"
	"errors"
	"fmt"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

type createUpgradeTarget struct {
	commands.Operation
}

var _ commands.Interface = &createUpgradeTarget{}

func NewCreateUpgradeTarget() commands.Interface {
	return &createUpgradeTarget{
		Operation: commands.Operation{
			Name: commands.CreateUpgradeTarget,
		},
	}
}

func (i *createUpgradeTarget) Execute(ctx context.Context, p any) (res any, err error) {
	req, ok := p.(state.UpgradeTarget)
	if !ok {
		return nil, errors.New("invalid param")
	}

	req.Downloaded = false
	if err = req.Save(); err != nil {
		return nil, fmt.Errorf("failed to create upgrade target: %v", err)
	}

	state.StateTrigger <- struct{}{}

	return NewExecutionRes(true, nil), nil
}
