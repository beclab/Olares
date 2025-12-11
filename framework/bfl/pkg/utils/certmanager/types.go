package certmanager

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type DNSAddPayload struct {
	Name     string `json:"name"`
	PublicIP string `json:"public-ip,omitempty"`
	LocalIP  string `json:"local-ip,omitempty"`
	Domain   string `json:"domain,omitempty"`
}

type CustomDomainPayload struct {
	Name         string `json:"name"`
	CustomDomain string `json:"custom-host-name"`
}

type ResponseCert struct {
	Zone string `json:"zone" yaml:"zone"`
	Cert string `json:"cert" yaml:"cert"`
	Key  string `json:"key" yaml:"key"`

	// expired datetime
	ExpiredAt string `json:"enddate"`
}

type ResponseDownloadCert struct {
	Response
	Data *ResponseCert `json:"data"`
}

type ResponseCustomDomainStatus struct {
	CnameTarget    string `json:"cname-target,omitempty"`
	HostnameStatus string `json:"status-cname,omitempty"`
	SSLStatus      string `json:"status-ssl,omitempty"`
}

type ResponseAddCustomDomain struct {
	Response
	Data *ResponseCustomDomainStatus `json:"data,omitempty"`
}

type ResponseGetCustomDomain struct {
	Response
	Data *ResponseCustomDomainStatus `json:"data,omitempty"`
}

type ResponseDeleteCustomDomain struct {
	Response
}
