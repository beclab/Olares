package changehost

import (
	"context"
	"errors"
	"os"

	"bytetrade.io/web3os/terminusd/pkg/cluster/state"
	"bytetrade.io/web3os/terminusd/pkg/commands"
	"bytetrade.io/web3os/terminusd/pkg/nets"
	"k8s.io/klog/v2"
)

type changeHost struct {
	commands.Operation
}

var _ commands.Interface = &changeHost{}

func New() commands.Interface {
	return &changeHost{
		Operation: commands.Operation{
			Name: commands.ChangeHost,
		},
	}
}

func (i *changeHost) Execute(ctx context.Context, p any) (res any, err error) {
	param, ok := p.(*Param)
	if !ok {
		err = errors.New("invalid param")
		return
	}

	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	hostip, err := nets.GetHostIpFromHostsFile(host)
	if err != nil {
		return nil, err
	}

	if hostip == param.IP {
		return nil, errors.New("the ip you chose is current host ip")
	}

	err = nets.WriteIpToHostsFile(param.IP, host)
	if err != nil {
		return nil, err
	}

	// trigger the state watcher
	state.StateTrigger <- struct{}{}

	klog.Info("host ip is changed to ", param.IP)
	return nil, nil
}
