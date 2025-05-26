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

// NodeFeatureDiscovery detects hardware features available on each node in a Kubernetes cluster, and advertises those
// features using node labels.

var (
	NodeFeatureDiscovery = template.Must(template.New("node-feature-discovery.yaml").Parse(
		dedent.Dedent(`---
apiVersion: v1
kind: Namespace
metadata:
  name: node-feature-discovery
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.7.0
  creationTimestamp: null
  name: nodefeaturerules.nfd.k8s-sigs.io
spec:
  group: nfd.k8s-sigs.io
  names:
    kind: NodeFeatureRule
    listKind: NodeFeatureRuleList
    plural: nodefeaturerules
    singular: nodefeaturerule
  scope: Cluster
  versions:
  - name: v1alpha1
    schema:
      openAPIV3Schema:
        description: NodeFeatureRule resource specifies a configuration for feature-based customization of node objects, such as node labeling.
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: NodeFeatureRuleSpec describes a NodeFeatureRule.
            properties:
              rules:
                description: Rules is a list of node customization rules.
                items:
                  description: Rule defines a rule for node customization such as labeling.
                  properties:
                    labels:
                      additionalProperties:
                        type: string
                      description: Labels to create if the rule matches.
                      type: object
                    labelsTemplate:
                      description: LabelsTemplate specifies a template to expand for dynamically generating multiple labels. Data (after template expansion) must be keys with an optional value (<key>[=<value>]) separated by newlines.
                      type: string
                    matchAny:
                      description: MatchAny specifies a list of matchers one of which must match.
                      items:
                        description: MatchAnyElem specifies one sub-matcher of MatchAny.
                        properties:
                          matchFeatures:
                            description: MatchFeatures specifies a set of matcher terms all of which must match.
                            items:
                              description: FeatureMatcherTerm defines requirements against one feature set. All requirements (specified as MatchExpressions) are evaluated against each element in the feature set.
                              properties:
                                feature:
                                  type: string
                                matchExpressions:
                                  additionalProperties:
                                    description: "MatchExpression specifies an expression to evaluate against a set of input values. It contains an operator that is applied when matching the input and an array of values that the operator evaluates the input against. \n NB: CreateMatchExpression or MustCreateMatchExpression() should be used for     creating new instances. NB: Validate() must be called if Op or Value fields are modified or if a new     instance is created from scratch without using the helper functions."
                                    properties:
                                      op:
                                        description: Op is the operator to be applied.
                                        enum:
                                        - In
                                        - NotIn
                                        - InRegexp
                                        - Exists
                                        - DoesNotExist
                                        - Gt
                                        - Lt
                                        - GtLt
                                        - IsTrue
                                        - IsFalse
                                        type: string
                                      value:
                                        description: Value is the list of values that the operand evaluates the input against. Value should be empty if the operator is Exists, DoesNotExist, IsTrue or IsFalse. Value should contain exactly one element if the operator is Gt or Lt and exactly two elements if the operator is GtLt. In other cases Value should contain at least one element.
                                        items:
                                          type: string
                                        type: array
                                    required:
                                    - op
                                    type: object
                                  description: MatchExpressionSet contains a set of MatchExpressions, each of which is evaluated against a set of input values.
                                  type: object
                              required:
                              - feature
                              - matchExpressions
                              type: object
                            type: array
                        required:
                        - matchFeatures
                        type: object
                      type: array
                    matchFeatures:
                      description: MatchFeatures specifies a set of matcher terms all of which must match.
                      items:
                        description: FeatureMatcherTerm defines requirements against one feature set. All requirements (specified as MatchExpressions) are evaluated against each element in the feature set.
                        properties:
                          feature:
                            type: string
                          matchExpressions:
                            additionalProperties:
                              description: "MatchExpression specifies an expression to evaluate against a set of input values. It contains an operator that is applied when matching the input and an array of values that the operator evaluates the input against. \n NB: CreateMatchExpression or MustCreateMatchExpression() should be used for     creating new instances. NB: Validate() must be called if Op or Value fields are modified or if a new     instance is created from scratch without using the helper functions."
                              properties:
                                op:
                                  description: Op is the operator to be applied.
                                  enum:
                                  - In
                                  - NotIn
                                  - InRegexp
                                  - Exists
                                  - DoesNotExist
                                  - Gt
                                  - Lt
                                  - GtLt
                                  - IsTrue
                                  - IsFalse
                                  type: string
                                value:
                                  description: Value is the list of values that the operand evaluates the input against. Value should be empty if the operator is Exists, DoesNotExist, IsTrue or IsFalse. Value should contain exactly one element if the operator is Gt or Lt and exactly two elements if the operator is GtLt. In other cases Value should contain at least one element.
                                  items:
                                    type: string
                                  type: array
                              required:
                              - op
                              type: object
                            description: MatchExpressionSet contains a set of MatchExpressions, each of which is evaluated against a set of input values.
                            type: object
                        required:
                        - feature
                        - matchExpressions
                        type: object
                      type: array
                    name:
                      description: Name of the rule.
                      type: string
                    vars:
                      additionalProperties:
                        type: string
                      description: Vars is the variables to store if the rule matches. Variables do not directly inflict any changes in the node object. However, they can be referenced from other rules enabling more complex rule hierarchies, without exposing intermediary output values as labels.
                      type: object
                    varsTemplate:
                      description: VarsTemplate specifies a template to expand for dynamically generating multiple variables. Data (after template expansion) must be keys with an optional value (<key>[=<value>]) separated by newlines.
                      type: string
                  required:
                  - name
                  type: object
                type: array
            required:
            - rules
            type: object
        required:
        - spec
        type: object
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: nfd-master
  namespace: node-feature-discovery
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: nfd-master
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - patch
  - update
  - list
- apiGroups:
  - topology.node.k8s.io
  resources:
  - noderesourcetopologies
  verbs:
  - create
  - get
  - update
- apiGroups:
  - nfd.k8s-sigs.io
  resources:
  - nodefeaturerules
  verbs:
  - get
  - list
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: nfd-master
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: nfd-master
subjects:
- kind: ServiceAccount
  name: nfd-master
  namespace: node-feature-discovery
---
apiVersion: v1
data:
  nfd-worker.conf: |
    #core:
    #  labelWhiteList:
    #  noPublish: false
    #  sleepInterval: 60s
    #  featureSources: [all]
    #  labelSources: [all]
    #  klog:
    #    addDirHeader: false
    #    alsologtostderr: false
    #    logBacktraceAt:
    #    logtostderr: true
    #    skipHeaders: false
    #    stderrthreshold: 2
    #    v: 0
    #    vmodule:
    ##   NOTE: the following options are not dynamically run-time configurable
    ##         and require a nfd-worker restart to take effect after being changed
    #    logDir:
    #    logFile:
    #    logFileMaxSize: 1800
    #    skipLogHeaders: false
    #sources:
    #  cpu:
    #    cpuid:
    ##     NOTE: whitelist has priority over blacklist
    #      attributeBlacklist:
    #        - "BMI1"
    #        - "BMI2"
    #        - "CLMUL"
    #        - "CMOV"
    #        - "CX16"
    #        - "ERMS"
    #        - "F16C"
    #        - "HTT"
    #        - "LZCNT"
    #        - "MMX"
    #        - "MMXEXT"
    #        - "NX"
    #        - "POPCNT"
    #        - "RDRAND"
    #        - "RDSEED"
    #        - "RDTSCP"
    #        - "SGX"
    #        - "SSE"
    #        - "SSE2"
    #        - "SSE3"
    #        - "SSE4"
    #        - "SSE42"
    #        - "SSSE3"
    #      attributeWhitelist:
    #  kernel:
    #    kconfigFile: "/path/to/kconfig"
    #    configOpts:
    #      - "NO_HZ"
    #      - "X86"
    #      - "DMI"
    #  pci:
    #    deviceClassWhitelist:
    #      - "0200"
    #      - "03"
    #      - "12"
    #    deviceLabelFields:
    #      - "class"
    #      - "vendor"
    #      - "device"
    #      - "subsystem_vendor"
    #      - "subsystem_device"
    #  usb:
    #    deviceClassWhitelist:
    #      - "0e"
    #      - "ef"
    #      - "fe"
    #      - "ff"
    #    deviceLabelFields:
    #      - "class"
    #      - "vendor"
    #      - "device"
    #  custom:
    #    # The following feature demonstrates the capabilities of the matchFeatures
    #    - name: "my custom rule"
    #      labels:
    #        my-ng-feature: "true"
    #      # matchFeatures implements a logical AND over all matcher terms in the
    #      # list (i.e. all of the terms, or per-feature matchers, must match)
    #      matchFeatures:
    #        - feature: cpu.cpuid
    #          matchExpressions:
    #            AVX512F: {op: Exists}
    #        - feature: cpu.cstate
    #          matchExpressions:
    #            enabled: {op: IsTrue}
    #        - feature: cpu.pstate
    #          matchExpressions:
    #            no_turbo: {op: IsFalse}
    #            scaling_governor: {op: In, value: ["performance"]}
    #        - feature: cpu.rdt
    #          matchExpressions:
    #            RDTL3CA: {op: Exists}
    #        - feature: cpu.sst
    #          matchExpressions:
    #            bf.enabled: {op: IsTrue}
    #        - feature: cpu.topology
    #          matchExpressions:
    #            hardware_multithreading: {op: IsFalse}
    #
    #        - feature: kernel.config
    #          matchExpressions:
    #            X86: {op: Exists}
    #            LSM: {op: InRegexp, value: ["apparmor"]}
    #        - feature: kernel.loadedmodule
    #          matchExpressions:
    #            e1000e: {op: Exists}
    #        - feature: kernel.selinux
    #          matchExpressions:
    #            enabled: {op: IsFalse}
    #        - feature: kernel.version
    #          matchExpressions:
    #            major: {op: In, value: ["5"]}
    #            minor: {op: Gt, value: ["10"]}
    #
    #        - feature: storage.block
    #          matchExpressions:
    #            rotational: {op: In, value: ["0"]}
    #            dax: {op: In, value: ["0"]}
    #
    #        - feature: network.device
    #          matchExpressions:
    #            operstate: {op: In, value: ["up"]}
    #            speed: {op: Gt, value: ["100"]}
    #
    #        - feature: memory.numa
    #          matchExpressions:
    #            node_count: {op: Gt, value: ["2"]}
    #        - feature: memory.nv
    #          matchExpressions:
    #            devtype: {op: In, value: ["nd_dax"]}
    #            mode: {op: In, value: ["memory"]}
    #
    #        - feature: system.osrelease
    #          matchExpressions:
    #            ID: {op: In, value: ["fedora", "centos"]}
    #        - feature: system.name
    #          matchExpressions:
    #            nodename: {op: InRegexp, value: ["^worker-X"]}
    #
    #        - feature: local.label
    #          matchExpressions:
    #            custom-feature-knob: {op: Gt, value: ["100"]}
    #
    #    # The following feature demonstrates the capabilities of the matchAny
    #    - name: "my matchAny rule"
    #      labels:
    #        my-ng-feature-2: "my-value"
    #      # matchAny implements a logical IF over all elements (sub-matchers) in
    #      # the list (i.e. at least one feature matcher must match)
    #      matchAny:
    #        - matchFeatures:
    #            - feature: kernel.loadedmodule
    #              matchExpressions:
    #                driver-module-X: {op: Exists}
    #            - feature: pci.device
    #              matchExpressions:
    #                vendor: {op: In, value: ["8086"]}
    #                class: {op: In, value: ["0200"]}
    #        - matchFeatures:
    #            - feature: kernel.loadedmodule
    #              matchExpressions:
    #                driver-module-Y: {op: Exists}
    #            - feature: usb.device
    #              matchExpressions:
    #                vendor: {op: In, value: ["8086"]}
    #                class: {op: In, value: ["02"]}
    #
    #    # The following features demonstreate label templating capabilities
    #    - name: "my template rule"
    #      labelsTemplate: |
    #      matchFeatures:
    #        - feature: system.osrelease
    #          matchExpressions:
    #            ID: {op: InRegexp, value: ["^open.*"]}
    #            VERSION_ID.major: {op: In, value: ["13", "15"]}
    #
    #    - name: "my template rule 2"
    #      matchFeatures:
    #        - feature: pci.device
    #          matchExpressions:
    #            class: {op: InRegexp, value: ["^06"]}
    #            vendor: ["8086"]
    #        - feature: cpu.cpuid
    #          matchExpressions:
    #            AVX: {op: Exists}
    #
    #    # The following examples demonstrate vars field and back-referencing
    #    # previous labels and vars
    #    - name: "my dummy kernel rule"
    #      labels:
    #        "my.kernel.feature": "true"
    #      matchFeatures:
    #        - feature: kernel.version
    #          matchExpressions:
    #            major: {op: Gt, value: ["2"]}
    #
    #    - name: "my dummy rule with no labels"
    #      vars:
    #        "my.dummy.var": "1"
    #      matchFeatures:
    #        - feature: cpu.cpuid
    #          matchExpressions: {}
    #
    #    - name: "my rule using backrefs"
    #      labels:
    #        "my.backref.feature": "true"
    #      matchFeatures:
    #        - feature: rule.matched
    #          matchExpressions:
    #            my.kernel.feature: {op: IsTrue}
    #            my.dummy.var: {op: Gt, value: ["0"]}
    #
kind: ConfigMap
metadata:
  name: nfd-worker-conf
  namespace: node-feature-discovery
---
apiVersion: v1
kind: Service
metadata:
  name: nfd-master
  namespace: node-feature-discovery
spec:
  ports:
  - port: 8080
    protocol: TCP
  selector:
    app: nfd-master
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nfd
  name: nfd-master
  namespace: node-feature-discovery
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nfd-master
  template:
    metadata:
      labels:
        app: nfd-master
    spec:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - preference:
              matchExpressions:
              - key: node-role.kubernetes.io/master
                operator: In
                values:
                - ""
            weight: 1
          - preference:
              matchExpressions:
              - key: node-role.kubernetes.io/control-plane
                operator: In
                values:
                - ""
            weight: 1
      containers:
      - args: []
        command:
        - nfd-master
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        image: {{ .NFDImage }}
        imagePullPolicy: IfNotPresent
        livenessProbe:
          exec:
            command:
            - /usr/bin/grpc_health_probe
            - -addr=:8080
          initialDelaySeconds: 10
          periodSeconds: 10
        name: nfd-master
        readinessProbe:
          exec:
            command:
            - /usr/bin/grpc_health_probe
            - -addr=:8080
          failureThreshold: 10
          initialDelaySeconds: 5
          periodSeconds: 10
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: true
          runAsNonRoot: true
        volumeMounts: []
      serviceAccount: nfd-master
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
        operator: Equal
        value: ""
      - effect: NoSchedule
        key: node-role.kubernetes.io/control-plane
        operator: Equal
        value: ""
      volumes: []
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: nfd
  name: nfd-worker
  namespace: node-feature-discovery
spec:
  selector:
    matchLabels:
      app: nfd-worker
  template:
    metadata:
      labels:
        app: nfd-worker
    spec:
      containers:
      - args:
        - -server=nfd-master:8080
        command:
        - nfd-worker
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        image: {{ .NFDImage }}
        imagePullPolicy: IfNotPresent
        name: nfd-worker
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: true
          runAsNonRoot: true
        volumeMounts:
        - mountPath: /host-boot
          name: host-boot
          readOnly: true
        - mountPath: /host-etc/os-release
          name: host-os-release
          readOnly: true
        - mountPath: /host-sys
          name: host-sys
          readOnly: true
        - mountPath: /host-usr/lib
          name: host-usr-lib
          readOnly: true
        - mountPath: /etc/kubernetes/node-feature-discovery/source.d/
          name: source-d
          readOnly: true
        - mountPath: /etc/kubernetes/node-feature-discovery/features.d/
          name: features-d
          readOnly: true
        - mountPath: /etc/kubernetes/node-feature-discovery
          name: nfd-worker-conf
          readOnly: true
      dnsPolicy: ClusterFirstWithHostNet
      volumes:
      - hostPath:
          path: /boot
        name: host-boot
      - hostPath:
          path: /etc/os-release
        name: host-os-release
      - hostPath:
          path: /sys
        name: host-sys
      - hostPath:
          path: /usr/lib
        name: host-usr-lib
      - hostPath:
          path: /etc/kubernetes/node-feature-discovery/source.d/
        name: source-d
      - hostPath:
          path: /etc/kubernetes/node-feature-discovery/features.d/
        name: features-d
      - configMap:
          name: nfd-worker-conf
        name: nfd-worker-conf

    `)))
)

