/*
 Copyright 2022 The KubeSphere Authors.
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

package plugins

import (
	"path/filepath"
	"text/template"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/images"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
)

// Kata Containers is an open source community working to build a secure container runtime with lightweight virtual
// machines that feel and perform like containers, but provide stronger workload isolation using hardware virtualization
// technology as a second layer of defense.

var (
	KataDeploy = template.Must(template.New("kata-deploy.yaml").Parse(
		dedent.Dedent(`---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kata-label-node
  namespace: kube-system
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: node-labeler
rules:
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get", "patch"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kata-label-node-rb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: node-labeler
subjects:
- kind: ServiceAccount
  name: kata-label-node
  namespace: kube-system
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: kata-deploy
  namespace: kube-system
spec:
  selector:
      matchLabels:
        name: kata-deploy
  template:
    metadata:
        labels:
          name: kata-deploy
    spec:
      serviceAccountName: kata-label-node
      containers:
      - name: kube-kata
        image: {{ .KataDeployImage }}
        imagePullPolicy: Always
        lifecycle:
          preStop:
            exec:
              command: ["bash", "-c", "/opt/kata-artifacts/scripts/kata-deploy.sh cleanup"]
        command: [ "bash", "-c", "/opt/kata-artifacts/scripts/kata-deploy.sh install" ]
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        securityContext:
          privileged: false
        volumeMounts:
        - name: crio-conf
          mountPath: /etc/crio/
        - name: containerd-conf
          mountPath: /etc/containerd/
        - name: kata-artifacts
          mountPath: /opt/kata/
        - name: dbus
          mountPath: /var/run/dbus
        - name: systemd
          mountPath: /run/systemd
        - name: local-bin
          mountPath: /usr/local/bin/
      volumes:
        - name: crio-conf
          hostPath:
            path: /etc/crio/
        - name: containerd-conf
          hostPath:
            path: /etc/containerd/
        - name: kata-artifacts
          hostPath:
            path: /opt/kata/
            type: DirectoryOrCreate
        - name: dbus
          hostPath:
            path: /var/run/dbus
        - name: systemd
          hostPath:
            path: /run/systemd
        - name: local-bin
          hostPath:
            path: /usr/local/bin/
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
---
kind: RuntimeClass
apiVersion: node.k8s.io/v1beta1
metadata:
    name: kata-qemu
handler: kata-qemu
overhead:
    podFixed:
        memory: "160Mi"
        cpu: "250m"
---
kind: RuntimeClass
apiVersion: node.k8s.io/v1beta1
metadata:
    name: kata-clh
handler: kata-clh
overhead:
    podFixed:
        memory: "130Mi"
        cpu: "250m"
---
kind: RuntimeClass
apiVersion: node.k8s.io/v1beta1
metadata:
    name: kata-fc
handler: kata-fc
overhead:
    podFixed:
        memory: "130Mi"
        cpu: "250m"
    `)))
)

func DeployKataTasks(d *DeployPluginsModule) []task.Interface {
	generateKataDeployManifests := &task.RemoteTask{
		Name:    "GenerateKataDeployManifests",
		Desc:    "Generate kata-deploy manifests",
		Hosts:   d.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action: &action.Template{
			Name:     "GenerateKataDeployManifests",
			Template: KataDeploy,
			Data: util.Data{
				"KataDeployImage": images.GetImage(d.Runtime, d.KubeConf, "kata-deploy").ImageName(),
			},
			Dst: filepath.Join(common.KubeAddonsDir, KataDeploy.Name()),
		},
		Parallel: false,
	}

	deployKata := &task.RemoteTask{
		Name:    "ApplyKataDeployManifests",
		Desc:    "Apply kata-deploy manifests",
		Hosts:   d.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(ApplyKataDeployManifests),
	}

	return []task.Interface{
		generateKataDeployManifests,
		deployKata,
	}
}

type ApplyKataDeployManifests struct {
	common.KubeAction
}

func (a *ApplyKataDeployManifests) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/addons/kata-deploy.yaml", true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "apply kata-deploy manifests failed")
	}
	return nil
}
