package client

import "context"

const (
	ClIENT_CONTEXT = "binding-client"
)

type Client interface {
	OlaresID() string
}

var _ Client = &termipass{}

type termipass struct {
	jws      string
	olaresID string
}

// OlaresID implements Client.
func (c *termipass) OlaresID() string {
	return c.olaresID
}

func NewTermipassClient(ctx context.Context, jws string) (Client, error) {
	c := &termipass{jws: jws}
	err, olaresID := c.validateJWS(ctx)
	if err != nil {
		return nil, err
	}

	c.olaresID = olaresID
	return c, nil
}
