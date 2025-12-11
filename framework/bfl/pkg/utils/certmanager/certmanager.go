package certmanager

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1/operator"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils"

	"k8s.io/klog/v2"
)

type Interface interface {
	// GenerateCert generate cert, and wait for complete
	GenerateCert() error

	// DownloadCert download the cert, when generate is complete
	DownloadCert() (*ResponseCert, error)

	// AddDNSRecord add dns record, ip or domain
	AddDNSRecord(publicIP, domain *string) error

	// DeleteDNSRecord delete dns record, delete zone domain record
	DeleteDNSRecord() error

	AddCustomDomainOnCloudflare(customDomain string) (*ResponseCustomDomainStatus, error)

	GetCustomDomainOnCloudflare(customDomain string) (*ResponseCustomDomainStatus, error)
	GetCustomDomainCnameStatus(customDomain string) (*Response, error)

	DeleteCustomDomainOnCloudflare(customDomain string) (*Response, error)

	GetCustomDomainErrorStatus(err error) string
}

type certManager struct {
	terminusName constants.TerminusName

	httpClientTimeout   time.Duration
	waitGenerateTimeout time.Duration
}

var _ Interface = &certManager{}

func NewCertManager(terminusName constants.TerminusName) Interface {
	c := &certManager{terminusName: terminusName}

	c.httpClientTimeout = 30 * time.Second
	c.waitGenerateTimeout = 5 * time.Minute

	return c
}

func (c *certManager) GenerateCert() error {
	err := c.request(context.TODO(), "GET", fmt.Sprintf(constants.APIFormatCertGenerateRequest, c.terminusName), nil, nil)
	if err != nil {
		return err
	}

	t := time.After(c.waitGenerateTimeout)

	for {
		err = c.request(context.TODO(), "GET", fmt.Sprintf(constants.APIFormatCertGenerateStatus, c.terminusName), nil, nil)
		if err == nil {
			break
		}

		select {
		case <-t:
			return fmt.Errorf("timeout in %v minutes", c.waitGenerateTimeout.Minutes())
		default:
			time.Sleep(2 * time.Second)
		}
	}
	return nil
}

func (c *certManager) DownloadCert() (*ResponseCert, error) {
	var r ResponseDownloadCert

	err := c.request(context.TODO(), "GET", fmt.Sprintf(constants.APIFormatCertDownload, c.terminusName), nil, &r)
	if err != nil {
		return nil, err
	}

	if r.Data.Cert == "" || r.Data.Key == "" || r.Data.Zone == "" {
		return nil, fmt.Errorf("unexpected error, field 'zone' or 'cert', 'key' empty")
	}
	return r.Data, nil
}

func (c *certManager) AddDNSRecord(publicIP, domain *string) error {
	payload := DNSAddPayload{Name: string(c.terminusName)}
	if publicIP != nil {
		payload.PublicIP = *publicIP
	}

	if domain != nil {
		payload.Domain = *domain
	}

	return c.request(context.TODO(), "POST", constants.APIDNSAddRecord, payload, nil)
}

func (c *certManager) DeleteDNSRecord() error {
	return c.request(context.TODO(), "GET", fmt.Sprintf(constants.APIFormatDNSDeleteRecord, c.terminusName), nil, nil)
}

func (c *certManager) getCertManagerJWSToken() (string, error) {
	op, err := operator.NewUserOperator()
	if err != nil {
		return "", err
	}
	user, err := op.GetUser(constants.Username)
	if err != nil {
		return "", err
	}

	jws := op.GetUserAnnotation(user, constants.UserCertManagerJWSToken)
	if jws == "" {
		return "", fmt.Errorf("cert manager JWS not exists")
	}
	return jws, nil
}

