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
	"bufio"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/version/kubesphere"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

var (
	kubeReleaseRegex = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)([-0-9a-zA-Z_\.+]*)?$`)
)

type Loader interface {
	Load() (*kubekeyapiv1alpha2.Cluster, error)
}

type Options map[string]interface{}

func NewLoader(flag string, arg Argument) Loader {
	switch flag {
	case File:
		return NewFileLoader(arg)
	case Operator:
		return &ConfigMapLoader{}
	case AllInOne:
		return NewDefaultLoader(arg)
	case CommandLine:
		return NewCommandLineLoader(arg)
	default:
		return NewDefaultLoader(arg)
	}
}

type CommandLineLoader struct {
	arg               Argument
	hostname          string
	kubernetesVersion string
}

func NewCommandLineLoader(arg Argument) *CommandLineLoader {
	return &CommandLineLoader{
		arg:               arg,
		kubernetesVersion: arg.KubernetesVersion,
	}
}

func (c *CommandLineLoader) validate() error {
	hostname, err := os.Hostname()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to get hostname: %v\n", err))
	}

	c.hostname = hostname

	// flags
	if c.arg.MasterNodeName == "" {
		return errors.New("No master nodeName provided, with flag '--master-node-name'")
	}
	if c.arg.MasterHost == "" {
		return errors.New("No master host provided, with flag '--master-host'")
	}
	if c.arg.MasterSSHUser == "" {
		return errors.New("No master ssh user provided, with flag '--master-ssh-user'")
	}
	if c.arg.KubernetesVersion == "" {
		return errors.New("No kubernetes version provided, with flag '--with-kubernetes'")
	}
	if c.arg.MasterSSHPassword == "" && c.arg.MasterSSHPrivateKeyPath == "" {
		return errors.New("No master ssh password and private key file, with flag '--master-ssh-password' or '--master-ssh-private-keyfile'")
	}

	if err := localSSH(c.arg.SystemInfo.GetOsType()); err != nil {
		return err
	}

	return nil
}

func (c *CommandLineLoader) Load() (*kubekeyapiv1alpha2.Cluster, error) {
	u, err := currentUser(c.arg.SystemInfo.GetOsType())
	if err != nil {
		return nil, err
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	cluster := &kubekeyapiv1alpha2.Cluster{}

	// current node
	cluster.Spec.Hosts = append(cluster.Spec.Hosts, kubekeyapiv1alpha2.HostCfg{
		Name:            c.hostname,
		Address:         "",
		InternalAddress: "",
		Port:            c.arg.LocalSSHPort,
		User:            u.Name,
		PrivateKeyPath:  fmt.Sprintf("%s/.ssh/id_rsa", u.HomeDir),
		Arch:            "",
	})

	cluster.Spec.RoleGroups = map[string][]string{
		Worker: {c.hostname},
	}

	// master node
	masterHostCfg := kubekeyapiv1alpha2.HostCfg{
		Name:            c.arg.MasterNodeName,
		Address:         c.arg.MasterHost,
		InternalAddress: c.arg.MasterHost,
		Port:            c.arg.MasterSSHPort,
		User:            c.arg.MasterSSHUser,
		Arch:            "",
	}
	if c.arg.MasterSSHPassword != "" {
		masterHostCfg.Password = c.arg.MasterSSHPassword
	}
	if c.arg.MasterSSHPrivateKeyPath != "" {
		masterHostCfg.Password = ""
		masterHostCfg.PrivateKeyPath = c.arg.MasterSSHPrivateKeyPath
	}

	cluster.Spec.Hosts = append(cluster.Spec.Hosts, masterHostCfg)

	cluster.Spec.RoleGroups[Master] = []string{c.arg.MasterNodeName}
	cluster.Spec.RoleGroups[ETCD] = []string{c.arg.MasterNodeName}

	if err := defaultCommonClusterConfig(cluster, c.arg); err != nil {
		return nil, err
	}

	return cluster, nil
}

type DefaultLoader struct {
	arg               Argument
	KubernetesVersion string
	KubeSphereVersion string
	KubeSphereEnable  bool
}

func NewDefaultLoader(arg Argument) *DefaultLoader {
	return &DefaultLoader{
		arg:               arg,
		KubernetesVersion: arg.KubernetesVersion,
		KubeSphereVersion: arg.KsVersion,
		KubeSphereEnable:  arg.KsEnable,
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

	if d.KubeSphereEnable {
		ver := normalizedBuildVersion(d.KubeSphereVersion)
		if ver == "" {
			return nil, errors.New(fmt.Sprintf("Unsupported Kubesphere Version: %v\n", d.KubeSphereVersion))
		}
		if err := defaultKSConfig(&allInOne.Spec.KubeSphere, ver); err != nil {
			return nil, err
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

type FileLoader struct {
	arg               Argument
	FilePath          string
	KubernetesVersion string
	KubeSphereVersion string
	KubeSphereEnable  bool
}

func NewFileLoader(arg Argument) *FileLoader {
	return &FileLoader{
		arg:               arg,
		FilePath:          arg.FilePath,
		KubernetesVersion: arg.KubernetesVersion,
		KubeSphereVersion: arg.KsVersion,
		KubeSphereEnable:  arg.KsEnable,
	}
}

func (f FileLoader) Load() (*kubekeyapiv1alpha2.Cluster, error) {
	var objName string

	clusterCfg := kubekeyapiv1alpha2.Cluster{}
	fp, err := filepath.Abs(f.FilePath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to look up current directory")
	}
	// fixme: It will lead to nil pointer err
	//if len(f.KubernetesVersion) != 0 {
	//	_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("sed -i \"/version/s/\\:.*/\\: %s/g\" %s", f.KubernetesVersion, fp)).Run()
	//}
	file, err := os.Open(fp)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to open the given cluster configuration file")
	}
	defer file.Close()
	b1 := bufio.NewReader(file)
	for {
		result := make(map[string]interface{})
		content, err := k8syaml.NewYAMLReader(b1).Read()
		if len(content) == 0 {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "Unable to read the given cluster configuration file")
		}
		err = yaml.Unmarshal(content, &result)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to unmarshal the given cluster configuration file")
		}

		if result["kind"] == "Cluster" {
			contentToJson, err := k8syaml.ToJSON(content)
			if err != nil {
				return nil, errors.Wrap(err, "Unable to convert configuration to json")
			}
			if err := json.Unmarshal(contentToJson, &clusterCfg); err != nil {
				return nil, errors.Wrap(err, "Failed to unmarshal configuration")
			}
			metadata := result["metadata"].(map[interface{}]interface{})
			objName = metadata["name"].(string)
		}

		if result["kind"] == "ConfigMap" || result["kind"] == "ClusterConfiguration" {
			metadata := result["metadata"].(map[interface{}]interface{})
			labels := metadata["labels"].(map[interface{}]interface{})
			clusterCfg.Spec.KubeSphere.Enabled = true

			v, ok := labels["version"]
			if !ok {
				return nil, errors.New("Unknown version")
			}

			version := v.(string)
			_, stable := kubesphere.StabledVersionSupport(version)
			_, latest := kubesphere.LatestRelease(version)
			_, dev := kubesphere.DevRelease(version)
			if stable || latest || dev {
				clusterCfg.Spec.KubeSphere.Configurations = "---\n" + string(content)
				clusterCfg.Spec.KubeSphere.Version = version
			} else {
				return nil, errors.New(fmt.Sprintf("Unsupported KubeSphere version: %s", version))
			}
		}
	}

	if f.KubeSphereEnable {
		ver := normalizedBuildVersion(f.KubeSphereVersion)
		if ver == "" {
			return nil, errors.New(fmt.Sprintf("Unsupported Kubesphere Version: %v\n", f.KubeSphereVersion))
		}
		if err := defaultKSConfig(&clusterCfg.Spec.KubeSphere, ver); err != nil {
			return nil, err
		}
	}

	if ver := normalizedBuildVersion(f.KubernetesVersion); ver != "" {
		s := strings.Split(ver, "-")
		if len(s) > 1 {
			clusterCfg.Spec.Kubernetes.Version = s[0]
			clusterCfg.Spec.Kubernetes.Type = s[1]
		} else {
			clusterCfg.Spec.Kubernetes.Version = ver
		}
	}

	if f.arg.ContainerManager != "" && f.arg.ContainerManager != Docker {
		clusterCfg.Spec.Kubernetes.ContainerManager = f.arg.ContainerManager
	}

	clusterCfg.Spec.Kubernetes.Version = normalizedBuildVersion(clusterCfg.Spec.Kubernetes.Version)
	clusterCfg.Spec.KubeSphere.Version = normalizedBuildVersion(clusterCfg.Spec.KubeSphere.Version)
	clusterCfg.Name = objName
	return &clusterCfg, nil
}

type ConfigMapLoader struct {
}

func (c ConfigMapLoader) Load() (*kubekeyapiv1alpha2.Cluster, error) {
	return nil, nil
}

func defaultKSConfig(ks *kubekeyapiv1alpha2.KubeSphere, version string) error {
	ks.Enabled = true
	version = strings.TrimSpace(version)
	ksInstaller, ok := kubesphere.StabledVersionSupport(version)
	if ok {
		ks.Version = ksInstaller.Version
		ks.Configurations = ksInstaller.CCToString()
	} else if latest, ok := kubesphere.LatestRelease(version); ok {
		ks.Version = version
		ks.Configurations = latest.CCToString()
	} else if dev, ok := kubesphere.DevRelease(version); ok {
		ks.Version = version
		ks.Configurations = dev.CCToString()
	} else {
		return errors.New(fmt.Sprintf("Unsupported KubeSphere version: %s", version))
	}
	return nil
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

type ExtraAddon struct {
	Addons []kubekeyapiv1alpha2.Addon `yaml:"Addons"`
}

// load addon from argument
func loadExtraAddons(cluster *kubekeyapiv1alpha2.Cluster, addonFile string) error {
	if addonFile == "" {
		return nil
	}

	fp, err := filepath.Abs(addonFile)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to load addon in file: %s", addonFile))
	}

	content, err := ioutil.ReadFile(fp)
	if err != nil {
		return errors.Wrap(err, "Unable to open the given addon config file")
	}

	if len(content) == 0 {
		return nil
	}

	var result ExtraAddon
	err = yaml.Unmarshal(content, &result)
	if err != nil {
		return errors.Wrap(err, "Unable to read the given addon configuration file")
	}

	if len(result.Addons) > 0 {
		cluster.Spec.Addons = append(cluster.Spec.Addons, result.Addons...)
	}

	return nil
}

func currentUser(osType string) (*user.User, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	if osType != Darwin && osType != Windows {
		if u.Username != "root" {
			return nil, errors.New(fmt.Sprintf("Current user is %s. Please use root!", u.Username))
		}
	}
	return u, nil
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
