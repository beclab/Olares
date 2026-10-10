package handlers

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/gofiber/fiber/v2"
)

type installRecorder struct {
	commands.Operation
	called bool
	param  any
}

func (c *installRecorder) Execute(_ context.Context, param any) (any, error) {
	c.called, c.param = true, param
	return nil, nil
}

func TestInstallAcceptsNoUserInformation(t *testing.T) {
	cmd := &installRecorder{Operation: commands.Operation{Name: commands.Install}}
	h := &Handlers{mainCtx: context.Background()}
	app := fiber.New()
	app.Post("/command/install", func(ctx *fiber.Ctx) error { return h.PostTerminusInit(ctx, cmd) })
	resp, err := app.Test(httptest.NewRequest("POST", "/command/install", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 || !cmd.called || cmd.param != nil {
		t.Fatalf("install did not start without user information: %d", resp.StatusCode)
	}
}
