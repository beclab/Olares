/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"

	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/cache"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/ending"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/rollback"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type RemoteTask struct {
	Name        string
	Desc        string
	Hosts       []connector.Host
	Prepare     prepare.Prepare
	Action      action.Action
	Rollback    rollback.Rollback
	Parallel    bool
	Retry       int
	Delay       time.Duration
	Timeout     time.Duration
	Concurrency float64

	PipelineCache *cache.Cache
	ModuleCache   *cache.Cache
	Runtime       connector.Runtime
	tag           string
	IgnoreError   bool
	TaskResult    *ending.TaskResult
}

func (t *RemoteTask) GetName() string {
	return t.Name
}

func (t *RemoteTask) GetDesc() string {
	return t.Desc
}

func (t *RemoteTask) Init(runtime connector.Runtime, moduleCache *cache.Cache, pipelineCache *cache.Cache) {
	t.ModuleCache = moduleCache
	t.PipelineCache = pipelineCache
	t.Runtime = runtime
	t.Default()
}

func (t *RemoteTask) Execute() *ending.TaskResult {
	if t.TaskResult.IsFailed() {
		return t.TaskResult
	}
	routinePool := make(chan struct{}, DefaultCon)
	defer close(routinePool)

	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout)
	defer cancel()
	wg := &sync.WaitGroup{}
	for i := range t.Hosts {
		if t.Hosts[i] == nil || t.Runtime.HostIsDeprecated(t.Hosts[i]) {
			continue
		}
		selfRuntime := t.Runtime.Copy()

		wg.Add(1)
		if t.Parallel {
			go t.RunWithTimeout(ctx, selfRuntime, t.Hosts[i], i, wg, routinePool)
		} else {
			t.RunWithTimeout(ctx, selfRuntime, t.Hosts[i], i, wg, routinePool)
		}
	}
	wg.Wait()

	if t.TaskResult.IsFailed() {
		t.TaskResult.ErrResult()
		return t.TaskResult
	}

	t.TaskResult.NormalResult()
	return t.TaskResult
}

func (t *RemoteTask) RunWithTimeout(ctx context.Context, runtime connector.Runtime, host connector.Host, index int,
	wg *sync.WaitGroup, pool chan struct{}) {
	pool <- struct{}{}

	resCh := make(chan error)
	go t.Run(runtime, host, index, resCh)

	select {
	case <-ctx.Done():
		t.TaskResult.AppendErr(host, fmt.Errorf("execute task timeout, Timeout=%s", util.ShortDur(t.Timeout)))
	case e := <-resCh:
		if e != nil {
			t.TaskResult.AppendErr(host, e)
		}
	}

	<-pool
	wg.Done()
}

func (t *RemoteTask) Run(runtime connector.Runtime, host connector.Host, index int, resCh chan error) {
	var res error
	defer func() {
		//runtime.GetConnector().Close(host)

		resCh <- res
		close(resCh)
	}()

	if err := t.ConfigureSelfRuntime(runtime, host, index); err != nil {
		res = err
		return
	}

	t.Prepare.Init(t.ModuleCache, t.PipelineCache)
	t.Prepare.AutoAssert(runtime)
	if ok, err := t.WhenWithRetry(runtime); !ok {
		if err != nil {
			res = err
			return
		} else {
			t.TaskResult.AppendSkip(host)
			return
		}
	}

	t.Action.Init(t.ModuleCache, t.PipelineCache)
	t.Action.AutoAssert(runtime)
	if err := t.ExecuteWithRetry(runtime); err != nil {
		res = err
		return
	}

	t.TaskResult.AppendSuccess(host)
	return
}

func (t *RemoteTask) ConfigureSelfRuntime(runtime connector.Runtime, host connector.Host, index int) error {
	var conn connector.Connection
	var err error
	if host.GetAddress() != runtime.GetSystemInfo().GetLocalIp() || host.GetInternalAddress() != runtime.GetSystemInfo().GetLocalIp() {
		conn, err = runtime.GetConnector().Connect(host)
		if err != nil {
			return errors.Wrapf(err, "failed to connect to %s", host.GetAddress())
		}
	}
	r := &connector.Runner{
		Conn: conn,
		//Debug: runtime.Arg.Debug,
		Host:  host,
		Index: index,
	}
	runtime.SetRunner(r)
	return nil
}

func (t *RemoteTask) When(runtime connector.Runtime) (bool, error) {
	if t.Prepare == nil {
		return true, nil
	}
	if ok, err := t.Prepare.PreCheck(runtime); err != nil {
		return false, err
	} else if !ok {
		return false, nil
	}
	return true, nil
}

