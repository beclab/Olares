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

package certs

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"bytetrade.io/web3os/installer/pkg/certs/templates"
	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	certutil "k8s.io/client-go/util/cert"
)

type Certificate struct {
	Name          string
	Expires       string
	Residual      string
	AuthorityName string
	NodeName      string
}

type CaCertificate struct {
	AuthorityName string
	Expires       string
	Residual      string
	NodeName      string
}

var (
	certificateList = []string{
		"apiserver.crt",
		"apiserver-kubelet-client.crt",
		"front-proxy-client.crt",
	}
	caCertificateList = []string{
		"ca.crt",
		"front-proxy-ca.crt",
	}
	kubeConfigList = []string{
		"admin.conf",
		"controller-manager.conf",
		"scheduler.conf",
	}
)

type ListClusterCerts struct {
	common.KubeAction
}

func (l *ListClusterCerts) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()

	certificates := make([]*Certificate, 0)
	caCertificates := make([]*CaCertificate, 0)

	for _, certFileName := range certificateList {
		certPath := filepath.Join(common.KubeCertDir, certFileName)
		certContext, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("cat %s", certPath), false, false)
		if err != nil {
			return errors.Wrap(err, "get cluster certs failed")
		}
		if cert, err := getCertInfo(certContext, certFileName, host.GetName()); err != nil {
			return err
		} else {
			certificates = append(certificates, cert)
		}
	}
	for _, kubeConfigFileName := range kubeConfigList {
		kubeConfigPath := filepath.Join(common.KubeConfigDir, kubeConfigFileName)
		newConfig := clientcmdapi.NewConfig()
		kubeconfigBytes, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("cat %s", kubeConfigPath), false, false)
		decoded, _, err := clientcmdlatest.Codec.Decode([]byte(kubeconfigBytes), &schema.GroupVersionKind{Version: clientcmdlatest.Version, Kind: "Config"}, newConfig)
		if err != nil {
			return err
		}
		newConfig = decoded.(*clientcmdapi.Config)
		for _, a := range newConfig.AuthInfos {
			certContextBase64 := a.ClientCertificateData
			tmp := base64.StdEncoding.EncodeToString(certContextBase64)
			certContext, err := base64.StdEncoding.DecodeString(tmp)
			if err != nil {
				return err
			}
			if cert, err := getCertInfo(string(certContext), kubeConfigFileName, host.GetName()); err != nil {
				return err
			} else {
				certificates = append(certificates, cert)
			}
		}
	}

	for _, caCertFileName := range caCertificateList {
		certPath := filepath.Join(common.KubeCertDir, caCertFileName)
		caCertContext, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("cat %s", certPath), false, false)
		if err != nil {
			return errors.Wrap(err, "Failed to get cluster certs")
		}
		if cert, err := getCaCertInfo(caCertContext, caCertFileName, host.GetName()); err != nil {
			return err
		} else {
			caCertificates = append(caCertificates, cert)
		}
	}

	host.GetCache().Set(common.Certificate, certificates)
	host.GetCache().Set(common.CaCertificate, caCertificates)
	return nil
}

func getCertInfo(certContext, certFileName, nodeName string) (*Certificate, error) {
	certs, err1 := certutil.ParseCertsPEM([]byte(certContext))
	if err1 != nil {
		return nil, errors.Wrap(err1, "Failed to get cluster certs")
	}
	var authorityName string
	switch certFileName {
	case "apiserver.crt":
		authorityName = "ca"
	case "apiserver-kubelet-client.crt":
		authorityName = "ca"
	case "front-proxy-client.crt":
		authorityName = "front-proxy-ca"
	default:
		authorityName = ""
	}
	cert := Certificate{
		Name:          certFileName,
		Expires:       certs[0].NotAfter.Format("Jan 02, 2006 15:04 MST"),
		Residual:      ResidualTime(certs[0].NotAfter),
		AuthorityName: authorityName,
		NodeName:      nodeName,
	}
	return &cert, nil
}

func getCaCertInfo(certContext, certFileName, nodeName string) (*CaCertificate, error) {
	certs, err := certutil.ParseCertsPEM([]byte(certContext))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get cluster certs")
	}
	cert1 := CaCertificate{
		AuthorityName: certFileName,
		Expires:       certs[0].NotAfter.Format("Jan 02, 2006 15:04 MST"),
		Residual:      ResidualTime(certs[0].NotAfter),
		NodeName:      nodeName,
	}
	return &cert1, nil
}

func ResidualTime(t time.Time) string {
	d := time.Until(t)
	if seconds := int(d.Seconds()); seconds < -1 {
		return fmt.Sprintf("<invalid>")
	} else if seconds < 0 {
		return fmt.Sprintf("0s")
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	} else if hours := int(d.Hours()); hours < 24 {
		return fmt.Sprintf("%dh", hours)
	} else if hours < 24*365 {
		return fmt.Sprintf("%dd", hours/24)
	}
	return fmt.Sprintf("%dy", int(d.Hours()/24/365))
}

