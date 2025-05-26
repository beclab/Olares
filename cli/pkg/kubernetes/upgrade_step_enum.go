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

package kubernetes

type UpgradeStep int

const (
	ToV121 UpgradeStep = iota + 1
	ToV122
)

var UpgradeStepList = []UpgradeStep{
	ToV121,
	ToV122,
}

func (u UpgradeStep) String() string {
	switch u {
	case ToV121:
		return "to v1.21"
	case ToV122:
		return "to v1.22"
	default:
		return "invalid option"
	}
}
