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

package precheck

const (
	// command software
	sudo       = "sudo"
	curl       = "curl"
	openssl    = "openssl"
	ebtables   = "ebtables"
	socat      = "socat"
	ipset      = "ipset"
	ipvsadm    = "ipvsadm"
	conntrack  = "conntrack"
	chrony     = "chronyd"
	docker     = "docker"
	containerd = "containerd"
	showmount  = "showmount"
	rbd        = "rbd"
	glusterfs  = "glusterfs"

	// extra command tools
	nfs  = "nfs"
	ceph = "ceph"

	UnknownVersion = "UnknownVersion"
)

// defines the base software to be checked.
var baseSoftware = []string{
	sudo,
	curl,
	openssl,
	ebtables,
	socat,
	ipset,
	ipvsadm,
	conntrack,
	chrony,
	docker,
	containerd,
	showmount,
	rbd,
	glusterfs,
}