type DisplayForm struct {
	common.KubeAction
}

func (d *DisplayForm) Execute(runtime connector.Runtime) error {
	certificates := make([]*Certificate, 0)
	caCertificates := make([]*CaCertificate, 0)

	for _, host := range runtime.GetHostsByRole(common.Master) {
		certs, ok := host.GetCache().Get(common.Certificate)
		if !ok {
			return errors.New("get certificate failed by pipeline cache")
		}
		ca, ok := host.GetCache().Get(common.CaCertificate)
		if !ok {
			return errors.New("get ca certificate failed by pipeline cache")
		}
		hostCertificates := certs.([]*Certificate)
		hostCaCertificates := ca.([]*CaCertificate)
		certificates = append(certificates, hostCertificates...)
		caCertificates = append(caCertificates, hostCaCertificates...)
	}

	w := tabwriter.NewWriter(os.Stdout, 10, 4, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "CERTIFICATE\tEXPIRES\tRESIDUAL TIME\tCERTIFICATE AUTHORITY\tNODE")
	for _, cert := range certificates {
		s := fmt.Sprintf("%s\t%s\t%s\t%s\t%-8v",
			cert.Name,
			cert.Expires,
			cert.Residual,
			cert.AuthorityName,
			cert.NodeName,
		)

		_, _ = fmt.Fprintln(w, s)
		continue
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "CERTIFICATE AUTHORITY\tEXPIRES\tRESIDUAL TIME\tNODE")
	for _, caCert := range caCertificates {
		c := fmt.Sprintf("%s\t%s\t%s\t%-8v",
			caCert.AuthorityName,
			caCert.Expires,
			caCert.Residual,
			caCert.NodeName,
		)

		_, _ = fmt.Fprintln(w, c)
		continue
	}

	_ = w.Flush()
	return nil
}

type RenewCerts struct {
	common.KubeAction
}

func (r *RenewCerts) Execute(runtime connector.Runtime) error {
	var kubeadmAlphaList = []string{
		"/usr/local/bin/kubeadm alpha certs renew apiserver",
		"/usr/local/bin/kubeadm alpha certs renew apiserver-kubelet-client",
		"/usr/local/bin/kubeadm alpha certs renew front-proxy-client",
		"/usr/local/bin/kubeadm alpha certs renew admin.conf",
		"/usr/local/bin/kubeadm alpha certs renew controller-manager.conf",
		"/usr/local/bin/kubeadm alpha certs renew scheduler.conf",
	}

	var kubeadmList = []string{
		"/usr/local/bin/kubeadm certs renew apiserver",
		"/usr/local/bin/kubeadm certs renew apiserver-kubelet-client",
		"/usr/local/bin/kubeadm certs renew front-proxy-client",
		"/usr/local/bin/kubeadm certs renew admin.conf",
		"/usr/local/bin/kubeadm certs renew controller-manager.conf",
		"/usr/local/bin/kubeadm certs renew scheduler.conf",
	}

	var restartList = []string{
		"docker ps -af name=k8s_kube-apiserver* -q | xargs --no-run-if-empty docker rm -f",
		"docker ps -af name=k8s_kube-scheduler* -q | xargs --no-run-if-empty docker rm -f",
		"docker ps -af name=k8s_kube-controller-manager* -q | xargs --no-run-if-empty docker rm -f",
		"systemctl restart kubelet",
	}

	version, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubeadm version -o short", true, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeadm get version failed")
	}
	cmp, err := versionutil.MustParseSemantic(version).Compare("v1.20.0")
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "parse kubeadm version failed")
	}
	if cmp == -1 {
		_, err := runtime.GetRunner().SudoCmd(strings.Join(kubeadmAlphaList, " && "), false, false)
		if err != nil {
			return errors.Wrap(err, "kubeadm alpha certs renew failed")
		}
	} else {
		_, err := runtime.GetRunner().SudoCmd(strings.Join(kubeadmList, " && "), false, false)
		if err != nil {
			return errors.Wrap(err, "kubeadm alpha certs renew failed")
		}
	}

	_, err = runtime.GetRunner().SudoCmd(strings.Join(restartList, " && "), false, false)
	if err != nil {
		return errors.Wrap(err, "kube-apiserver, kube-schedule, kube-controller-manager or kubelet restart failed")
	}
	return nil
}

type FetchKubeConfig struct {
	common.KubeAction
}

