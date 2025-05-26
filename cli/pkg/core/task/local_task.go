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

type LocalTask struct {
	Name string
	Desc string

	Prepare  prepare.Prepare
	Action   action.Action
	Rollback rollback.Rollback
	Retry    int
	Delay    time.Duration
	Timeout  time.Duration

	PipelineCache *cache.Cache
	ModuleCache   *cache.Cache
	Runtime       connector.Runtime
	tag           string
	IgnoreError   bool
	TaskResult    *ending.TaskResult
}

func (l *LocalTask) GetName() string {
	return l.Name
}

func (l *LocalTask) GetDesc() string {
	return l.Desc
}

func (l *LocalTask) Init(runtime connector.Runtime, moduleCache *cache.Cache, pipelineCache *cache.Cache) {
	l.ModuleCache = moduleCache
	l.PipelineCache = pipelineCache
	l.Runtime = runtime
	l.Default()
}

func (l *LocalTask) Default() {
	l.TaskResult = ending.NewTaskResult()
	if l.Name == "" {
		l.Name = DefaultTaskName
	}

	if l.Prepare == nil {
		l.Prepare = new(prepare.BasePrepare)
	}

	if l.Action == nil {
		l.TaskResult.AppendErr(nil, errors.New("the action is nil"))
		return
	}

	if l.Retry <= 0 {
		l.Retry = 1
	}

	if l.Delay <= 0 {
		l.Delay = 5 * time.Second
	}

	if l.Timeout <= 0 {
		l.Timeout = DefaultTimeout * time.Minute
	}
}

func (l *LocalTask) Execute() *ending.TaskResult {
	if l.TaskResult.IsFailed() {
		return l.TaskResult
	}

	host := l.Runtime.GetLocalHost()
	selfRuntime := l.Runtime.Copy()
	l.RunWithTimeout(selfRuntime, host)

	if l.TaskResult.IsFailed() {
		l.TaskResult.ErrResult()
		return l.TaskResult
	}

	l.TaskResult.NormalResult()
	return l.TaskResult
}

func (l *LocalTask) RunWithTimeout(runtime connector.Runtime, host connector.Host) {
	ctx, cancel := context.WithTimeout(context.Background(), l.Timeout)
	defer cancel()

	resCh := make(chan error)

	go l.Run(runtime, host, resCh)
	select {
	case <-ctx.Done():
		l.TaskResult.AppendErr(host, fmt.Errorf("execute task timeout, Timeout=%s", util.ShortDur(l.Timeout)))
	case e := <-resCh:
		if e != nil {
			l.TaskResult.AppendErr(host, e)
		}
	}
}

func (l *LocalTask) Run(runtime connector.Runtime, host connector.Host, resCh chan error) {
	var res error
	defer func() {
		resCh <- res
		close(resCh)
	}()

	runtime.SetRunner(&connector.Runner{
		Conn: nil,
		//Debug: runtime.Arg.Debug,
		Host: host,
	})

	l.Prepare.Init(l.ModuleCache, l.PipelineCache)
	l.Prepare.AutoAssert(runtime)
	if ok, err := l.WhenWithRetry(runtime, host); !ok {
		if err != nil {
			res = err
			return
		} else {
			l.TaskResult.AppendSkip(host)
			return
		}
	}

	l.Action.Init(l.ModuleCache, l.PipelineCache)
	l.Action.AutoAssert(runtime)
	if err := l.ExecuteWithRetry(runtime, host); err != nil {
		res = err
		return
	}
	l.TaskResult.AppendSuccess(host)
}

func (l *LocalTask) WhenWithRetry(runtime connector.Runtime, host connector.Host) (bool, error) {
	pass := false
	err := fmt.Errorf("pre-check exec failed after %d retires ", l.Retry)
	for i := 0; i < l.Retry; i++ {
		if res, e := l.When(runtime); e != nil {
			logger.Errorf("[P] %s: %s", host.GetName(), e.Error())

			if i == l.Retry-1 {
				err = errors.New(err.Error() + e.Error())
				continue
			}
			logger.Infof("retry: [%s]", host.GetName())
			time.Sleep(l.Delay)
			continue
		} else {
			err = nil
			pass = res
			break
		}
	}

	return pass, err
}

func (l *LocalTask) When(runtime connector.Runtime) (bool, error) {
	if l.Prepare == nil {
		return true, nil
	}
	if ok, err := l.Prepare.PreCheck(runtime); err != nil {
		return false, err
	} else if !ok {
		return false, nil
	}
	return true, nil
}

func (l *LocalTask) ExecuteWithRetry(runtime connector.Runtime, host connector.Host) error {
	err := fmt.Errorf("[A] %s: %s exec failed after %d retires: ",
		host.GetName(), l.Name, l.Retry)
	for i := 0; i < l.Retry; i++ {

		e := l.Action.Execute(runtime)
		if e != nil {
			logger.Errorf("%s %s", host.GetName(), e.Error())

			if i == l.Retry-1 {
				err = errors.New(err.Error() + e.Error())
				continue
			}
			logger.Infof("Local retry: [%s] %s", host.GetName(), l.Name)
			time.Sleep(l.Delay)
			continue
		} else {
			err = nil
			break
		}
	}
	return err
}

func (l *LocalTask) ExecuteRollback() {
	if l.Rollback == nil {
		return
	}
	if !l.TaskResult.IsFailed() {
		return
	}
}
