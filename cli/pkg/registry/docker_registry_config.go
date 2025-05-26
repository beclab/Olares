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

package registry

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"bytetrade.io/web3os/installer/pkg/core/logger"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type DockerRegistryEntry struct {
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	SkipTLSVerify bool   `json:"skipTLSVerify,omitempty"`
	PlainHTTP     bool   `json:"plainHTTP,omitempty"`
	CertsPath     string `json:"certsPath,omitempty"`
	// CAFile is an SSL Certificate Authority file used to secure etcd communication.
	CAFile string `yaml:"caFile" json:"caFile,omitempty"`
	// CertFile is an SSL certification file used to secure etcd communication.
	CertFile string `yaml:"certFile" json:"certFile,omitempty"`
	// KeyFile is an SSL key file used to secure etcd communication.
	KeyFile string `yaml:"keyFile" json:"keyFile,omitempty"`
}

func DockerRegistryAuthEntries(auths runtime.RawExtension) (entries map[string]*DockerRegistryEntry) {

	if len(auths.Raw) == 0 {
		return
	}

	err := json.Unmarshal(auths.Raw, &entries)
	if err != nil {
		logger.Fatalf("Failed to Parse Registry Auths configuration: %v", auths.Raw)
		return
	}

	for _, v := range entries {
		if v.CertsPath != "" {
			ca, cert, key, err := LookupCertsFile(v.CertsPath)
			if err != nil {
				logger.Warnf("Failed to lookup certs file from the specific cert path %s: %s", v.CertsPath, err.Error())
				return
			}
			v.CAFile = ca
			v.CertFile = cert
			v.KeyFile = key
		}
		if v.PlainHTTP {
			v.SkipTLSVerify = true
		}
	}

	return
}

func LookupCertsFile(path string) (ca string, cert string, key string, err error) {
	logger.Debugf("Looking for TLS certificates and private keys in %s", path)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}
	logger.Debugf("Looking for TLS certificates and private keys in abs path %s", absPath)
	fs, err := ioutil.ReadDir(absPath)
	if err != nil {
		return ca, cert, key, err
	}

	for _, f := range fs {
		fullPath := filepath.Join(path, f.Name())
		if strings.HasSuffix(f.Name(), ".crt") {
			logger.Debugf(" crt: %s", fullPath)
			ca = fullPath
		}
		if strings.HasSuffix(f.Name(), ".cert") {
			certName := f.Name()
			keyName := certName[:len(certName)-5] + ".key"
			logger.Debugf(" cert: %s", fullPath)
			if !hasFile(fs, keyName) {
				return ca, cert, key, errors.Errorf("missing key %s for client certificate %s. Note that CA certificates should use the extension .crt", keyName, certName)
			}
			cert = fullPath
		}
		if strings.HasSuffix(f.Name(), ".key") {
			keyName := f.Name()
			certName := keyName[:len(keyName)-4] + ".cert"
			logger.Debugf(" key: %s", fullPath)
			if !hasFile(fs, certName) {
				return ca, cert, key, errors.Errorf("missing client certificate %s for key %s", certName, keyName)
			}
			key = fullPath
		}
	}
	return ca, cert, key, nil
}

func hasFile(files []os.FileInfo, name string) bool {
	for _, f := range files {
		if f.Name() == name {
			return true
		}
	}
	return false
}
