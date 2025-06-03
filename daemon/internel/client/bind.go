package client

import "context"

const (
	ClIENT_CONTEXT = "binding-client"
)

type Client interface {
}

type termipass struct {
	Client
	jws string
}

func NewTermipassClient(ctx context.Context, jws string) (Client, error) {
	c := &termipass{jws: jws}

	return c, c.validateJWS(ctx)
}
