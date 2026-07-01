package entity

type CertStatus struct {
	WebDomain        string   `json:"webDomain"`
	HTTPSActive      bool     `json:"httpsActive"`
	MissingFiles     bool     `json:"missingFiles"`
	WebCertFile      string   `json:"webCertFile"`
	WebKeyFile       string   `json:"webKeyFile"`
	CertExists       bool     `json:"certExists"`
	KeyExists        bool     `json:"keyExists"`
	WebCertStatus    string   `json:"webCertStatus"`
	WebCertExpireAt  int64    `json:"webCertExpireAt"`
	WebCertIssuer    string   `json:"webCertIssuer"`
	WebCertAutoRenew bool     `json:"webCertAutoRenew"`
	WebCertMode      string   `json:"webCertMode"`
	WebCertProvider  string   `json:"webCertProvider"`
	Subject          string   `json:"subject"`
	DNSNames         []string `json:"dnsNames"`
	NotBefore        int64    `json:"notBefore"`
	NotAfter         int64    `json:"notAfter"`
	DaysLeft         int64    `json:"daysLeft"`
	IsExpired        bool     `json:"isExpired"`
	Warnings         []string `json:"warnings"`
}

type DomainCheckResult struct {
	Domain    string   `json:"domain"`
	Records   []string `json:"records"`
	ServerIPs []string `json:"serverIps"`
	Matched   bool     `json:"matched"`
	Warnings  []string `json:"warnings"`
}

type AcmeIssueResult struct {
	Domain  string      `json:"domain"`
	Staging bool        `json:"staging"`
	Applied bool        `json:"applied"`
	Status  *CertStatus `json:"status"`
}

type UsableCertificate struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Mode     string `json:"mode"`
	Domain   string `json:"domain"`
	Issuer   string `json:"issuer"`
	ExpireAt int64  `json:"expireAt"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
	Active   bool   `json:"active"`
}
