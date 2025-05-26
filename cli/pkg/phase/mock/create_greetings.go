package mock

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/hello"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
)

func NewGreetingsPipeline(runtime *common.LocalRuntime) error {
	m := []module.Module{
		&hello.HelloModule{},
	}

	p := pipeline.Pipeline{
		Name:    "GreetingsPipeline",
		Modules: m,
		Runtime: runtime,
	}

	go p.Start()

	return nil
}

func Greetings() error {
	runtime, err := common.NewLocalRuntime(false, false)
	if err != nil {
		return err
	}

	if err := NewGreetingsPipeline(&runtime); err != nil {
		return err
	}
	return nil
}
