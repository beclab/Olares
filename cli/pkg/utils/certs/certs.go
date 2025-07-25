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
	"crypto"
	cryptorand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math"
	"math/big"
	"net"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/logger"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	certutil "k8s.io/client-go/util/cert"
)

const (
	// CertificateValidity defines the validity for all the signed certificates generated by kubeadm
	CertificateValidity = time.Hour * 24 * 365 * 100
	// CertificateBlockType is a possible value for pem.Block.Type.
	CertificateBlockType = "CERTIFICATE"
	rsaKeySize           = 2048
)

// KubekeyCert represents a certificate that Kubeadm will create to function properly.
type KubekeyCert struct {
	Name     string
	LongName string
	BaseName string
	CAName   string
	Config   CertConfig
}

// GetConfig returns the definition for the given cert given the provided InitConfiguration
func (k *KubekeyCert) GetConfig(_ *common.KubeConf) (*CertConfig, error) {

	return &k.Config, nil
}

// CreateFromCA makes and writes a certificate using the given CA cert and key.
func (k *KubekeyCert) CreateFromCA(kubeConf *common.KubeConf, pkiPath string, caCert *x509.Certificate, caKey crypto.Signer) error {
	cfg, err := k.GetConfig(kubeConf)
	if err != nil {
		return errors.Wrapf(err, "couldn't create %q certificate", k.Name)
	}
	cert, key, err := NewCertAndKey(caCert, caKey, cfg)
	if err != nil {
		return err
	}
	err = writeCertificateFilesIfNotExist(
		pkiPath,
		k.BaseName,
		caCert,
		cert,
		key,
		cfg,
	)

	if err != nil {
		return errors.Wrapf(err, "failed to write or validate certificate %q", k.Name)
	}

	return nil
}

func GenerateCA(ca *KubekeyCert, pkiPath string, kubeConf *common.KubeConf) error {

	if cert, err := TryLoadCertFromDisk(pkiPath, ca.BaseName); err == nil {
		CheckCertificatePeriodValidity(ca.BaseName, cert)

		if _, err := TryLoadKeyFromDisk(pkiPath, ca.BaseName); err == nil {
			fmt.Printf("[certs] Using existing %s certificate authority\n", ca.BaseName)
			return nil
		}
		fmt.Printf("[certs] Using existing %s keyless certificate authority\n", ca.BaseName)
		return nil
	}

	// create the new certificate authority (or use existing)
	return CreateCACertAndKeyFiles(ca, pkiPath, kubeConf)

}

// CreateCACertAndKeyFiles generates and writes out a given certificate authority.
// The certSpec should be one of the variables from this package.
func CreateCACertAndKeyFiles(certSpec *KubekeyCert, pkiPath string, kubeConf *common.KubeConf) error {
	if certSpec.CAName != "" {
		return errors.Errorf("this function should only be used for CAs, but cert %s has CA %s", certSpec.Name, certSpec.CAName)
	}

	certConfig, err := certSpec.GetConfig(kubeConf)
	if err != nil {
		return err
	}

	caCert, caKey, err := NewCertificateAuthority(certConfig)
	if err != nil {
		return err
	}

	return writeCertificateAuthorityFilesIfNotExist(
		pkiPath,
		certSpec.BaseName,
		caCert,
		caKey,
	)
}

func GenerateCerts(cert *KubekeyCert, caCert *KubekeyCert, pkiPath string, kubeConf *common.KubeConf) error {
	// TODO: if using external etcd, skips etcd certificates generation

	if certData, intermediates, err := TryLoadCertChainFromDisk(pkiPath, cert.BaseName); err == nil {
		CheckCertificatePeriodValidity(cert.BaseName, certData)

		caCertData, err := TryLoadCertFromDisk(pkiPath, caCert.BaseName)
		if err != nil {
			return errors.Wrapf(err, "couldn't load CA certificate %s", caCert.Name)
		}

		CheckCertificatePeriodValidity(caCert.BaseName, caCertData)

		if err := VerifyCertChain(certData, intermediates, caCertData); err != nil {
			return errors.Wrapf(err, "[certs] certificate %s not signed by CA certificate %s", cert.BaseName, caCert.BaseName)
		}

		fmt.Printf("[certs] Using existing %s certificate and key on disk\n", cert.BaseName)
		return nil
	}

	// create the new certificate (or use existing)
	return CreateCertAndKeyFilesWithCA(caCert, cert, pkiPath, kubeConf)
}

// CreateCertAndKeyFilesWithCA loads the given certificate authority from disk, then generates and writes out the given certificate and key.
// The certSpec and caCertSpec should both be one of the variables from this package.
func CreateCertAndKeyFilesWithCA(caCertSpec *KubekeyCert, certSpec *KubekeyCert, pkiPath string, kubeConf *common.KubeConf) error {
	if certSpec.CAName != caCertSpec.Name {
		return errors.Errorf("expected CAname for %s to be %q, but was %s", certSpec.Name, certSpec.CAName, caCertSpec.Name)
	}

	caCert, caKey, err := LoadCertificateAuthority(pkiPath, caCertSpec.BaseName)
	if err != nil {
		return errors.Wrapf(err, "couldn't load CA certificate %s", caCertSpec.Name)
	}

	return certSpec.CreateFromCA(kubeConf, pkiPath, caCert, caKey)
}

