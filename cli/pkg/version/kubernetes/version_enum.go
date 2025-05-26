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

import (
	"fmt"

	versionutil "k8s.io/apimachinery/pkg/util/version"
)

type Version int

const (
	V119 Version = iota
	V120
	V121
	V122
	V123
	V124
)

var VersionList = []Version{
	V119,
	V120,
	V121,
	V122,
	V123,
	V124,
}

func (v Version) String() string {
	switch v {
	case V119:
		return "v1.19"
	case V120:
		return "v1.20"
	case V121:
		return "v1.21"
	case V122:
		return "v1.22"
	case V123:
		return "v1.23"
	case V124:
		return "v1.24"
	default:
		return "invalid option"
	}
}

func VersionSupport(version string) bool {
	K8sTargetVersion := versionutil.MustParseSemantic(version)
	for i := range VersionList {
		if VersionList[i].String() == fmt.Sprintf("v%v.%v", K8sTargetVersion.Major(), K8sTargetVersion.Minor()) {
			return true
		}
	}
	return false
}

// SupportedK8sVersionList returns the supported list of Kubernetes
func SupportedK8sVersionList() []string {
	return []string{
		"v1.19.0",
		"v1.19.8",
		"v1.19.9",
		"v1.20.4",
		"v1.20.6",
		"v1.20.10",
		"v1.21.0",
		"v1.21.1",
		"v1.21.2",
		"v1.21.3",
		"v1.21.4",
		"v1.21.5",
		"v1.21.6",
		"v1.21.7",
		"v1.21.8",
		"v1.21.9",
		"v1.21.10",
		"v1.21.11",
		"v1.21.12",
		"v1.21.13",
		"v1.21.14",
		"v1.22.0",
		"v1.22.1",
		"v1.22.2",
		"v1.22.3",
		"v1.22.4",
		"v1.22.5",
		"v1.22.6",
		"v1.22.7",
		"v1.22.8",
		"v1.22.9",
		"v1.22.10",
		"v1.22.11",
		"v1.22.12",
		"v1.23.0",
		"v1.23.1",
		"v1.23.2",
		"v1.23.3",
		"v1.23.4",
		"v1.23.5",
		"v1.23.6",
		"v1.23.7",
		"v1.23.8",
		"v1.23.9",
		"v1.23.10",
		"v1.24.0",
		"v1.24.1",
		"v1.24.2",
		"v1.24.3",
	}
}
