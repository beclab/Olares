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

package confirm

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/modood/table"
	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
)

// PreCheckResults defines the items to be checked.
type PreCheckResults struct {
	Name       string `table:"name"`
	Sudo       string `table:"sudo"`
	Curl       string `table:"curl"`
	Openssl    string `table:"openssl"`
	Ebtables   string `table:"ebtables"`
	Socat      string `table:"socat"`
	Ipset      string `table:"ipset"`
	Ipvsadm    string `table:"ipvsadm"`
	Conntrack  string `table:"conntrack"`
	Chronyd    string `table:"chrony"`
	Docker     string `table:"docker"`
	Containerd string `table:"containerd"`
	Nfs        string `table:"nfs client"`
	Ceph       string `table:"ceph client"`
	Glusterfs  string `table:"glusterfs client"`
	Time       string `table:"time"`
}

type InstallationConfirm struct {
	common.KubeAction
}

func (i *InstallationConfirm) Execute(runtime connector.Runtime) error {
	var (
		results  []PreCheckResults
		stopFlag bool
	)

	pre := make([]map[string]string, 0, len(runtime.GetAllHosts()))
	for _, host := range runtime.GetAllHosts() {
		if v, ok := host.GetCache().Get(common.NodePreCheck); ok {
			pre = append(pre, v.(map[string]string))
		} else {
			return errors.New("get node check result failed by host cache")
		}
	}

	for node := range pre {
		var result PreCheckResults
		_ = mapstructure.Decode(pre[node], &result)
		results = append(results, result)
	}
	table.OutputA(results)
	reader := bufio.NewReader(os.Stdin)

	if i.KubeConf.Arg.Artifact == "" {
		for _, host := range results {
			if host.Sudo == "" {
				logger.Errorf("%s: sudo is required.", host.Name)
				stopFlag = true
			}

			if host.Conntrack == "" {
				logger.Errorf("%s: conntrack is required.", host.Name)
				stopFlag = true
			}

			if host.Socat == "" {
				logger.Errorf("%s: socat is required.", host.Name)
				stopFlag = true
			}
		}
	}

	fmt.Println("")
	fmt.Println("This is a simple check of your environment.")
	fmt.Println("Before installation, ensure that your machines meet all requirements specified at")
	fmt.Println("https://github.com/kubesphere/kubekey#requirements-and-recommendations")
	fmt.Println("")

	if k8sVersion, err := versionutil.ParseGeneric(i.KubeConf.Cluster.Kubernetes.Version); err == nil {
		if k8sVersion.AtLeast(versionutil.MustParseSemantic("v1.24.0")) && i.KubeConf.Cluster.Kubernetes.ContainerManager == common.Docker {
			fmt.Println("[Notice]")
			fmt.Println("Incorrect runtime. Please specify a container runtime other than Docker to install Kubernetes v1.24 or later.")
			fmt.Println("You can set \"spec.kubernetes.containerManager\" in the configuration file to \"containerd\" or add \"--container-manager containerd\" to the \"./kk create cluster\" command.")
			fmt.Println("For more information, see:")
			fmt.Println("https://github.com/kubesphere/kubekey/blob/master/docs/commands/kk-create-cluster.md")
			fmt.Println("https://kubernetes.io/docs/setup/production-environment/container-runtimes/#container-runtimes")
			fmt.Println("https://kubernetes.io/blog/2022/02/17/dockershim-faq/")
			fmt.Println("")
			stopFlag = true
		}
	}

	if stopFlag {
		os.Exit(1)
	}

	confirmOK := true // TODO: force skip
	for !confirmOK {
		fmt.Printf("Continue this installation? [yes/no]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			logger.Fatal(err)
		}
		input = strings.TrimSpace(strings.ToLower(input))

		switch strings.ToLower(input) {
		case "yes", "y":
			confirmOK = true
		case "no", "n":
			os.Exit(0)
		default:
			continue
		}
	}
	return nil
}