// LoadCertificateAuthority tries to load a CA in the given directory with the given name.
func LoadCertificateAuthority(pkiDir string, baseName string) (*x509.Certificate, crypto.Signer, error) {
	// Checks if certificate authority exists in the PKI directory
	if !CertOrKeyExist(pkiDir, baseName) {
		return nil, nil, errors.Errorf("couldn't load %s certificate authority from %s", baseName, pkiDir)
	}

	// Try to load certificate authority .crt and .key from the PKI directory
	caCert, caKey, err := TryLoadCertAndKeyFromDisk(pkiDir, baseName)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failure loading %s certificate authority", baseName)
	}
	// Validate period
	CheckCertificatePeriodValidity(baseName, caCert)

	// Make sure the loaded CA cert actually is a CA
	if !caCert.IsCA {
		return nil, nil, errors.Errorf("%s certificate is not a certificate authority", baseName)
	}

	return caCert, caKey, nil
}

// NewCertAndKey creates new certificate and key by passing the certificate authority certificate and key
func NewCertAndKey(caCert *x509.Certificate, caKey crypto.Signer, config *CertConfig) (*x509.Certificate, crypto.Signer, error) {
	if len(config.Usages) == 0 {
		return nil, nil, errors.New("must specify at least one ExtKeyUsage")
	}

	key, err := NewPrivateKey(config.PublicKeyAlgorithm)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to create private key")
	}

	cert, err := NewSignedCert(config, key, caCert, caKey, false)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to sign certificate")
	}

	return cert, key, nil
}

// NewSignedCert creates a signed certificate using the given CA certificate and key
func NewSignedCert(cfg *CertConfig, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer, isCA bool) (*x509.Certificate, error) {
	serial, err := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	if len(cfg.CommonName) == 0 {
		return nil, errors.New("must specify a CommonName")
	}

	keyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	if isCA {
		keyUsage |= x509.KeyUsageCertSign
	}

	RemoveDuplicateAltNames(&cfg.AltNames)

	notAfter := time.Now().Add(CertificateValidity).UTC()
	if cfg.NotAfter != nil {
		notAfter = *cfg.NotAfter
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:              cfg.AltNames.DNSNames,
		IPAddresses:           cfg.AltNames.IPs,
		SerialNumber:          serial,
		NotBefore:             caCert.NotBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           cfg.Usages,
		BasicConstraintsValid: true,
		IsCA:                  isCA,
	}
	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}

// RemoveDuplicateAltNames removes duplicate items in altNames.
func RemoveDuplicateAltNames(altNames *certutil.AltNames) {
	if altNames == nil {
		return
	}

	if altNames.DNSNames != nil {
		altNames.DNSNames = sets.NewString(altNames.DNSNames...).List()
	}

	ipsKeys := make(map[string]struct{})
	var ips []net.IP
	for _, one := range altNames.IPs {
		if _, ok := ipsKeys[one.String()]; !ok {
			ipsKeys[one.String()] = struct{}{}
			ips = append(ips, one)
		}
	}
	altNames.IPs = ips
}

// CheckCertificatePeriodValidity takes a certificate and prints a warning if its period
// is not valid related to the current time. It does so only if the certificate was not validated already
// by keeping track with a cache.
func CheckCertificatePeriodValidity(baseName string, cert *x509.Certificate) {
	certPeriodValidationMutex.Lock()
	defer certPeriodValidationMutex.Unlock()
	if _, exists := certPeriodValidation[baseName]; exists {
		return
	}
	certPeriodValidation[baseName] = struct{}{}

	if err := ValidateCertPeriod(cert, 0); err != nil {
		logger.Warnf("WARNING: could not validate bounds for certificate %s: %v", baseName, err)
	}
}

// writeCertificateFilesIfNotExist write a new certificate to the given path.
// If there already is a certificate file at the given path; kubeadm tries to load it and check if the values in the
// existing and the expected certificate equals. If they do; kubeadm will just skip writing the file as it's up-to-date,
// otherwise this function returns an error.
func writeCertificateFilesIfNotExist(pkiDir string, baseName string, signingCert *x509.Certificate, cert *x509.Certificate, key crypto.Signer, cfg *CertConfig) error {

	// Checks if the signed certificate exists in the PKI directory
	if CertOrKeyExist(pkiDir, baseName) {
		// Try to load key from the PKI directory
		_, err := TryLoadKeyFromDisk(pkiDir, baseName)
		if err != nil {
			return errors.Wrapf(err, "failure loading %s key", baseName)
		}

		// Try to load certificate from the PKI directory
		signedCert, intermediates, err := TryLoadCertChainFromDisk(pkiDir, baseName)
		if err != nil {
			return errors.Wrapf(err, "failure loading %s certificate", baseName)
		}
		// Validate period
		CheckCertificatePeriodValidity(baseName, signedCert)

		// Check if the existing cert is signed by the given CA
		if err := VerifyCertChain(signedCert, intermediates, signingCert); err != nil {
			return errors.Errorf("certificate %s is not signed by corresponding CA", baseName)
		}

		// Check if the certificate has the correct attributes
		if err := validateCertificateWithConfig(signedCert, baseName, cfg); err != nil {
			return err
		}

		fmt.Printf("[certs] Using the existing %q certificate and key\n", baseName)
	} else {
		if err := WriteCertAndKey(pkiDir, baseName, cert, key); err != nil {
			return errors.Wrapf(err, "failure while saving %s certificate and key", baseName)
		}
		if HasServerAuth(cert) {
			fmt.Printf("[certs] %s serving cert is signed for DNS names %v and IPs %v\n", baseName, cert.DNSNames, cert.IPAddresses)
		}
	}

	return nil
}
