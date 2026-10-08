package upgrade

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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

	// A repeated signed request while the target exists still names the same
	// run. After cancellation removes it, a new authorization gets a new run
	// even when the operator selects the same version again.
	existing, err := state.GetOlaresUpgradeTarget()
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if !req.Version.Equal(&existing.Version) {
			return nil, fmt.Errorf("different upgrade version %s is already selected", existing.Version.Original())
		}
		if req.CliURL != existing.CliURL || req.WizardURL != existing.WizardURL || req.DownloadOnly != existing.DownloadOnly {
			return nil, fmt.Errorf("an upgrade to %s already exists with different release settings; cancel it first", req.Version.Original())
		}
		req = *existing
	} else {
		req.Downloaded = false
		var nonce [16]byte
		if _, err := rand.Read(nonce[:]); err != nil {
			return nil, fmt.Errorf("generate upgrade request nonce: %w", err)
		}
		req.RequestNonce = hex.EncodeToString(nonce[:])
	}
	if err = req.Save(); err != nil {
		return nil, fmt.Errorf("failed to create upgrade target: %v", err)
	}

	state.StateTrigger <- struct{}{}

	return NewExecutionRes(true, nil), nil
}
