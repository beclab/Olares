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

package pipeline

import (
	"os"
	"sync"
	"time"

	"bytetrade.io/web3os/installer/pkg/core/cache"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/ending"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"github.com/pkg/errors"
)

// var pipelineQueue *PipelineQueue

// type PipelineQueue struct {
// 	pipelines []*Pipeline
// 	sync.Mutex
// }

type Pipeline struct {
	Name            string
	StartAt         time.Time
	Modules         []module.Module
	Runtime         connector.Runtime
	SpecHosts       int
	PipelineCache   *cache.Cache
	ModuleCachePool sync.Pool
	ModulePostHooks []module.PostHookInterface
}

func (p *Pipeline) Init() error {
	if p.PipelineCache == nil {
		p.PipelineCache = cache.NewCache()
	}

	p.SpecHosts = len(p.Runtime.GetAllHosts())
	if err := p.Runtime.GenerateWorkDir(); err != nil {
		return err
	}

	p.StartAt = time.Now()

	return nil
}

func (p *Pipeline) Start() error {
	logger.Infof("[Job] [%s] start ...", p.Name)
	if err := p.Init(); err != nil {
		logger.Errorf("[Job] %s execute failed %v", p.Name, err)
		return errors.Wrapf(err, "Job %s execute failed", p.Name)
	}
	for i := range p.Modules {
		m := p.Modules[i]
		if m.IsSkip() {
			continue
		}

		moduleCache := p.newModuleCache()
		m.Default(p.Runtime, p.PipelineCache, moduleCache)
		m.AutoAssert()
		m.Init()
		logger.Infof("[Module] %s", m.GetName())
		for j := range p.ModulePostHooks {
			m.AppendPostHook(p.ModulePostHooks[j])
		}
		res := p.RunModule(m)
		err := m.CallPostHook(res)
		if res.IsFailed() {
			logger.Errorf("[Job] [%s] execute failed %v", p.Name, err)
			// return errors.Wrapf(res.CombineResult, "Pipeline[%s] execute failed", p.Name)
			return res.CombineResult
		}
		if err != nil {
			logger.Errorf("[Job] [%s] execute failed %v", p.Name, err)
			// return errors.Wrapf(err, "Job[%s] execute failed", p.Name)
			return err
		}
		p.releaseModuleCache(moduleCache)
	}
	p.releasePipelineCache()

	// close ssh connect
	for _, host := range p.Runtime.GetAllHosts() {
		p.Runtime.GetConnector().Close(host)
	}

	if p.SpecHosts != len(p.Runtime.GetAllHosts()) {
		logger.Errorf("[Job] %s execute failed: there are some error in your spec hosts", p.Name)
		return errors.Errorf("[Job] %s execute failed: there are some error in your spec hosts", p.Name)
	}
	logger.Infof("[Job] %s execute successfully!!! (%s)", p.Name, p.since())
	logger.Sync()

	return nil
}

func (p *Pipeline) RunModule(m module.Module) *ending.ModuleResult {
	m.Slogan()

	result := ending.NewModuleResult()
	for {
		switch m.Is() {
		case module.TaskModuleType:
			m.Run(result)
			if result.IsFailed() {
				return result
			}

		case module.GoroutineModuleType:
			go func() {
				m.Run(result)
				if result.IsFailed() {
					os.Exit(1)
				}
			}()
		default:
			m.Run(result)
			if result.IsFailed() {
				return result
			}
		}

		stop, err := m.Until()
		if err != nil {
			result.LocalErrResult(err)
			return result
		}
		if stop == nil || *stop == true {
			break
		}
	}
	return result
}

func (p *Pipeline) newModuleCache() *cache.Cache {
	moduleCache, ok := p.ModuleCachePool.Get().(*cache.Cache)
	if ok {
		return moduleCache
	}
	return cache.NewCache()
}

func (p *Pipeline) releasePipelineCache() {
	p.PipelineCache.Clean()
}

func (p *Pipeline) releaseModuleCache(c *cache.Cache) {
	c.Clean()
	p.ModuleCachePool.Put(c)
}

func (p *Pipeline) since() string {
	return util.ShortDur(time.Since(p.StartAt))
}
