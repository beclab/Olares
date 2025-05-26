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

package templates

import (
	"fmt"
	"text/template"

	"bytetrade.io/web3os/installer/pkg/bootstrap/registry"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"github.com/lithammer/dedent"
)

var InitOsScriptTmpl = template.Must(template.New("initOS.sh").Parse(
	dedent.Dedent(`#!/usr/bin/env bash

# Copyright 2020 The KubeSphere Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# See https://github.com/kubernetes/website/issues/14457
if [ -f /etc/selinux/config ]; then 
  sed -ri 's/SELINUX=enforcing/SELINUX=disabled/' /etc/selinux/config
fi
# for ubuntu: sudo apt install selinux-utils
# for centos: yum install selinux-policy
if command -v setenforce &> /dev/null
then
  setenforce 0
  getenforce
fi

echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.conf
echo 'net.bridge.bridge-nf-call-arptables = 1' >> /etc/sysctl.conf
echo 'net.bridge.bridge-nf-call-ip6tables = 1' >> /etc/sysctl.conf
echo 'net.bridge.bridge-nf-call-iptables = 1' >> /etc/sysctl.conf
echo 'net.ipv4.ip_local_reserved_ports = 30000-32767' >> /etc/sysctl.conf
echo 'vm.max_map_count = 262144' >> /etc/sysctl.conf
echo 'fs.inotify.max_user_instances = 524288' >> /etc/sysctl.conf
echo 'kernel.pid_max = 65535' >> /etc/sysctl.conf

echo 'net.ipv4.tcp_tw_reuse = 1' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_tw_buckets = 32768' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_timestamps = 0' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_syncookies = 1' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_keepalive_time = 1800' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_keepalive_probes = 3' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_keepalive_intvl = 15' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_fin_timeout = 10' >> /etc/sysctl.conf
echo 'vm.overcommit_memory = 1' >> /etc/sysctl.conf
echo 'net.core.somaxconn = 10240' >> /etc/sysctl.conf


#See https://imroc.io/posts/kubernetes/troubleshooting-with-kubernetes-network/
sed -r -i "s@#{0,}?net.ipv4.tcp_tw_recycle ?= ?(0|1)@net.ipv4.tcp_tw_recycle = 0@g" /etc/sysctl.conf

sed -r -i  "s@#{0,}?net.ipv4.ip_forward ?= ?(0|1)@net.ipv4.ip_forward = 1@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.bridge.bridge-nf-call-arptables ?= ?(0|1)@net.bridge.bridge-nf-call-arptables = 1@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.bridge.bridge-nf-call-ip6tables ?= ?(0|1)@net.bridge.bridge-nf-call-ip6tables = 1@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.bridge.bridge-nf-call-iptables ?= ?(0|1)@net.bridge.bridge-nf-call-iptables = 1@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.ip_local_reserved_ports ?= ?([0-9]{1,}-{0,1},{0,1}){1,}@net.ipv4.ip_local_reserved_ports = 30000-32767@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?vm.max_map_count ?= ?([0-9]{1,})@vm.max_map_count = 262144@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?fs.inotify.max_user_instances ?= ?([0-9]{1,})@fs.inotify.max_user_instances = 524288@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?kernel.pid_max ?= ?([0-9]{1,})@kernel.pid_max = 65535@g" /etc/sysctl.conf

tmpfile="$$.tmp"
awk ' !x[$0]++{print > "'$tmpfile'"}' /etc/sysctl.conf
mv $tmpfile /etc/sysctl.conf

systemctl stop firewalld 1>/dev/null 2>/dev/null
systemctl disable firewalld 1>/dev/null 2>/dev/null
systemctl stop ufw 1>/dev/null 2>/dev/null
systemctl disable ufw 1>/dev/null 2>/dev/null

modinfo br_netfilter > /dev/null 2>&1
if [ $? -eq 0 ]; then
   modprobe br_netfilter
   mkdir -p /etc/modules-load.d
   echo 'br_netfilter' > /etc/modules-load.d/kubekey-br_netfilter.conf
fi

modinfo overlay > /dev/null 2>&1
if [ $? -eq 0 ]; then
   modprobe overlay
   echo 'overlay' >> /etc/modules-load.d/kubekey-br_netfilter.conf
fi

modprobe ip_vs
modprobe ip_vs_rr
modprobe ip_vs_wrr
modprobe ip_vs_sh

cat > /etc/modules-load.d/kube_proxy-ipvs.conf << EOF
ip_vs
ip_vs_rr
ip_vs_wrr
ip_vs_sh
EOF

modprobe nf_conntrack_ipv4 1>/dev/null 2>/dev/null
if [ $? -eq 0 ]; then
   echo 'nf_conntrack_ipv4' > /etc/modules-load.d/kube_proxy-ipvs.conf
else
   modprobe nf_conntrack
   echo 'nf_conntrack' > /etc/modules-load.d/kube_proxy-ipvs.conf
fi
sysctl -p

sed -i ':a;$!{N;ba};s@# kubekey hosts BEGIN.*# kubekey hosts END@@' /etc/hosts
sed -i '/^$/N;/\n$/N;//D' /etc/hosts

cat >>/etc/hosts<<EOF
# kubekey hosts BEGIN
{{- range .Hosts }}
{{ . }}
{{- end }}
# kubekey hosts END
EOF

echo 3 > /proc/sys/vm/drop_caches

# Make sure the iptables utility doesn't use the nftables backend.
update-alternatives --set iptables /usr/sbin/iptables-legacy >/dev/null 2>&1 || true
update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy >/dev/null 2>&1 || true
update-alternatives --set arptables /usr/sbin/arptables-legacy >/dev/null 2>&1 || true
update-alternatives --set ebtables /usr/sbin/ebtables-legacy >/dev/null 2>&1 || true

for i in $(systemctl list-unit-files --no-legend --no-pager -l | grep --color=never -o .*.slice | grep kubepod); do sudo systemctl stop $i; done

ulimit -u 65535
ulimit -n 65535
`)))

func GenerateHosts(runtime connector.ModuleRuntime, kubeConf *common.KubeConf) []string {
	var lbHost string
	var hostsList []string

	if kubeConf.Cluster.ControlPlaneEndpoint.Address != "" {
		lbHost = fmt.Sprintf("%s  %s", kubeConf.Cluster.ControlPlaneEndpoint.Address, kubeConf.Cluster.ControlPlaneEndpoint.Domain)
	} else {
		lbHost = fmt.Sprintf("%s  %s", runtime.GetHostsByRole(common.Master)[0].GetInternalAddress(), kubeConf.Cluster.ControlPlaneEndpoint.Domain)
	}

	for _, host := range runtime.GetAllHosts() {
		if host.GetName() != "" {
			hostsList = append(hostsList, fmt.Sprintf("%s  %s.%s %s",
				host.GetInternalAddress(),
				host.GetName(),
				kubeConf.Cluster.Kubernetes.ClusterName,
				host.GetName()))
		}
	}

	if len(runtime.GetHostsByRole(common.Registry)) > 0 {
		hostsList = append(hostsList, fmt.Sprintf("%s  %s", runtime.GetHostsByRole(common.Registry)[0].GetInternalAddress(), registry.RegistryCertificateBaseName))
	}

	hostsList = append(hostsList, lbHost)
	return hostsList
}
