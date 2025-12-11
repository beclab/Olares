package utils

import (
	"github.com/enobufs/go-nats/nats"
)

var stunServer = "stun.sipgate.net:3478"

func IsNat() (bool, error) {
	n, err := nats.NewNATS(&nats.Config{
		Server: stunServer,
	})

	if err != nil {
		return false, err
	}

	var r *nats.DiscoverResult

	r, err = n.Discover()
	if err != nil {
		return false, err
	}
	return r.IsNatted, nil
}