func (t *RemoteTask) WhenWithRetry(runtime connector.Runtime) (bool, error) {
	pass := false
	err := fmt.Errorf("pre-check exec failed after %d retires ", t.Retry)
	for i := 0; i < t.Retry; i++ {
		if res, e := t.When(runtime); e != nil {
			logger.Errorf("%s %s", runtime.RemoteHost().GetName(), e.Error())

			if i == t.Retry-1 {
				err = errors.New(err.Error() + e.Error())
				continue
			}
			logger.Infof("retry: [%s]", runtime.RemoteHost().GetName())
			time.Sleep(t.Delay)
			continue
		} else {
			err = nil
			pass = res
			break
		}
	}

	return pass, err
}

func (t *RemoteTask) ExecuteWithRetry(runtime connector.Runtime) error {
	err := fmt.Errorf("[A] %s: %s exec failed after %d retires: ", t.GetName(), t.Name, t.Retry)
	for i := 0; i < t.Retry; i++ {
		e := t.Action.Execute(runtime)
		if e != nil {
			logger.Errorf("%s %s", runtime.RemoteHost().GetName(), e.Error())

			if i == t.Retry-1 {
				err = errors.New(err.Error() + e.Error())
				continue
			}
			logger.Debugf("Remote retry: [%s] %s", runtime.RemoteHost().GetName(), t.Name)
			time.Sleep(t.Delay)
			continue
		} else {
			err = nil
			break
		}
	}
	return err
}

func (t *RemoteTask) ExecuteRollback() {
	if t.Rollback == nil {
		return
	}
	if !t.TaskResult.IsFailed() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout)
	defer cancel()
	routinePool := make(chan struct{}, DefaultCon)
	defer close(routinePool)

	rwg := &sync.WaitGroup{}
	for i, ar := range t.TaskResult.ActionResults {
		if ar.Host == nil || t.Runtime.HostIsDeprecated(ar.Host) {
			continue
		}
		selfRuntime := t.Runtime.Copy()

		if t.Parallel {
			go t.RollbackWithTimeout(ctx, selfRuntime, ar.Host, i, ar, rwg, routinePool)
		} else {
			t.RollbackWithTimeout(ctx, selfRuntime, ar.Host, i, ar, rwg, routinePool)
		}
		rwg.Add(1)
	}
	rwg.Wait()
}

func (t *RemoteTask) RollbackWithTimeout(ctx context.Context, runtime connector.Runtime, host connector.Host, index int,
	result *ending.ActionResult, wg *sync.WaitGroup, pool chan struct{}) {

	pool <- struct{}{}

	resCh := make(chan error)
	go t.RunRollback(runtime, host, index, result, resCh)

	select {
	case <-ctx.Done():
		logger.Warnf("rollback-failed: [%s]", runtime.RemoteHost().GetName())
		logger.Errorf("%s execute task timeout, Timeout=%d", runtime.RemoteHost().GetName(), t.Timeout)
	case e := <-resCh:
		if e != nil {
			logger.Warnf("rollback-failed: [%s]", runtime.RemoteHost().GetName())
			logger.Errorf("%s %s", runtime.RemoteHost().GetName(), e.Error())
		}
	}

	<-pool
	wg.Done()
}

func (t *RemoteTask) RunRollback(runtime connector.Runtime, host connector.Host, index int, result *ending.ActionResult, resCh chan error) {
	var res error
	defer func() {
		//runtime.GetConnector().Close(host)

		resCh <- res
		close(resCh)
	}()

	if err := t.ConfigureSelfRuntime(runtime, host, index); err != nil {
		res = err
		return
	}

	logger.Infof("rollback: [%s]", runtime.RemoteHost().GetName())

	t.Rollback.Init(t.ModuleCache, t.PipelineCache)
	t.Rollback.AutoAssert(runtime)
	if err := t.Rollback.Execute(runtime, result); err != nil {
		res = err
		return
	}
	return
}

func (t *RemoteTask) Default() {
	t.TaskResult = ending.NewTaskResult()
	if t.Name == "" {
		t.Name = DefaultTaskName
	}

	if t.Prepare == nil {
		t.Prepare = new(prepare.BasePrepare)
	}

	if t.Retry <= 0 {
		t.Retry = 3
	}

	if t.Delay <= 0 {
		t.Delay = 5 * time.Second
	}

	if t.Timeout <= 0 {
		t.Timeout = DefaultTimeout * time.Minute
	}

	if t.Concurrency <= 0 || t.Concurrency > 1 {
		t.Concurrency = 1
	}
}

func (t *RemoteTask) calculateConcurrency() int {
	num := t.Concurrency * float64(len(t.Hosts))
	res := int(util.Round(num, 0))
	if res < 1 {
		res = 1
	}
	return res
}
