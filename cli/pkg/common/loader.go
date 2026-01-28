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

package common

import (
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/util"

	kubekeyapiv1alpha2 "github.com/beclab/Olares/cli/apis/kubekey/v1alpha2"
	"github.com/pkg/errors"
)

var (
	kubeReleaseRegex = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)([-0-9a-zA-Z_\.+]*)?$`)
)

type Loader interface {
	Load() (*kubekeyapiv1alpha2.Cluster, error)
}

func NewLoader(arg Argument) Loader {
	return NewDefaultLoader(arg)
}

type DefaultLoader struct {
	arg               Argument
	KubernetesVersion string
}

func NewDefaultLoader(arg Argument) *DefaultLoader {
	return &DefaultLoader{
		arg:               arg,
		KubernetesVersion: arg.KubernetesVersion,
	}
}

func (d *DefaultLoader) Load() (*kubekeyapiv1alpha2.Cluster, error) {
	osType := d.arg.SystemInfo.GetOsType()
	user := d.arg.SystemInfo.GetUsername()
	homeDir := d.arg.SystemInfo.GetHomeDir()

	if osType != Darwin && osType != Windows {
		if user != "root" {
			return nil, errors.New(fmt.Sprintf("Current user is %s. Please use root!", user))
		}
	}

	// u, err := currentUser(osType)
	// if err != nil {
	// 	return nil, err
	// }

	fmt.Printf("current: %s\n", user)

	allInOne := &kubekeyapiv1alpha2.Cluster{}

	if osType != Darwin && osType != Windows {
		if err := installSUDOIfMissing(); err != nil {
			return nil, err
		}
	}

	if err := localSSH(osType); err != nil {
		return nil, err
	}

	ip := d.arg.SystemInfo.GetLocalIp()
	hostname := d.arg.SystemInfo.GetHostname()

	allInOne.Spec.Hosts = append(allInOne.Spec.Hosts, kubekeyapiv1alpha2.HostCfg{
		Name:            hostname,
		Address:         ip,
		InternalAddress: ip,
		Port:            kubekeyapiv1alpha2.DefaultSSHPort,
		User:            user,
		Password:        "",
		PrivateKeyPath:  fmt.Sprintf("%s/.ssh/id_rsa", homeDir),
		Arch:            d.arg.SystemInfo.GetOsArch(),
	})

	if d.arg.MasterHost == "" {
		allInOne.Spec.RoleGroups = map[string][]string{
			Master:   {hostname},
			ETCD:     {hostname},
			Worker:   {hostname},
			Registry: {hostname},
		}
	} else {
		allInOne.Spec.Hosts = append(allInOne.Spec.Hosts, kubekeyapiv1alpha2.HostCfg{
			Name:            d.arg.MasterNodeName,
			Address:         d.arg.MasterHost,
			InternalAddress: d.arg.MasterHost,
			Port:            d.arg.MasterSSHPort,
			User:            d.arg.MasterSSHUser,
			Password:        d.arg.MasterSSHPassword,
			PrivateKeyPath:  d.arg.MasterSSHPrivateKeyPath,
		})
		allInOne.Spec.RoleGroups = map[string][]string{
			Master:   {d.arg.MasterNodeName},
			ETCD:     {d.arg.MasterNodeName},
			Worker:   {d.arg.MasterNodeName, hostname},
			Registry: {d.arg.MasterNodeName},
		}
	}

	if ver := normalizedBuildVersion(d.KubernetesVersion); ver != "" {
		s := strings.Split(ver, "-")
		if len(s) > 1 {
			allInOne.Spec.Kubernetes = kubekeyapiv1alpha2.Kubernetes{
				Version: s[0],
				Type:    s[1],
			}
		} else {
			allInOne.Spec.Kubernetes = kubekeyapiv1alpha2.Kubernetes{
				Version: ver,
			}
		}
	} else {
		allInOne.Spec.Kubernetes = kubekeyapiv1alpha2.Kubernetes{
			Version: kubekeyapiv1alpha2.DefaultKubeVersion,
		}
	}

	if err := defaultCommonClusterConfig(allInOne, d.arg); err != nil {
		return nil, err
	}

	// certs renew
	enableAutoRenewCerts := true
	allInOne.Spec.Kubernetes.AutoRenewCerts = &enableAutoRenewCerts

	return allInOne, nil
}

// normalizedBuildVersion used to returns normalized build version (with "v" prefix if needed)
// If input doesn't match known version pattern, returns empty string.
func normalizedBuildVersion(version string) string {
	if kubeReleaseRegex.MatchString(version) {
		if strings.HasPrefix(version, "v") {
			return version
		}
		return "v" + version
	}
	return ""
}

func installSUDOIfMissing() error {
	p, _ := util.GetCommand("sudo")
	if p != "" {
		return nil
	}
	output, err := exec.Command("/bin/sh", "-c", "apt install -y sudo").CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "failed to install the sudo command that's missing: %s", string(output))
	}
	return nil
}

func localSSH(osType string) error {
	switch osType {
	case Windows:
		return nil
	default:
	}
	if output, err := exec.Command("/bin/sh", "-c", "if [ ! -f \"$HOME/.ssh/id_rsa\" ]; then mkdir -p \"$HOME/.ssh\" && ssh-keygen -t rsa-sha2-512 -P \"\" -f $HOME/.ssh/id_rsa && ls $HOME/.ssh;fi;").CombinedOutput(); err != nil {
		return errors.New(fmt.Sprintf("Failed to generate public key: %v\n%s", err, string(output)))
	}
	if output, err := exec.Command("/bin/sh", "-c", "sudo -E /bin/bash -c 'echo \"\n$(cat $HOME/.ssh/id_rsa.pub)\" >> $HOME/.ssh/authorized_keys' && awk ' !x[$0]++{print > \"'$HOME'/.ssh/authorized_keys.tmp\"}' $HOME/.ssh/authorized_keys && mv $HOME/.ssh/authorized_keys.tmp $HOME/.ssh/authorized_keys").CombinedOutput(); err != nil {
		return errors.New(fmt.Sprintf("Failed to copy public key to authorized_keys: %v\n%s", err, string(output)))
	}

	return nil
}

// defaultCommonClusterConfig kubernetes version, registry mirrors, container manager, etc.
func defaultCommonClusterConfig(cluster *kubekeyapiv1alpha2.Cluster, arg Argument) error {
	if ver := normalizedBuildVersion(arg.KubernetesVersion); ver != "" {
		s := strings.Split(ver, "-")
		if len(s) > 1 {
			cluster.Spec.Kubernetes = kubekeyapiv1alpha2.Kubernetes{
				Version: s[0],
				Type:    s[1],
			}
		} else {
			cluster.Spec.Kubernetes = kubekeyapiv1alpha2.Kubernetes{
				Version: ver,
			}
		}
	} else {
		cluster.Spec.Kubernetes = kubekeyapiv1alpha2.Kubernetes{
			Version: kubekeyapiv1alpha2.DefaultKubeVersion,
		}
	}

	if arg.RegistryMirrors != "" {
		mirrors := strings.Split(arg.RegistryMirrors, ",")

		for i := range mirrors {
			mirror := mirrors[i]
			if !(strings.HasPrefix(mirror, "http://") || strings.HasPrefix(mirror, "https://")) {
				return errors.New(fmt.Sprintf("Invalid registry mirror: %s, missing scheme 'http://' or 'https://'", mirror))
			}
			u, err := url.Parse(mirror)
			if err != nil {
				return fmt.Errorf("invalid registry mirror: %s: %w", mirror, err)
			}

			// match against paths containing only "/"(s)
			// e.g. "/", "//", "///" (they're all considered valid by url.Parse)
			if strings.Count(u.Path, "/") == len(u.Path) {
				u.Path = strings.ReplaceAll(u.Path, "/", "")
			}
			mirrors[i] = u.String()
		}

		cluster.Spec.Registry.RegistryMirrors = mirrors
	}

	if arg.ContainerManager != "" && arg.ContainerManager != Docker {
		cluster.Spec.Kubernetes.ContainerManager = arg.ContainerManager
	}

	// must be a lower case
	cluster.Name = "kubekey" + time.Now().Format("2006-01-02")

	return nil
}
