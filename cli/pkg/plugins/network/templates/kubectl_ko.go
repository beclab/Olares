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
	"text/template"

	"github.com/lithammer/dedent"
)

var KubectlKo = template.Must(template.New("kubectl-ko").Parse(
	dedent.Dedent(`#!/bin/bash
set -euo pipefail

KUBE_OVN_NS=kube-system
WITHOUT_KUBE_PROXY=false
OVN_NB_POD=
OVN_SB_POD=
KUBE_OVN_VERSION=
REGISTRY="kubeovn"

showHelp(){
  echo "kubectl ko {subcommand} [option...]"
  echo "Available Subcommands:"
  echo "  [nb|sb] [status|kick|backup|dbstatus|restore]     ovn-db operations show cluster status, kick stale server, backup database, get db consistency status or restore ovn nb db when met 'inconsistent data' error"
  echo "  nbctl [ovn-nbctl options ...]    invoke ovn-nbctl"
  echo "  sbctl [ovn-sbctl options ...]    invoke ovn-sbctl"
  echo "  vsctl {nodeName} [ovs-vsctl options ...]   invoke ovs-vsctl on the specified node"
  echo "  ofctl {nodeName} [ovs-ofctl options ...]   invoke ovs-ofctl on the specified node"
  echo "  dpctl {nodeName} [ovs-dpctl options ...]   invoke ovs-dpctl on the specified node"
  echo "  appctl {nodeName} [ovs-appctl options ...]   invoke ovs-appctl on the specified node"
  echo "  tcpdump {namespace/podname} [tcpdump options ...]     capture pod traffic"
  echo "  trace {namespace/podname} {target ip address} [target mac address] {icmp|tcp|udp} [target tcp or udp port]    trace ovn microflow of specific packet"
  echo "  diagnose {all|node} [nodename]    diagnose connectivity of all nodes or a specific node"
  echo "  tuning {install-fastpath|local-install-fastpath|remove-fastpath|install-stt|local-install-stt|remove-stt} {centos7|centos8}} [kernel-devel-version]  deploy  kernel optimisation components to the system"
  echo "  reload restart all kube-ovn components"
  echo "  env-check check the environment configuration"
}

# usage: ipv4_to_hex 192.168.0.1
ipv4_to_hex(){
  printf "%02x" ${1//./ }
}

# convert hex to dec (portable version)
hex2dec(){
	for i in $(echo "$@"); do
		printf "%d\n" "$(( 0x$i ))"
	done
}

# https://github.com/chmduquesne/wg-ip
# usage: expand_ipv6 2001::1
expand_ipv6(){
	local ip=$1

	# prepend 0 if we start with :
	echo $ip | grep -qs "^:" && ip="0${ip}"

	# expand ::
	if echo $ip | grep -qs "::"; then
		local colons=$(echo $ip | sed 's/[^:]//g')
		local missing=$(echo ":::::::::" | sed "s/$colons//")
		local expanded=$(echo $missing | sed 's/:/:0/g')
		ip=$(echo $ip | sed "s/::/$expanded/")
	fi

	local blocks=$(echo $ip | grep -o "[0-9a-f]\+")
	set $blocks

	printf "%04x:%04x:%04x:%04x:%04x:%04x:%04x:%04x\n" \
		$(hex2dec $@)
}

# convert an IPv6 address to bytes
ipv6_bytes(){
  for x in $(expand_ipv6 $1 | tr ':' ' '); do
    printf "%d %d " $((0x$x >> 8 & 0xff)) $((0x$x & 0xff))
  done
  echo
}

# usage: ipIsInCidr 192.168.0.1 192.168.0.0/24
# return: 0 for true, 1 for false
ipIsInCidr(){
  local ip=$1
  local cidr=$2

  if [[ $ip =~ .*:.* ]]; then
    # IPv6
    cidr=${cidr#*,}
    local network=${cidr%/*}
    local prefix=${cidr#*/}
    local ip_bytes=($(ipv6_bytes $ip))
    local network_bytes=($(ipv6_bytes $network))
    for ((i=0; i<${#ip_bytes[*]}; i++)); do
      if [ ${ip_bytes[$i]} -eq ${network_bytes[$i]} ]; then
        continue
      fi

      if [ $((($i+1)*8)) -le $prefix ]; then
        return 1
      fi
      if [ $(($i*8)) -ge $prefix ]; then
        return 0
      fi
      if [ $((($i+1)*8)) -le $prefix ]; then
        return 1
      fi

      local bits=$(($prefix-$i*8))
      local mask=$((0xff<<$bits & 0xff))
      # TODO: check whether the IP is network/broadcast address
      if [ $((${ip_bytes[$i]} & $mask)) -ne ${network_bytes[$i]} ]; then
        return 1
      fi
    done

    return 0
  fi

  # IPv4
  cidr=${cidr%,*}
  local network=${cidr%/*}
  local prefix=${cidr#*/}
  local ip_hex=$(ipv4_to_hex $ip)
  local ip_dec=$((0x$ip_hex))
  local network_hex=$(ipv4_to_hex $network)
  local network_dec=$((0x$network_hex))
  local broadcast_dec=$(($network_dec + 2**(32-$prefix) - 1))
  # TODO: check whether the IP is network/broadcast address
  if [ $ip_dec -gt $network_dec -a $ip_dec -lt $broadcast_dec ]; then
    return 0
  fi

  return 1
}

tcpdump(){
  namespacedPod="$1"; shift
  namespace=$(echo "$namespacedPod" | cut -d "/" -f1)
  podName=$(echo "$namespacedPod" | cut -d "/" -f2)
  if [ "$podName" = "$namespacedPod" ]; then
    namespace="default"
  fi

  nodeName=$(kubectl get pod "$podName" -n "$namespace" -o jsonpath={.spec.nodeName})
  hostNetwork=$(kubectl get pod "$podName" -n "$namespace" -o jsonpath={.spec.hostNetwork})

  if [ -z "$nodeName" ]; then
    echo "Pod $namespacedPod not exists on any node"
    exit 1
  fi

  ovnCni=$(kubectl get pod -n $KUBE_OVN_NS -l app=kube-ovn-cni -o 'jsonpath={.items[?(@.spec.nodeName=="'$nodeName'")].metadata.name}')
  if [ -z "$ovnCni" ]; then
    echo "kube-ovn-cni not exist on node $nodeName"
    exit 1
  fi

  if [ "$hostNetwork" = "true" ]; then
    set -x
    kubectl exec "$ovnCni" -n $KUBE_OVN_NS -- tcpdump -nn "$@"
  else
    nicName=$(kubectl exec "$ovnCni" -n $KUBE_OVN_NS -- ovs-vsctl --data=bare --no-heading --columns=name find interface external-ids:iface-id="$podName"."$namespace" | tr -d '\r')
    if [ -z "$nicName" ]; then
      echo "nic doesn't exist on node $nodeName"
      exit 1
    fi
    podNicType=$(kubectl get pod "$podName" -n "$namespace" -o jsonpath={.metadata.annotations.ovn\\.kubernetes\\.io/pod_nic_type})
    podNetNs=$(kubectl exec "$ovnCni" -n $KUBE_OVN_NS -- ovs-vsctl --data=bare --no-heading get interface "$nicName" external-ids:pod_netns | tr -d '\r' | sed -e 's/^"//' -e 's/"$//')
    set -x
    if [ "$podNicType" = "internal-port" ]; then
      kubectl exec "$ovnCni" -n $KUBE_OVN_NS -- nsenter --net="$podNetNs" tcpdump -nn -i "$nicName" "$@"
    else
      kubectl exec "$ovnCni" -n $KUBE_OVN_NS -- nsenter --net="$podNetNs" tcpdump -nn -i eth0 "$@"
    fi
  fi
}

trace(){
  namespacedPod="$1"
  namespace=$(echo "$namespacedPod" | cut -d "/" -f1)
  podName=$(echo "$namespacedPod" | cut -d "/" -f2)
  if [ "$podName" = "$namespacedPod" ]; then
    namespace="default"
  fi

  dst="$2"
  if [ -z "$dst" ]; then
    echo "need a target ip address"
    exit 1
  fi

  hostNetwork=$(kubectl get pod "$podName" -n "$namespace" -o jsonpath={.spec.hostNetwork})
  if [ "$hostNetwork" = "true" ]; then
    echo "Can not trace host network pod"
    exit 1
  fi

  af="4"
  nw="nw"
  proto=""
  if [[ "$dst" =~ .*:.* ]]; then
    af="6"
    nw="ipv6"
    proto="6"
  fi

  podIPs=($(kubectl get pod "$podName" -n "$namespace" -o jsonpath="{.status.podIPs[*].ip}"))
  if [ ${#podIPs[@]} -eq 0 ]; then
    podIPs=($(kubectl get pod "$podName" -n "$namespace" -o jsonpath={.metadata.annotations.ovn\\.kubernetes\\.io/ip_address} | sed 's/,/ /g'))
    if [ ${#podIPs[@]} -eq 0 ]; then
      echo "pod address not ready"
      exit 1
    fi
  fi

  podIP=""
  for ip in ${podIPs[@]}; do
    if [ "$af" = "4" ]; then
      if [[ ! "$ip" =~ .*:.* ]]; then
        podIP=$ip
        break
      fi
    elif [[ "$ip" =~ .*:.* ]]; then
      podIP=$ip
      break
    fi
  done

  if [ -z "$podIP" ]; then
    echo "Pod $namespacedPod has no IPv$af address"
    exit 1
  fi

  nodeName=$(kubectl get pod "$podName" -n "$namespace" -o jsonpath={.spec.nodeName})
  ovnCni=$(kubectl get pod -n $KUBE_OVN_NS -l app=kube-ovn-cni -o 'jsonpath={.items[?(@.spec.nodeName=="'$nodeName'")].metadata.name}')
  if [ -z "$ovnCni" ]; then
    echo "No kube-ovn-cni Pod running on node $nodeName"
    exit 1
  fi

  ls=$(kubectl get pod "$podName" -n "$namespace" -o jsonpath={.metadata.annotations.ovn\\.kubernetes\\.io/logical_switch})
  if [ -z "$ls" ]; then
    echo "pod address not ready"
    exit 1
  fi

  local cidr=$(kubectl get pod "$podName" -n "$namespace" -o jsonpath={.metadata.annotations.ovn\\.kubernetes\\.io/cidr})
  mac=$(kubectl get pod "$podName" -n "$namespace" -o jsonpath={.metadata.annotations.ovn\\.kubernetes\\.io/mac_address})

  dstMac=""
  if echo "$3" | grep -qE '^([[:xdigit:]]{1,2}:){5}[[:xdigit:]]{1,2}$'; then
    dstMac=$3
    shift
  elif ipIsInCidr $dst $cidr; then
    set +o pipefail
    if [ $af -eq 4 ]; then
      dstMac=$(kubectl exec $OVN_NB_POD -n $KUBE_OVN_NS -c ovn-central -- ovn-nbctl --data=bare --no-heading --columns=addresses list logical_switch_port | grep -w "$(echo $dst | tr . '\.')" | awk '{print $1}')
    else
      dstMac=$(kubectl exec $OVN_NB_POD -n $KUBE_OVN_NS -c ovn-central -- ovn-nbctl --data=bare --no-heading --columns=addresses list logical_switch_port | grep -i " $dst\$" | awk '{print $1}')
    fi
    set -o pipefail
  fi

  if [ -z "$dstMac" ]; then
    vlan=$(kubectl get subnet "$ls" -o jsonpath={.spec.vlan})
    logicalGateway=$(kubectl get subnet "$ls" -o jsonpath={.spec.logicalGateway})
    if [ ! -z "$vlan" -a "$logicalGateway" != "true" ]; then
      gateway=$(kubectl get subnet "$ls" -o jsonpath={.spec.gateway})
      if [[ "$gateway" =~ .*,.* ]]; then
        if [ "$af" = "4" ]; then
          gateway=${gateway%%,*}
        else
          gateway=${gateway##*,}
        fi
      fi

      nicName=$(kubectl exec "$ovnCni" -c cni-server -n $KUBE_OVN_NS -- ovs-vsctl --data=bare --no-heading --columns=name find interface external-ids:iface-id="$podName"."$namespace" | tr -d '\r')
      if [ -z "$nicName" ]; then
        echo "failed to find ovs interface for Pod namespacedPod on node $nodeName"
        exit 1
      fi

      podNicType=$(kubectl get pod "$podName" -n "$namespace" -o jsonpath={.metadata.annotations.ovn\\.kubernetes\\.io/pod_nic_type})
      podNetNs=$(kubectl exec "$ovnCni" -c cni-server -n $KUBE_OVN_NS -- ovs-vsctl --data=bare --no-heading get interface "$nicName" external-ids:pod_netns | tr -d '\r' | sed -e 's/^"//' -e 's/"$//')
      if [ "$podNicType" != "internal-port" ]; then
        interface=$(kubectl exec "$ovnCni" -c cni-server -n $KUBE_OVN_NS -- ovs-vsctl --format=csv --data=bare --no-heading --columns=name find interface external_id:iface-id="$podName"."$namespace")
        peer=$(kubectl exec "$ovnCni" -c cni-server -n $KUBE_OVN_NS -- ip link show $interface | grep -oE "^[0-9]+:\\s$interface@if[0-9]+" | awk -F @ '{print $2}')
        peerIndex=${peer//if/}
        peer=$(kubectl exec "$ovnCni" -c cni-server -n $KUBE_OVN_NS -- nsenter --net="$podNetNs" ip link show type veth | grep "^$peerIndex:" | awk -F @ '{print $1}')
        nicName=$(echo $peer | awk '{print $2}')
      fi

      set +o pipefail
      master=$(kubectl exec "$ovnCni" -c cni-server -n $KUBE_OVN_NS -- nsenter --net="$podNetNs" ip link show $nicName | grep -Eo '\smaster\s\w+\s' | awk '{print $2}')
      set -o pipefail
      if [ ! -z "$master" ]; then
        echo "Error: Pod nic $nicName is a slave of $master, please set the destination mac address."
        exit 1
      fi

      if [[ "$gateway" =~ .*:.* ]]; then
        cmd="ndisc6 -q $gateway $nicName"
        output=$(kubectl exec "$ovnCni" -c cni-server -n $KUBE_OVN_NS -- nsenter --net="$podNetNs" ndisc6 -q "$gateway" "$nicName")
      else
        cmd="arping -c3 -C1 -i1 -I $nicName $gateway"
        output=$(kubectl exec "$ovnCni" -c cni-server -n $KUBE_OVN_NS -- nsenter --net="$podNetNs" arping -c3 -C1 -i1 -I "$nicName" "$gateway")
      fi

      if [ $? -ne 0 ]; then
        echo "Error: failed to execute '$cmd' in Pod's netns"
        exit 1
      fi

      dstMac=$(echo "$output" | grep -oE '([[:xdigit:]]{1,2}:){5}[[:xdigit:]]{1,2}')
    fi
  fi

  if [ -z "$dstMac" ]; then
    echo "Using the gateway mac address as destination"
    lr=$(kubectl get pod "$podName" -n "$namespace" -o jsonpath={.metadata.annotations.ovn\\.kubernetes\\.io/logical_router})
    if [ -z "$lr" ]; then
      lr=$(kubectl get subnet "$ls" -o jsonpath={.spec.vpc})
    fi
    dstMac=$(kubectl exec $OVN_NB_POD -n $KUBE_OVN_NS -c ovn-central -- ovn-nbctl --data=bare --no-heading --columns=mac find logical_router_port name="$lr"-"$ls" | tr -d '\r')
  fi

  if [ -z "$dstMac" ]; then
    echo "failed to get destination mac"
    exit 1
  fi

  lsp="$podName.$namespace"
  lspUUID=$(kubectl exec $OVN_NB_POD -n $KUBE_OVN_NS -c ovn-central -- ovn-nbctl --data=bare --no-heading --columns=_uuid find logical_switch_port name="$lsp")
  if [ -z "$lspUUID" ]; then
    echo "Notice: LSP $lsp does not exist"
  fi
  vmOwner=$(kubectl get pod "$podName" -n "$namespace" -o jsonpath='{.metadata.ownerReferences[?(@.kind=="VirtualMachineInstance")].name}')
  if [ ! -z "$vmOwner" ]; then
    lsp="$vmOwner.$namespace"
  fi

  if [ -z "$lsp" ]; then
    echo "failed to get LSP of Pod $namespace/$podName"
    exit 1
  fi

  type="$3"
  case $type in
    icmp)
      set -x
      kubectl exec "$OVN_SB_POD" -n $KUBE_OVN_NS -c ovn-central -- ovn-trace --ct=new "$ls" "inport == \"$lsp\" && ip.ttl == 64 && icmp && eth.src == $mac && ip$af.src == $podIP && eth.dst == $dstMac && ip$af.dst == $dst"
      ;;
    tcp|udp)
      set -x
      kubectl exec "$OVN_SB_POD" -n $KUBE_OVN_NS -c ovn-central -- ovn-trace --ct=new "$ls" "inport == \"$lsp\" && ip.ttl == 64 && eth.src == $mac && ip$af.src == $podIP && eth.dst == $dstMac && ip$af.dst == $dst && $type.src == 10000 && $type.dst == $4"
      ;;
    *)
      echo "type $type not supported"
      echo "kubectl ko trace {namespace/podname} {target ip address} [target mac address] {icmp|tcp|udp} [target tcp or udp port]"
      exit 1
      ;;
  esac

  set +x
  echo "--------"
  echo "Start OVS Tracing"
  echo ""
  echo ""

  inPort=$(kubectl exec "$ovnCni" -c cni-server -n $KUBE_OVN_NS -- ovs-vsctl --format=csv --data=bare --no-heading --columns=ofport find interface external_id:iface-id="$podName"."$namespace")
  case $type in
    icmp)
      set -x
      kubectl exec "$ovnCni" -c cni-server -n $KUBE_OVN_NS -- ovs-appctl ofproto/trace br-int "in_port=$inPort,icmp$proto,nw_ttl=64,${nw}_src=$podIP,${nw}_dst=$dst,dl_src=$mac,dl_dst=$dstMac"
      ;;
    tcp|udp)
      set -x
      kubectl exec "$ovnCni" -c cni-server -n $KUBE_OVN_NS -- ovs-appctl ofproto/trace br-int "in_port=$inPort,$type$proto,nw_ttl=64,${nw}_src=$podIP,${nw}_dst=$dst,dl_src=$mac,dl_dst=$dstMac,${type}_src=1000,${type}_dst=$4"
      ;;
    *)
      echo "type $type not supported"
      echo "kubectl ko trace {namespace/podname} {target ip address} [target mac address] {icmp|tcp|udp} [target tcp or udp port]"
      exit 1
      ;;
  esac
}

xxctl(){
  subcommand="$1"; shift
  nodeName="$1"; shift
  kubectl get no "$nodeName" > /dev/null
  ovsPod=$(kubectl get pod -n $KUBE_OVN_NS -l app=ovs -o 'jsonpath={.items[?(@.spec.nodeName=="'$nodeName'")].metadata.name}')
  if [ -z "$ovsPod" ]; then
    echo "ovs pod  doesn't exist on node $nodeName"
    exit 1
  fi
  kubectl exec "$ovsPod" -n $KUBE_OVN_NS -- ovs-$subcommand "$@"
}

checkLeader(){
  component="$1"; shift
  set +o pipefail
  count=$(kubectl get ep ovn-$component -n $KUBE_OVN_NS -o yaml | grep ip | wc -l)
  set -o pipefail
  if [ $count -eq 0 ]; then
    echo "no ovn-$component exists !!"
    exit 1
  fi

  if [ $count -gt 1 ]; then
    echo "ovn-$component has more than one leader !!"
    exit 1
  fi

  echo "ovn-$component leader check ok"
}

diagnose(){
  kubectl get crd vpcs.kubeovn.io
  kubectl get crd vpc-nat-gateways.kubeovn.io
  kubectl get crd subnets.kubeovn.io
  kubectl get crd ips.kubeovn.io
  kubectl get crd vlans.kubeovn.io
  kubectl get crd provider-networks.kubeovn.io
  set +eu
  if ! kubectl get svc kube-dns -n kube-system ; then
     echo "Warning: kube-dns doesn't exist, maybe there is coredns service."
  fi
  set -eu
  kubectl get svc kubernetes -n default
  kubectl get sa -n kube-system ovn
  kubectl get clusterrole system:ovn
  kubectl get clusterrolebinding ovn

  kubectl get no -o wide
  kubectl ko nbctl show
  kubectl ko nbctl lr-policy-list ovn-cluster
  kubectl ko nbctl lr-route-list ovn-cluster
  kubectl ko nbctl ls-lb-list ovn-default
  kubectl ko nbctl list address_set
  kubectl ko nbctl list acl
  kubectl ko sbctl show

  if [ "${WITHOUT_KUBE_PROXY}" = "false" ]; then
    checkKubeProxy
  fi

  checkDeployment ovn-central
  checkDeployment kube-ovn-controller
  checkDaemonSet kube-ovn-cni
  checkDaemonSet ovs-ovn
  checkDeployment coredns

  checkLeader nb
  checkLeader sb
  checkLeader northd

  type="$1"
  case $type in
    all)
      echo "### kube-ovn-controller recent log"
      set +e
      kubectl logs -n $KUBE_OVN_NS -l app=kube-ovn-controller --tail=100 | grep E$(date +%m%d)
      set -e
      echo ""
      pingers=$(kubectl -n $KUBE_OVN_NS get po --no-headers -o custom-columns=NAME:.metadata.name -l app=kube-ovn-pinger)
      for pinger in $pingers
      do
        nodeName=$(kubectl get pod "$pinger" -n "$KUBE_OVN_NS" -o jsonpath={.spec.nodeName})
        echo "### start to diagnose node $nodeName"
        echo "#### ovn-controller log:"
        kubectl exec -n $KUBE_OVN_NS "$pinger" -- tail /var/log/ovn/ovn-controller.log
        echo ""
        echo "#### ovs-vswitchd log:"
        kubectl exec -n $KUBE_OVN_NS "$pinger" -- tail /var/log/openvswitch/ovs-vswitchd.log
        echo ""
        echo "#### ovs-vsctl show results:"
        kubectl exec -n $KUBE_OVN_NS "$pinger" -- ovs-vsctl show
        echo ""
        echo "#### pinger diagnose results:"
        kubectl exec -n $KUBE_OVN_NS "$pinger" -- /kube-ovn/kube-ovn-pinger --mode=job
        echo "### finish diagnose node $nodeName"
        echo ""
      done
      ;;
    node)
      nodeName="$2"
      kubectl get no "$nodeName" > /dev/null
      pinger=$(kubectl -n $KUBE_OVN_NS get po -l app=kube-ovn-pinger -o 'jsonpath={.items[?(@.spec.nodeName=="'$nodeName'")].metadata.name}')
      if [ ! -n "$pinger" ]; then
        echo "Error: No kube-ovn-pinger running on node $nodeName"
        exit 1
      fi
      echo "### start to diagnose node $nodeName"
      echo "#### ovn-controller log:"
      kubectl exec -n $KUBE_OVN_NS "$pinger" -- tail /var/log/ovn/ovn-controller.log
      echo ""
      echo "#### ovs-vswitchd log:"
      kubectl exec -n $KUBE_OVN_NS "$pinger" -- tail /var/log/openvswitch/ovs-vswitchd.log
      echo ""
      kubectl exec -n $KUBE_OVN_NS "$pinger" -- /kube-ovn/kube-ovn-pinger --mode=job
      echo "### finish diagnose node $nodeName"
      echo ""
      ;;
    *)
      echo "type $type not supported"
      echo "kubectl ko diagnose {all|node} [nodename]"
      ;;
    esac
}

getOvnCentralPod(){
    NB_POD=$(kubectl get pod -n $KUBE_OVN_NS -l ovn-nb-leader=true | grep ovn-central | head -n 1 | awk '{print $1}')
    if [ -z "$NB_POD" ]; then
      echo "nb leader not exists"
      exit 1
    fi
    OVN_NB_POD=$NB_POD
    SB_POD=$(kubectl get pod -n $KUBE_OVN_NS -l ovn-sb-leader=true | grep ovn-central | head -n 1 | awk '{print $1}')
    if [ -z "$SB_POD" ]; then
      echo "nb leader not exists"
      exit 1
    fi
    OVN_SB_POD=$SB_POD
    VERSION=$(kubectl  -n kube-system get pods -l ovn-sb-leader=true -o yaml | grep  "image: $REGISTRY/kube-ovn:" | head -n 1 | awk -F ':' '{print $3}')
    if [ -z "$VERSION" ]; then
          echo "kubeovn version not exists"
          exit 1
    fi
    KUBE_OVN_VERSION=$VERSION
}

checkDaemonSet(){
  name="$1"
  currentScheduled=$(kubectl get ds -n $KUBE_OVN_NS "$name" -o jsonpath={.status.currentNumberScheduled})
  desiredScheduled=$(kubectl get ds -n $KUBE_OVN_NS "$name" -o jsonpath={.status.desiredNumberScheduled})
  available=$(kubectl get ds -n $KUBE_OVN_NS "$name" -o jsonpath={.status.numberAvailable})
  ready=$(kubectl get ds -n $KUBE_OVN_NS "$name" -o jsonpath={.status.numberReady})
  if [ "$currentScheduled" = "$desiredScheduled" ] && [ "$desiredScheduled" = "$available" ] && [ "$available" = "$ready" ]; then
    echo "ds $name ready"
  else
    echo "Error ds $name not ready"
    exit 1
  fi
}

checkDeployment(){
  name="$1"
  ready=$(kubectl get deployment -n $KUBE_OVN_NS "$name" -o jsonpath={.status.readyReplicas})
  updated=$(kubectl get deployment -n $KUBE_OVN_NS "$name" -o jsonpath={.status.updatedReplicas})
  desire=$(kubectl get deployment -n $KUBE_OVN_NS "$name" -o jsonpath={.status.replicas})
  available=$(kubectl get deployment -n $KUBE_OVN_NS "$name" -o jsonpath={.status.availableReplicas})
  if [ "$ready" = "$updated" ] && [ "$updated" = "$desire" ] && [ "$desire" = "$available" ]; then
    echo "deployment $name ready"
  else
    echo "Error deployment $name not ready"
    exit 1
  fi
}

checkKubeProxy(){
  if kubectl get ds -n kube-system --no-headers -o custom-columns=NAME:.metadata.name | grep '^kube-proxy$' >/dev/null; then
    checkDaemonSet kube-proxy
  else
    for node in $(kubectl get node --no-headers -o custom-columns=NAME:.metadata.name); do
      local pod=$(kubectl get pod -n $KUBE_OVN_NS -l app=kube-ovn-cni -o 'jsonpath={.items[?(@.spec.nodeName=="'$node'")].metadata.name}')
      local ip=$(kubectl get pod -n $KUBE_OVN_NS -l app=kube-ovn-cni -o 'jsonpath={.items[?(@.spec.nodeName=="'$node'")].status.podIP}')
      local arg=""
      if [[ $ip =~ .*:.* ]]; then
        arg="g6"
        ip="[$ip]"
      fi
      healthResult=$(kubectl -n $KUBE_OVN_NS exec $pod -- curl -s${arg} -m 3 -w %{http_code} http://$ip:10256/healthz -o /dev/null | grep -v 200 || true)
      if [ -n "$healthResult" ]; then
        echo "$node kube-proxy's health check failed"
        exit 1
      fi
    done
  fi
  echo "kube-proxy ready"
}

dbtool(){
  suffix=$(date +%m%d%H%M%s)
  component="$1"; shift
  action="$1"; shift
  case $component in
    nb)
      case $action in
        status)
          kubectl exec "$OVN_NB_POD" -n $KUBE_OVN_NS -c ovn-central -- ovs-appctl -t /var/run/ovn/ovnnb_db.ctl cluster/status OVN_Northbound
          kubectl exec "$OVN_NB_POD" -n $KUBE_OVN_NS -c ovn-central -- ovs-appctl -t /var/run/ovn/ovnnb_db.ctl ovsdb-server/get-db-storage-status OVN_Northbound
          ;;
        kick)
          kubectl exec "$OVN_NB_POD" -n $KUBE_OVN_NS -c ovn-central -- ovs-appctl -t /var/run/ovn/ovnnb_db.ctl cluster/kick OVN_Northbound "$1"
          ;;
        backup)
          kubectl exec "$OVN_NB_POD" -n $KUBE_OVN_NS -c ovn-central -- ovsdb-tool cluster-to-standalone /etc/ovn/ovnnb_db.$suffix.backup /etc/ovn/ovnnb_db.db
          kubectl cp $KUBE_OVN_NS/$OVN_NB_POD:/etc/ovn/ovnnb_db.$suffix.backup $(pwd)/ovnnb_db.$suffix.backup
          kubectl exec "$OVN_NB_POD" -n $KUBE_OVN_NS -c ovn-central -- rm -f /etc/ovn/ovnnb_db.$suffix.backup
          echo "backup ovn-$component db to $(pwd)/ovnnb_db.$suffix.backup"
          ;;
        dbstatus)
          kubectl exec "$OVN_NB_POD" -n $KUBE_OVN_NS -c ovn-central -- ovn-appctl -t /var/run/ovn/ovnnb_db.ctl ovsdb-server/get-db-storage-status OVN_Northbound
          ;;
        restore)
          # set ovn-central replicas to 0
          replicas=$(kubectl get deployment -n $KUBE_OVN_NS ovn-central -o jsonpath={.spec.replicas})
          kubectl scale deployment -n $KUBE_OVN_NS ovn-central --replicas=0
          echo "ovn-central original replicas is $replicas"

          # backup ovn-nb db
          declare nodeIpArray
          declare podNameArray
          declare nodeIps

          if [[ $(kubectl get deployment -n kube-system ovn-central -o jsonpath='{.spec.template.spec.containers[0].env[1]}') =~ "NODE_IPS" ]]; then
            nodeIpVals=$(kubectl get deployment -n kube-system ovn-central -o jsonpath='{.spec.template.spec.containers[0].env[1].value}')
            nodeIps=(${nodeIpVals//,/ })
          else
            nodeIps=$(kubectl get node -lkube-ovn/role=master -o wide | grep -v "INTERNAL-IP" | awk '{print $6}')
          fi
          firstIP=${nodeIps[0]}
          podNames=$(kubectl get pod -n $KUBE_OVN_NS | grep ovs-ovn | awk '{print $1}')
          echo "first nodeIP is $firstIP"

          i=0
          for nodeIp in ${nodeIps[@]}
          do
            for pod in $podNames
            do
              hostip=$(kubectl get pod -n $KUBE_OVN_NS $pod -o jsonpath={.status.hostIP})
              if [ $nodeIp = $hostip ]; then
                nodeIpArray[$i]=$nodeIp
                podNameArray[$i]=$pod
                i=$(expr $i + 1)
                echo "ovs-ovn pod on node $nodeIp is $pod"
                break
              fi
            done
          done

          echo "backup nb db file"
          kubectl exec -it -n $KUBE_OVN_NS ${podNameArray[0]} -- ovsdb-tool cluster-to-standalone  /etc/ovn/ovnnb_db_standalone.db  /etc/ovn/ovnnb_db.db

          # mv all db files
          for pod in ${podNameArray[@]}
          do
            kubectl exec -it -n $KUBE_OVN_NS $pod -- mv /etc/ovn/ovnnb_db.db /tmp
            kubectl exec -it -n $KUBE_OVN_NS $pod -- mv /etc/ovn/ovnsb_db.db /tmp
          done

          # restore db and replicas
          echo "restore nb db file, operate in pod ${podNameArray[0]}"
          kubectl exec -it -n $KUBE_OVN_NS ${podNameArray[0]} -- mv /etc/ovn/ovnnb_db_standalone.db /etc/ovn/ovnnb_db.db
          kubectl scale deployment -n $KUBE_OVN_NS ovn-central --replicas=$replicas
          echo "finish restore nb db file and ovn-central replicas"

          echo "recreate ovs-ovn pods"
          kubectl delete pod -n $KUBE_OVN_NS -l app=ovs
          ;;
        *)
          echo "unknown action $action"
      esac
      ;;
    sb)
      case $action in
        status)
          kubectl exec "$OVN_SB_POD" -n $KUBE_OVN_NS -c ovn-central -- ovs-appctl -t /var/run/ovn/ovnsb_db.ctl cluster/status OVN_Southbound
          kubectl exec "$OVN_SB_POD" -n $KUBE_OVN_NS -c ovn-central -- ovs-appctl -t /var/run/ovn/ovnsb_db.ctl ovsdb-server/get-db-storage-status OVN_Southbound
          ;;
        kick)
          kubectl exec "$OVN_SB_POD" -n $KUBE_OVN_NS -c ovn-central -- ovs-appctl -t /var/run/ovn/ovnsb_db.ctl cluster/kick OVN_Southbound "$1"
          ;;
        backup)
          kubectl exec "$OVN_SB_POD" -n $KUBE_OVN_NS -c ovn-central -- ovsdb-tool cluster-to-standalone /etc/ovn/ovnsb_db.$suffix.backup /etc/ovn/ovnsb_db.db
          kubectl cp $KUBE_OVN_NS/$OVN_SB_POD:/etc/ovn/ovnsb_db.$suffix.backup $(pwd)/ovnsb_db.$suffix.backup
          kubectl exec "$OVN_SB_POD" -n $KUBE_OVN_NS -c ovn-central -- rm -f /etc/ovn/ovnsb_db.$suffix.backup
          echo "backup ovn-$component db to $(pwd)/ovnsb_db.$suffix.backup"
          ;;
        dbstatus)
          kubectl exec "$OVN_NB_POD" -n $KUBE_OVN_NS -c ovn-central -- ovn-appctl -t /var/run/ovn/ovnsb_db.ctl ovsdb-server/get-db-storage-status OVN_Southbound
          ;;
        restore)
          echo "restore cmd is only used for nb db"
          ;;
        *)
          echo "unknown action $action"
      esac
      ;;
    *)
      echo "unknown subcommand $component"
  esac
}

tuning(){
  action="$1"; shift
  sys="$1"; shift
  case $action in
    install-fastpath)
      case $sys in
        centos7)
          docker run -it --privileged -v /lib/modules:/lib/modules -v /usr/src:/usr/src -v /tmp/:/tmp/ $REGISTRY/centos7-compile:"$KUBE_OVN_VERSION" bash -c "./module.sh  centos install"
          while [ ! -f /tmp/kube_ovn_fastpath.ko ];
          do
            sleep 1
          done
          for i in $(kubectl -n kube-system get pods | grep ovn-cni | awk '{print $1}');
          do
            kubectl cp /tmp/kube_ovn_fastpath.ko kube-system/"$i":/tmp/
          done
          ;;
        centos8)
          docker run -it --privileged -v /lib/modules:/lib/modules -v /usr/src:/usr/src -v /tmp/:/tmp/ $REGISTRY/centos8-compile:"$KUBE_OVN_VERSION" bash -c "./module.sh  centos install"
          while [ ! -f /tmp/kube_ovn_fastpath.ko ];
          do
            sleep 1
          done
          for i in $(kubectl -n kube-system get pods | grep ovn-cni | awk '{print $1}');
          do
            kubectl cp /tmp/kube_ovn_fastpath.ko kube-system/"$i":/tmp/
          done
          ;;
        *)
          echo "unknown system $sys"
      esac
      ;;
    local-install-fastpath)
      case $sys in
        centos7)
          # shellcheck disable=SC2145
          docker run -it --privileged -v /lib/modules:/lib/modules -v /usr/src:/usr/src -v /tmp:/tmp $REGISTRY/centos7-compile:"$KUBE_OVN_VERSION" bash -c "./module.sh centos local-install $@"
          for i in $(kubectl -n kube-system get pods | grep ovn-cni | awk '{print $1}');
          do
            kubectl cp /tmp/kube_ovn_fastpath.ko kube-system/"$i":/tmp/
          done
          ;;
        centos8)
          # shellcheck disable=SC2145
          docker run -it --privileged -v /lib/modules:/lib/modules -v /usr/src:/usr/src -v /tmp:/tmp $REGISTRY/centos8-compile:"$KUBE_OVN_VERSION" bash -c "./module.sh centos local-install $@"
          for i in $(kubectl -n kube-system get pods | grep ovn-cni | awk '{print $1}');
          do
            kubectl cp /tmp/kube_ovn_fastpath.ko kube-system/"$i":/tmp/
          done
          ;;
        *)
          echo "unknown system $sys"
      esac
      ;;
    remove-fastpath)
      case $sys in
        centos)
          for i in $(kubectl -n kube-system get pods | grep ovn-cni | awk '{print $1}');
          do
            kubectl -n kube-system exec "$i" -- rm -f /tmp/kube_ovn_fastpath.ko
          done
          ;;
        *)
          echo "unknown system $sys"
      esac
      ;;
    install-stt)
      case $sys in
        centos7)
          # shellcheck disable=SC2145
          docker run -it --privileged -v /lib/modules:/lib/modules -v /usr/src:/usr/src -v /tmp:/tmp $REGISTRY/centos7-compile:"$KUBE_OVN_VERSION" bash -c "./module.sh stt install"
          for i in $(kubectl -n kube-system get pods | grep ovn-cni | awk '{print $1}');
          do
            for k in /tmp/*.rpm; do
              kubectl cp "$k" kube-system/"$i":/tmp/
            done
          done
          ;;
        centos8)
          # shellcheck disable=SC2145
          docker run -it --privileged -v /lib/modules:/lib/modules -v /usr/src:/usr/src -v /tmp:/tmp $REGISTRY/centos8-compile:"$KUBE_OVN_VERSION" bash -c "./module.sh stt install"
          for i in $(kubectl -n kube-system get pods | grep ovn-cni | awk '{print $1}');
          do
            for k in /tmp/*.rpm; do
              kubectl cp "$k" kube-system/"$i":/tmp/
            done
          done
          ;;
        *)
          echo "unknown system $sys"
      esac
      ;;
    local-install-stt)
      case $sys in
        centos7)
          # shellcheck disable=SC2145
          docker run -it --privileged -v /lib/modules:/lib/modules -v /usr/src:/usr/src -v /tmp:/tmp $REGISTRY/centos7-compile:"$KUBE_OVN_VERSION" bash -c "./module.sh stt local-install $@"
          for i in $(kubectl -n kube-system get pods | grep ovn-cni | awk '{print $1}');
          do
            for k in /tmp/*.rpm; do
              kubectl cp "$k" kube-system/"$i":/tmp/
            done
          done
          ;;
        centos8)
          # shellcheck disable=SC2145
          docker run -it --privileged -v /lib/modules:/lib/modules -v /usr/src:/usr/src -v /tmp:/tmp $REGISTRY/centos8-compile:"$KUBE_OVN_VERSION" bash -c "./module.sh stt local-install $@"
          for i in $(kubectl -n kube-system get pods | grep ovn-cni | awk '{print $1}');
          do
            for k in /tmp/*.rpm; do
              kubectl cp "$k" kube-system/"$i":/tmp/
            done
          done
          ;;
        *)
          echo "unknown system $sys"
      esac
      ;;
    remove-stt)
      case $sys in
        centos)
          for i in $(kubectl -n kube-system get pods | grep ovn-cni | awk '{print $1}');
          do
            kubectl -n kube-system exec "$i" -- rm -f /tmp/openvswitch-kmod*.rpm
          done
          ;;
        *)
          echo "unknown system $sys"
      esac
      ;;
    *)
      echo "unknown action $action"
  esac
}

reload(){
  kubectl delete pod -n kube-system -l app=ovn-central
  kubectl rollout status deployment/ovn-central -n kube-system
  kubectl delete pod -n kube-system -l app=ovs
  kubectl delete pod -n kube-system -l app=kube-ovn-controller
  kubectl rollout status deployment/kube-ovn-controller -n kube-system
  kubectl delete pod -n kube-system -l app=kube-ovn-cni
  kubectl rollout status daemonset/kube-ovn-cni -n kube-system
  kubectl delete pod -n kube-system -l app=kube-ovn-pinger
  kubectl rollout status daemonset/kube-ovn-pinger -n kube-system
  kubectl delete pod -n kube-system -l app=kube-ovn-monitor
  kubectl rollout status deployment/kube-ovn-monitor -n kube-system
}

env-check(){
  set +e

  KUBE_OVN_NS=kube-system
  podNames=$(kubectl get pod --no-headers -n $KUBE_OVN_NS | grep kube-ovn-cni | awk '{print $1}')
  for pod in $podNames
  do
    nodeName=$(kubectl get pod $pod -n $KUBE_OVN_NS -o jsonpath={.spec.nodeName})
    echo "************************************************"
    echo "Start environment check for Node $nodeName"
    echo "************************************************"
    kubectl exec -it -n $KUBE_OVN_NS $pod -c cni-server -- bash /kube-ovn/env-check.sh
  done
}

if [ $# -lt 1 ]; then
  showHelp
  exit 0
else
  subcommand="$1"; shift
fi

getOvnCentralPod

case $subcommand in
  nbctl)
    kubectl exec "$OVN_NB_POD" -n $KUBE_OVN_NS -c ovn-central -- ovn-nbctl "$@"
    ;;
  sbctl)
    kubectl exec "$OVN_SB_POD" -n $KUBE_OVN_NS -c ovn-central -- ovn-sbctl "$@"
    ;;
  vsctl|ofctl|dpctl|appctl)
    xxctl "$subcommand" "$@"
    ;;
  nb|sb)
    dbtool "$subcommand" "$@"
    ;;
  tcpdump)
    tcpdump "$@"
    ;;
  trace)
    trace "$@"
    ;;
  diagnose)
    diagnose "$@"
    ;;
  reload)
    reload
    ;;
  tuning)
    tuning "$@"
    ;;
  env-check)
    env-check
    ;;
  *)
    showHelp
    exit 1
    ;;
esac
`)))
