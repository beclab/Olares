package hello

import (
	"context"
	"fmt"
	"time"

	"bytetrade.io/web3os/installer/pkg/core/ending"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type HelloHook struct {
	Module module.Module
	Result *ending.ModuleResult
}

func (h *HelloHook) Init(module module.Module, result *ending.ModuleResult) {
	fmt.Println("---hello hook / init---")
	h.Module = module
	h.Result = result
	h.Result.StartTime = time.Now()
}

func (h *HelloHook) Try() error {
	fmt.Println("---hello hook / try---", h.Result.StartTime.String())
	_, _, err := util.Exec(context.Background(), "echo 'hello, world!!!!!'", true, false)

	if err != nil {
		h.Result.ErrResult(err)
		return err
	}

	return nil
}

func (h *HelloHook) Catch(err error) error {
	fmt.Println("---hello hook / Cache---", err)
	time.Sleep(5 * time.Second)
	return nil
}

func (h *HelloHook) Finally() {
	fmt.Println("---hello hook / Finally---")
	h.Result.EndTime = time.Now()
	sayHello := h.Result.Status.String()
	logger.Infof(">>>> %s %s", sayHello, h.Result.EndTime.String())
}
