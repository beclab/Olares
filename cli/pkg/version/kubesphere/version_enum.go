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

package kubesphere

import (
	"fmt"
	"strings"

	versionutil "k8s.io/apimachinery/pkg/util/version"
)

type Version int

const (
	V211 Version = iota
	V300
	V310
	V311
	V320
	V321
	V330
	V331
)

var VersionList = []Version{
	V211,
	V300,
	V310,
	V311,
	V320,
	V321,
	V330,
	V331,
}

var VersionMap = map[string]*KsInstaller{
	V211.String(): KsV211,
	V300.String(): KsV300,
	V310.String(): KsV310,
	V311.String(): KsV311,
	V320.String(): KsV320,
	V321.String(): KsV321,
	V330.String(): KsV330,
	V331.String(): KsV331,
}

var CNSource = map[string]bool{
	V310.String(): true,
	V311.String(): true,
	V320.String(): true,
	V321.String(): true,
	V330.String(): true,
	V331.String(): true,
}

func (v Version) String() string {
	switch v {
	case V211:
		return "v2.1.1"
	case V300:
		return "v3.0.0"
	case V310:
		return "v3.1.0"
	case V311:
		return "v3.1.1"
	case V320:
		return "v3.2.0"
	case V321:
		return "v3.2.1"
	case V330:
		return "v3.3.0"
	case V331:
		return "v3.3.1"
	default:
		return "invalid option"
	}
}

func VersionsStringArr() []string {
	strArr := make([]string, 0, len(VersionList))
	for i, v := range VersionList {
		strArr[i] = v.String()
	}
	return strArr
}

func StabledVersionSupport(version string) (*KsInstaller, bool) {
	if ks, ok := VersionMap[version]; ok {
		return ks, true
	}
	return nil, false
}

func LatestRelease(version string) (*KsInstaller, bool) {
	if strings.HasPrefix(version, "nightly-") ||
		version == "latest" ||
		version == "master" ||
		strings.Contains(version, "release") {
		return Latest(), true
	}

	v, err := versionutil.ParseGeneric(version)
	if err != nil {
		return nil, false
	}

	str := fmt.Sprintf("v%s", v.String())
	if ks, ok := StabledVersionSupport(str); ok {
		if ks.Version == Latest().Version {
			return ks, true
		}
		return nil, false
	}

	return nil, false
}

func DevRelease(version string) (*KsInstaller, bool) {
	if strings.HasPrefix(version, "nightly-") ||
		version == "latest" ||
		version == "master" ||
		strings.Contains(version, "release") {
		return Latest(), true
	}

	if _, ok := StabledVersionSupport(version); ok {
		return nil, false
	}

	v, err := versionutil.ParseGeneric(version)
	if err != nil {
		return nil, false
	}

	if ks, ok := StabledVersionSupport(fmt.Sprintf("v%s", v.String())); ok {
		return ks, true
	}

	return nil, false
}

func Latest() *KsInstaller {
	return VersionMap[VersionList[len(VersionList)-1].String()]
}
