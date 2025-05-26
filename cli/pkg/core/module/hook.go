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
	"bytetrade.io/web3os/installer/pkg/core/ending"
	"bytetrade.io/web3os/installer/pkg/core/hook"
)

type PostHookInterface interface {
	hook.Interface
	Init(module Module, result *ending.ModuleResult)
}

type PostHook struct {
	Module Module
	Result *ending.ModuleResult
}

func (p *PostHook) Try() error {
	panic("implement me")
}

func (p *PostHook) Catch(err error) error {
	return err
}

func (p *PostHook) Finally() {
}

func (p *PostHook) Init(module Module, result *ending.ModuleResult) {
	p.Module = module
	p.Result = result
}
