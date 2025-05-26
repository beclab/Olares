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

package prepare

import (
	"bytetrade.io/web3os/installer/pkg/core/cache"
	"bytetrade.io/web3os/installer/pkg/core/connector"
)

type BasePrepare struct {
	ModuleCache   *cache.Cache
	PipelineCache *cache.Cache
}

func (b *BasePrepare) Init(moduleCache *cache.Cache, pipelineCache *cache.Cache) {
	b.ModuleCache = moduleCache
	b.PipelineCache = pipelineCache
}

func (b *BasePrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	return true, nil
}

func (b *BasePrepare) AutoAssert(runtime connector.Runtime) {
}

type PrepareCollection []Prepare

func (p *PrepareCollection) Init(cache *cache.Cache, rootCache *cache.Cache) {
	for _, v := range *p {
		v.Init(cache, rootCache)
	}
}

func (p *PrepareCollection) PreCheck(runtime connector.Runtime) (bool, error) {
	for _, v := range *p {
		res, err := v.PreCheck(runtime)
		if err != nil {
			return false, err
		}
		if res == false {
			return false, nil
		}
	}
	return true, nil
}

func (p *PrepareCollection) AutoAssert(runtime connector.Runtime) {
	for _, v := range *p {
		v.AutoAssert(runtime)
	}
}
