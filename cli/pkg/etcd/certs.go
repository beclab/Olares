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

package etcd

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/utils/certs"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/cert"
	certutil "k8s.io/client-go/util/cert"
	netutils "k8s.io/utils/net"
)

// KubekeyCertEtcdCA is the definition of the root CA used by the hosted etcd server.
func KubekeyCertEtcdCA() *certs.KubekeyCert {
	return &certs.KubekeyCert{
		Name:     "etcd-ca",
		LongName: "self-signed CA to provision identities for etcd",
		BaseName: "ca",
		Config: certs.CertConfig{
			Config: certutil.Config{
				CommonName: "etcd-ca",
			},
		},
	}
}

// KubekeyCertEtcdAdmin is the definition of the cert for etcd admin.
func KubekeyCertEtcdAdmin(hostname string, altNames *certutil.AltNames) *certs.KubekeyCert {
	l := strings.Split(hostname, ".")
	return &certs.KubekeyCert{
		Name:     "etcd-admin",
		LongName: "certificate for etcd admin",
		BaseName: fmt.Sprintf("admin-%s", hostname),
		CAName:   "etcd-ca",
		Config: certs.CertConfig{
			Config: certutil.Config{
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
				AltNames:   *altNames,
				CommonName: fmt.Sprintf("etcd-admin-%s", l[0]),
			},
		},
	}
}

// KubekeyCertEtcdMember is the definition of the cert for etcd member.
func KubekeyCertEtcdMember(hostname string, altNames *certutil.AltNames) *certs.KubekeyCert {
	l := strings.Split(hostname, ".")
	return &certs.KubekeyCert{
		Name:     "etcd-member",
		LongName: "certificate for etcd member",
		BaseName: fmt.Sprintf("member-%s", hostname),
		CAName:   "etcd-ca",
		Config: certs.CertConfig{
			Config: certutil.Config{
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
				AltNames:   *altNames,
				CommonName: fmt.Sprintf("etcd-member-%s", l[0]),
			},
		},
	}
}

// KubekeyCertEtcdClient is the definition of the cert for etcd client.
func KubekeyCertEtcdClient(hostname string, altNames *certutil.AltNames) *certs.KubekeyCert {
	l := strings.Split(hostname, ".")
	return &certs.KubekeyCert{
		Name:     "etcd-client",
		LongName: "certificate for etcd client",
		BaseName: fmt.Sprintf("node-%s", hostname),
		CAName:   "etcd-ca",
		Config: certs.CertConfig{
			Config: certutil.Config{
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
				AltNames:   *altNames,
				CommonName: fmt.Sprintf("etcd-node-%s", l[0]),
			},
		},
	}
}

type FetchCerts struct {
	common.KubeAction
}

func (f *FetchCerts) Execute(runtime connector.Runtime) error {
	src := "/etc/ssl/etcd/ssl"
	dst := fmt.Sprintf("%s/pki/etcd", runtime.GetWorkDir())

	v, ok := f.PipelineCache.Get(common.ETCDCluster)
	if !ok {
		return errors.New("get etcd status from pipeline cache failed")
	}

	c := v.(*EtcdCluster)

	if c.clusterExist {
		certs, err := runtime.GetRunner().SudoCmd("ls /etc/ssl/etcd/ssl/ | grep .pem", false, false)
		if err != nil {
			return errors.Wrap(err, "failed to find certificate files")
		}

		var split = "\r\n"
		if !strings.Contains(certs, split) && strings.Contains(certs, "\n") {
			split = "\n"
		}
		certsList := strings.Split(certs, split)
		if len(certsList) > 0 {
			for _, cert := range certsList {
				if err := runtime.GetRunner().Fetch(filepath.Join(dst, cert), filepath.Join(src, cert), false, false); err != nil {
					return errors.Wrap(err, fmt.Sprintf("Fetch %s failed", filepath.Join(src, cert)))
				}
			}
		}
	}

	return nil
}

type GenerateCerts struct {
	common.KubeAction
}

