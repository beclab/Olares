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

package module

import (
	"bytetrade.io/web3os/installer/pkg/core/cache"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/ending"
	"bytetrade.io/web3os/installer/pkg/core/hook"
	"github.com/pkg/errors"
)

type BaseModule struct {
	Name          string
	Desc          string
	Skip          bool
	ModuleCache   *cache.Cache
	PipelineCache *cache.Cache
	Runtime       connector.ModuleRuntime
	PostHook      []PostHookInterface
}

func (b *BaseModule) GetName() string {
	return b.Name
}

func (b *BaseModule) IsSkip() bool {
	return b.Skip
}

func (b *BaseModule) Default(runtime connector.ModuleRuntime, pipelineCache *cache.Cache, moduleCache *cache.Cache) {
	b.Runtime = runtime
	b.PipelineCache = pipelineCache
	b.ModuleCache = moduleCache
}

func (b *BaseModule) Init() {
	if b.Name == "" {
		b.Name = DefaultModuleName
	}
}

func (b *BaseModule) Is() string {
	return BaseModuleType
}

func (b *BaseModule) Run(result *ending.ModuleResult) {
	panic("implement me")
}

func (b *BaseModule) Until() (*bool, error) {
	return nil, nil
}

func (b *BaseModule) Slogan() {
}

func (b *BaseModule) AutoAssert() {
}

func (b *BaseModule) AppendPostHook(h PostHookInterface) {
	b.PostHook = append(b.PostHook, h)
}

func (b *BaseModule) CallPostHook(result *ending.ModuleResult) error {
	for i := range b.PostHook {
		h := b.PostHook[i]
		h.Init(b, result)
		if err := hook.Call(h); err != nil {
			return errors.Wrapf(err, "Module[%s] call post hook failed", b.Name)
		}
	}
	return nil
}