func DeployNodeFeatureDiscoveryTasks(d *DeployPluginsModule) []task.Interface {
	generateNFDManifests := &task.RemoteTask{
		Name:    "GenerateNFDManifests",
		Desc:    "Generate node-feature-discovery manifests",
		Hosts:   d.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action: &action.Template{
			Name:     "GenerateNFDManifests",
			Template: NodeFeatureDiscovery,
			Data: util.Data{
				"NFDImage": images.GetImage(d.Runtime, d.KubeConf, "node-feature-discovery").ImageName(),
			},
			Dst: filepath.Join(common.KubeAddonsDir, NodeFeatureDiscovery.Name()),
		},
		Parallel: false,
	}

	deployNFD := &task.RemoteTask{
		Name:    "ApplyNFDManifests",
		Desc:    "Apply node-feature-discovery manifests",
		Hosts:   d.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(ApplyNodeFeatureDiscoveryManifests),
	}

	return []task.Interface{
		generateNFDManifests,
		deployNFD,
	}
}

type ApplyNodeFeatureDiscoveryManifests struct {
	common.KubeAction
}

func (a *ApplyNodeFeatureDiscoveryManifests) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/addons/node-feature-discovery.yaml", true, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "apply node-feature-discovery manifests failed")
	}
	return nil
}