func (g *GenerateCerts) Execute(runtime connector.Runtime) error {
	pkiPath := fmt.Sprintf("%s/pki/etcd", runtime.GetWorkDir())

	altName := GenerateAltName(g.KubeConf, &runtime)

	files := []string{"ca.pem", "ca-key.pem"}

	// CA
	certsList := []*certs.KubekeyCert{KubekeyCertEtcdCA()}

	// Certs
	for _, host := range runtime.GetAllHosts() {
		if host.IsRole(common.ETCD) {
			certsList = append(certsList, KubekeyCertEtcdAdmin(host.GetName(), altName))
			files = append(files, []string{fmt.Sprintf("admin-%s.pem", host.GetName()), fmt.Sprintf("admin-%s-key.pem", host.GetName())}...)
			certsList = append(certsList, KubekeyCertEtcdMember(host.GetName(), altName))
			files = append(files, []string{fmt.Sprintf("member-%s.pem", host.GetName()), fmt.Sprintf("member-%s-key.pem", host.GetName())}...)
		}
		if host.IsRole(common.Master) {
			certsList = append(certsList, KubekeyCertEtcdClient(host.GetName(), altName))
			files = append(files, []string{fmt.Sprintf("node-%s.pem", host.GetName()), fmt.Sprintf("node-%s-key.pem", host.GetName())}...)
		}
	}

	var lastCACert *certs.KubekeyCert
	for _, c := range certsList {
		if c.CAName == "" {
			err := certs.GenerateCA(c, pkiPath, g.KubeConf)
			if err != nil {
				return err
			}
			lastCACert = c
		} else {
			err := certs.GenerateCerts(c, lastCACert, pkiPath, g.KubeConf)
			if err != nil {
				return err
			}
		}
	}

	g.ModuleCache.Set(LocalCertsDir, pkiPath)
	g.ModuleCache.Set(CertsFileList, files)

	return nil
}

func GenerateAltName(k *common.KubeConf, runtime *connector.Runtime) *cert.AltNames {
	var altName cert.AltNames

	dnsList := []string{"localhost", "etcd.kube-system.svc.cluster.local", "etcd.kube-system.svc", "etcd.kube-system", "etcd"}
	ipList := []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}

	if k.Cluster.ControlPlaneEndpoint.Domain == "" {
		dnsList = append(dnsList, kubekeyapiv1alpha2.DefaultLBDomain)
	} else {
		dnsList = append(dnsList, k.Cluster.ControlPlaneEndpoint.Domain)
	}

	for _, host := range k.Cluster.Hosts {
		dnsList = append(dnsList, host.Name)
		internalAddress := netutils.ParseIPSloppy(host.InternalAddress)
		if internalAddress != nil {
			ipList = append(ipList, internalAddress)
		}
	}

	altName.DNSNames = dnsList
	altName.IPs = ipList

	return &altName
}

type FetchCertsForExternalEtcd struct {
	common.KubeAction
}

func (f *FetchCertsForExternalEtcd) Execute(runtime connector.Runtime) error {
	pkiPath := fmt.Sprintf("%s/pki/etcd", runtime.GetWorkDir())

	if err := util.CreateDir(pkiPath); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create dir %s", pkiPath))
	}

	srcCertsFiles := []string{f.KubeConf.Cluster.Etcd.External.CAFile, f.KubeConf.Cluster.Etcd.External.CertFile, f.KubeConf.Cluster.Etcd.External.KeyFile}
	dstCertsFiles := []string{}
	for _, certFile := range srcCertsFiles {
		if len(certFile) != 0 {
			certPath, err := filepath.Abs(certFile)
			if err != nil {
				return errors.Wrap(err, "bad certificate file path")
			}
			_, err = os.Stat(certPath)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("%s does not exist", certPath))
			}

			dstCertFileName := filepath.Base(certPath)
			dstCert := fmt.Sprintf("%s/%s", pkiPath, dstCertFileName)
			dstCertsFiles = append(dstCertsFiles, dstCertFileName)

			data, err := ioutil.ReadFile(certPath)
			if err != nil {
				return errors.Wrap(err, "failed to copy certificate content")
			}

			if err := ioutil.WriteFile(dstCert, data, 0600); err != nil {
				return errors.Wrap(err, "failed to copy certificate content")
			}
		}
	}

	f.ModuleCache.Set(LocalCertsDir, pkiPath)
	f.ModuleCache.Set(CertsFileList, dstCertsFiles)

	return nil
}
