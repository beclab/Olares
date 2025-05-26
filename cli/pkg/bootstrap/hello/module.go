package hello

import (
	"bytetrade.io/web3os/installer/pkg/core/module"
)

type HelloModule struct {
	module.BaseTaskModule
}

func (h *HelloModule) Init() {
	h.Name = "HelloModule"
	h.Desc = "Say Hello"

	h.PostHook = []module.PostHookInterface{
		&HelloHook{},
	}
}