func (c *certManager) request(ctx context.Context, method, url string, body, to any) error {
	var (
		err      error
		reqBytes []byte
		br       io.Reader
	)

	if body != nil {
		reqBytes, err = json.Marshal(body)
		if err != nil {
			return err
		}
		br = bytes.NewReader(reqBytes)
	}

	var req *http.Request
	req, err = http.NewRequest(method, url, br)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Accept", "application/json")
	if utils.ListContains([]string{"POST", "PUT", "PATCH", "DELETE"}, method) && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	jws, err := c.getCertManagerJWSToken()
	if err != nil {
		klog.Warningf("no JWS signature provided")
	} else {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jws))
	}

	klog.Infof("requesting api: %s, method: %s, headers: '%s', payload: '%s'", url, method, utils.ToJSON(req.Header), string(reqBytes))

	var resp *http.Response
	resp, err = (&http.Client{Timeout: c.httpClientTimeout}).Do(req)
	if err != nil {
		klog.Errorf("requesting api: got error, %v", err)
		return err
	}

	var respBytes []byte
	respBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("requesting api: read response body err, %v", err)
		return err
	}
	defer resp.Body.Close()

	klog.Infof("request response code: %v, body: %v", resp.StatusCode, string(respBytes))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response code: %v", resp.StatusCode)
	}

	// check response
	var r Response
	if err = json.Unmarshal(respBytes, &r); err != nil {
		klog.Errorf("requesting api: unmarshal response err, %v", err)
		return err
	}

	if !r.Success {
		return fmt.Errorf("unexpected response, err: '%v'", r.Message)
	}

	if to != nil {
		return json.Unmarshal(respBytes, to)
	}

	return nil
}

func ValidPemKey(key string) error {
	keyPemBlock, _ := pem.Decode([]byte(key))
	if keyPemBlock == nil {
		return fmt.Errorf("not pem format key")
	}

	_, err := x509.ParsePKCS8PrivateKey(keyPemBlock.Bytes)
	if err != nil {
		_, err = x509.ParsePKCS1PrivateKey(keyPemBlock.Bytes)
		if err != nil {
			_, err = x509.ParseECPrivateKey(keyPemBlock.Bytes)
			if err != nil {
				return fmt.Errorf("parse pkcs private key error %v", err)
			}
		}
	}

	return nil
}

func ValidPemCert(cert string) error {
	pemBlock, _ := pem.Decode([]byte(cert))
	if pemBlock == nil {
		return fmt.Errorf("not pem format cert")
	}

	certs, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return fmt.Errorf("parse cert error %v", err)
	}

	roots := x509.NewCertPool()

	roots.AppendCertsFromPEM([]byte(cert))
	opts := x509.VerifyOptions{
		Roots: roots,
	}

	_, err = certs.Verify(opts)
	if err != nil {
		return fmt.Errorf("verify cert error %v", err)
	}

	return nil
}

func (c *certManager) AddCustomDomainOnCloudflare(customDomain string) (*ResponseCustomDomainStatus, error) {
	payload := CustomDomainPayload{Name: string(c.terminusName), CustomDomain: customDomain}
	var r ResponseAddCustomDomain
	err := c.request(context.TODO(), "POST", constants.APIDNSAddCustomDomain, payload, &r)
	if err != nil && !strings.Contains(err.Error(), "409 Conflict") {
		return nil, err
	}
	return r.Data, nil
}

func (c *certManager) GetCustomDomainOnCloudflare(customDomain string) (*ResponseCustomDomainStatus, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var r ResponseGetCustomDomain
	var url = fmt.Sprintf("%s?name=%s&custom-host-name=%s", constants.APIDNSAddCustomDomain, c.terminusName, customDomain)
	err := c.request(ctx, "GET", url, nil, &r)
	if err != nil {
		return nil, err
	}
	return r.Data, ctx.Err()
}

func (c *certManager) DeleteCustomDomainOnCloudflare(customDomain string) (*Response, error) {
	payload := CustomDomainPayload{Name: string(c.terminusName), CustomDomain: customDomain}
	var r Response
	err := c.request(context.Background(), "DELETE", constants.APIDNSSetCloudFlareTunnel, payload, &r)
	if err != nil && !strings.Contains(err.Error(), "The custom hostname was not found") {
		return nil, err
	}
	return &r, nil
}

func (c *certManager) GetCustomDomainCnameStatus(customDomain string) (*Response, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payload := CustomDomainPayload{Name: string(c.terminusName), CustomDomain: customDomain}
	var r Response

	err := c.request(ctx, "POST", constants.APIDNSCheckCustomDomainCname, payload, &r)
	if err != nil && !strings.Contains(err.Error(), "cname configuration error") {
		return nil, err
	}
	return &r, nil
}

func (c *certManager) GetCustomDomainErrorStatus(err error) string {
	var msg = err.Error()
	switch {
	case strings.Contains(msg, "The custom hostname was not found"):
		return constants.CustomDomainCnameStatusNone
	case strings.Contains(msg, "context deadline exceeded"):
		return constants.CustomDomainCnameStatusTimeout
	default:
		return constants.CustomDomainCnameStatusError
	}
}
