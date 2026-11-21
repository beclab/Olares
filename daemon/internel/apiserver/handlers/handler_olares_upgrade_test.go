package handlers

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/commands/upgrade"
	"github.com/gofiber/fiber/v2"
)

type refusingCancel struct{ commands.Operation }

func (*refusingCancel) Execute(context.Context, any) (any, error) {
	return nil, upgrade.ErrUpgradeCommitted
}

func TestCancellingACommittedUpgradeIsAConflict(t *testing.T) {
	h := &Handlers{}
	app := fiber.New()
	app.Delete("/upgrade", func(c *fiber.Ctx) error { return h.CancelOlaresUpgrade(c, &refusingCancel{}) })
	resp, err := app.Test(httptest.NewRequest("DELETE", "/upgrade", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusConflict)
	}
}