func (f *FetchKubeConfig) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	tmpConfigFile := filepath.Join(common.TmpDir, "admin.conf")
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("cp /etc/kubernetes/admin.conf %s", tmpConfigFile), false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "copy kube config to /tmp/ failed")
	}

	host := runtime.RemoteHost()
	if err := runtime.GetRunner().Fetch(filepath.Join(runtime.GetWorkDir(), host.GetName(), "admin.conf"), tmpConfigFile, false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "fetch kube config file failed")
	}
	return nil
}

type SyneKubeConfigToWorker struct {
	common.KubeAction
}

func (s *SyneKubeConfigToWorker) Execute(runtime connector.Runtime) error {
	createConfigDirCmd := "mkdir -p /root/.kube"
	if _, err := runtime.GetRunner().SudoCmd(createConfigDirCmd, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "create .kube dir failed")
	}

	firstMaster := runtime.GetHostsByRole(common.Master)[0]
	localFile := filepath.Join(runtime.GetWorkDir(), firstMaster.GetName(), "admin.conf")
	if err := runtime.GetRunner().SudoScp(localFile, "/root/.kube/config"); err != nil {
		return errors.Wrap(errors.WithStack(err), "sudo scp config file to worker /root/.kube/config failed")
	}

	if _, err := runtime.GetRunner().SudoCmd("chmod 0600 /root/.kube/config", false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "chmod 0600 /root/.kube/config failed")
	}

	// that doesn't work
	//if err := runtime.GetRunner().SudoScp(filepath.Join(runtime.GetWorkDir(), firstMaster.GetName(), "admin.conf"), "$HOME/.kube/config"); err != nil {
	//	return errors.Wrap(errors.WithStack(err), "sudo scp config file to worker $HOME/.kube/config failed")
	//}

	if host := runtime.RemoteHost(); host.GetUser() != "root" {
		userConfigDirCmd := "mkdir -p $HOME/.kube"
		if _, err := runtime.GetRunner().Cmd(userConfigDirCmd, false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "user mkdir $HOME/.kube failed")
		}

		getKubeConfigCmdUsr := "cp -f /root/.kube/config $HOME/.kube/config"
		if _, err := runtime.GetRunner().SudoCmd(getKubeConfigCmdUsr, false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "user copy /etc/kubernetes/admin.conf to $HOME/.kube/config failed")
		}

		if _, err := runtime.GetRunner().SudoCmd("chmod 0600 $HOME/.kube/config", false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "chmod 0600 $HOME/.kube/config failed")
		}

		// userId, err := runtime.GetRunner().Cmd("echo $(id -u)", false, false)
		// if err != nil {
		// 	return errors.Wrap(errors.WithStack(err), "get user id failed")
		// }

		// userGroupId, err := runtime.GetRunner().Cmd("echo $(id -g)", false, false)
		// if err != nil {
		// 	return errors.Wrap(errors.WithStack(err), "get user group id failed")
		// }

		userId, err := runtime.GetRunner().Cmd("echo $SUDO_UID", false, false)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "get user id failed")
		}

		userGroupId, err := runtime.GetRunner().Cmd("echo $SUDO_GID", false, false)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "get user group id failed")
		}

		chownKubeConfig := fmt.Sprintf("chown -R %s:%s $HOME/.kube", userId, userGroupId)
		if _, err := runtime.GetRunner().SudoCmd(chownKubeConfig, false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "chown user kube config failed")
		}
	}
	return nil
}

type EnableRenewService struct {
	common.KubeAction
}

func (e *EnableRenewService) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"chmod +x /usr/local/bin/kube-scripts/k8s-certs-renew.sh && systemctl enable --now k8s-certs-renew.timer",
		false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "enable k8s renew certs service failed")
	}
	return nil
}

type UninstallAutoRenewCerts struct {
	common.KubeAction
}

func (u *UninstallAutoRenewCerts) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().SudoCmd("systemctl disable k8s-certs-renew.timer 1>/dev/null 2>/dev/null", false, false)
	_, _ = runtime.GetRunner().SudoCmd("systemctl stop k8s-certs-renew.timer 1>/dev/null 2>/dev/null", false, false)

	files := []string{
		filepath.Join("/usr/local/bin/kube-scripts/", templates.K8sCertsRenewScript.Name()),
		filepath.Join("/etc/systemd/system/", templates.K8sCertsRenewService.Name()),
		filepath.Join("/etc/systemd/system/", templates.K8sCertsRenewTimer.Name()),
	}
	for _, file := range files {
		_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", file), false, false)
	}

	return nil
}

type UninstallCertsFiles struct {
	common.KubeAction
}

func (t *UninstallCertsFiles) Execute(runtime connector.Runtime) error {
	var p = path.Join(runtime.GetBaseDir(), cc.Cli)
	if util.IsExist(p) {
		return util.RemoveDir(p)
	}
	return nil
}
