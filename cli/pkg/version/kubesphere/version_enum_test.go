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
	"reflect"
	"testing"
)

func TestDevRelease(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    *KsInstaller
		ok      bool
	}{
		{
			name:    "test_v3.2.1-rc.1",
			version: "v3.2.1-rc.1",
			want:    KsV321,
			ok:      true,
		},
		{
			name:    "test_v3.2.1",
			version: "v3.2.1",
			want:    nil,
			ok:      false,
		},
		{
			name:    "test_v3.2.0",
			version: "v3.2.0",
			want:    nil,
			ok:      false,
		},
		{
			name:    "test_v3.2.0-alpha.1",
			version: "v3.2.0-alpha.1",
			want:    KsV320,
			ok:      true,
		},
		{
			name:    "test_v3.2.0-beta.1",
			version: "v3.2.0-beta.1",
			want:    KsV320,
			ok:      true,
		},
		{
			name:    "test_v3.1.0-alpha.1",
			version: "v3.1.0-alpha.1",
			want:    KsV310,
			ok:      true,
		},
		{
			name:    "test_latest",
			version: "latest",
			want:    KsV321,
			ok:      true,
		},
		{
			name:    "test_master",
			version: "master",
			want:    KsV321,
			ok:      true,
		},
		{
			name:    "test_release-3.2",
			version: "release-3.2",
			want:    KsV321,
			ok:      true,
		},
		{
			name:    "test_v1.2.0",
			version: "v1.2.0",
			want:    nil,
			ok:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := DevRelease(tt.version)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DevRelease() got = %v, want %v", got, tt.want)
			}
			if ok != tt.ok {
				t.Errorf("DevRelease() got1 = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func TestLatest(t *testing.T) {
	tests := []struct {
		name string
		want *KsInstaller
	}{
		{
			name: "test_latest",
			want: KsV321,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Latest(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Latest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLatestRelease(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    *KsInstaller
		ok      bool
	}{
		{
			name:    "test_latest",
			version: "latest",
			want:    KsV321,
			ok:      true,
		},
		{
			name:    "test_master",
			version: "master",
			want:    KsV321,
			ok:      true,
		},
		{
			name:    "test_release-3.2",
			version: "release-3.2",
			want:    KsV321,
			ok:      true,
		},
		{
			name:    "test_v3.2.1",
			version: "v3.2.1",
			want:    KsV321,
			ok:      true,
		},
		{
			name:    "test_v3.2.1-rc.1",
			version: "v3.2.1-rc.1",
			want:    KsV321,
			ok:      true,
		},
		{
			name:    "test_v3.2.0",
			version: "v3.2.0",
			want:    nil,
			ok:      false,
		},
		{
			name:    "test_v3.1.0",
			version: "v3.1.0",
			want:    nil,
			ok:      false,
		},
		{
			name:    "test_v3.2.0-alpha.1",
			version: "v3.2.0-alpha.1",
			want:    nil,
			ok:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := LatestRelease(tt.version)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LatestRelease() got = %v, want %v", got, tt.want)
			}
			if ok != tt.ok {
				t.Errorf("LatestRelease() got1 = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func TestStabledVersionSupport(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    *KsInstaller
		ok      bool
	}{
		{
			name:    "test_v3.2.1-rc.1",
			version: "v3.2.1-rc.1",
			want:    nil,
			ok:      false,
		},
		{
			name:    "test_v3.2.0",
			version: "v3.2.0",
			want:    KsV320,
			ok:      true,
		},
		{
			name:    "test_3.2.0",
			version: "3.2.0",
			want:    nil,
			ok:      false,
		},
		{
			name:    "test_v3.2.0-alpha.1",
			version: "v3.2.0-alpha.1",
			want:    nil,
			ok:      false,
		},
		{
			name:    "test_v1.2.0",
			version: "v1.2.0",
			want:    nil,
			ok:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := StabledVersionSupport(tt.version)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StabledVersionSupport() got = %v, want %v", got, tt.want)
			}
			if ok != tt.ok {
				t.Errorf("StabledVersionSupport() got1 = %v, want %v", ok, tt.ok)
			}
		})
	}
}
