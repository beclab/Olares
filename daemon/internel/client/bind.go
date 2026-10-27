package client

import "context"

const (
	ClIENT_CONTEXT = "binding-client"
	USER_CONTEXT   = "auth-user"
)

type Client interface {
	OlaresID() string
}

// SignedClient additionally exposes what the signature actually covered.
//
// Knowing only who signed makes every signature a bearer credential for every
// route that checks one. A handler for a dangerous route reads the signed body
// and refuses unless it names the request that arrived.
type SignedClient interface {
	Client
	SignedBody() any
}

var _ SignedClient = &termipass{}

type termipass struct {
	jws      string
	olaresID string
	body     any
}

// OlaresID implements Client.
func (c *termipass) OlaresID() string {
	return c.olaresID
}

// SignedBody implements SignedClient.
func (c *termipass) SignedBody() any {
	return c.body
}

func NewTermipassClient(ctx context.Context, jws string) (Client, error) {
	c := &termipass{jws: jws}
	err, olaresID, body := c.validateJWS(ctx)
	if err != nil {
		return nil, err
	}

	c.olaresID = olaresID
	c.body = body
	return c, nil
}
