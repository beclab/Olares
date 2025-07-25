/*
 Copyright 2022 The KubeSphere Authors.

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

package rollback

import (
	"github.com/beclab/Olares/cli/pkg/core/cache"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/ending"
)

type BaseRollback struct {
	ModuleCache   *cache.Cache
	PipelineCache *cache.Cache
}

func (b *BaseRollback) Init(moduleCache *cache.Cache, pipelineCache *cache.Cache) {
	b.ModuleCache = moduleCache
	b.PipelineCache = pipelineCache
}

func (b *BaseRollback) Execute(runtime connector.Runtime, result *ending.ActionResult) error {
	return nil
}

func (b *BaseRollback) AutoAssert(runtime connector.Runtime) {

}
