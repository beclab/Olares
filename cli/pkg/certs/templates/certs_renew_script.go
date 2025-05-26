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

// K8sCertsRenewScript defines the template of k8s-certs-renew timer for systemd.
var K8sCertsRenewScript = template.Must(template.New("k8s-certs-renew.sh").Parse(
	dedent.Dedent(`#!/bin/bash
{{- if .IsKubeadmAlphaCerts }}
kubeadmCerts='/usr/local/bin/kubeadm alpha certs'
{{- else}}
kubeadmCerts='/usr/local/bin/kubeadm certs'
{{- end }}
getCertValidDays() {
  local earliestExpireDate; earliestExpireDate=$(${kubeadmCerts} check-expiration | grep -o "[A-Za-z]\{3,4\}\s\w\w,\s[0-9]\{4,\}\s\w*:\w*\s\w*\s*" | xargs -I {} date -d {} +%s | sort | head -n 1)
  local today; today="$(date +%s)"
  echo -n $(( ($earliestExpireDate - $today) / (24 * 60 * 60) ))
}
echo "## Expiration before renewal ##"
${kubeadmCerts} check-expiration
if [ $(getCertValidDays) -lt 30 ]; then
  echo "## Renewing certificates managed by kubeadm ##"
  ${kubeadmCerts} renew all
  echo "## Restarting control plane pods managed by kubeadm ##"
{{- if .IsDocker}}
  $(which docker | grep docker) ps -af 'name=k8s_POD_(kube-apiserver|kube-controller-manager|kube-scheduler|etcd)-*' -q | /usr/bin/xargs $(which docker | grep docker) rm -f
{{- else}}
  $(which crictl | grep crictl) pods --namespace kube-system --name 'kube-scheduler-*|kube-controller-manager-*|kube-apiserver-*|etcd-*' -q | /usr/bin/xargs $(which crictl | grep crictl) rmp -f
{{- end }}
  echo "## Updating /root/.kube/config ##"
  cp /etc/kubernetes/admin.conf /root/.kube/config && chmod 0600 /root/.kube/config
fi
echo "## Waiting for apiserver to be up again ##"
until printf "" 2>>/dev/null >>/dev/tcp/127.0.0.1/6443; do sleep 1; done
echo "## Expiration after renewal ##"
${kubeadmCerts} check-expiration
    `)))
