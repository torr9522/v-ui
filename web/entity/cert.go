package entity

type CertStatus struct {
	WebDomain        string   `json:"webDomain"`
	WebCertFile      string   `json:"webCertFile"`
	WebKeyFile       string   `json:"webKeyFile"`
	WebCertStatus    string   `json:"webCertStatus"`
	WebCertExpireAt  int64    `json:"webCertExpireAt"`
	WebCertIssuer    string   `json:"webCertIssuer"`
	WebCertAutoRenew bool     `json:"webCertAutoRenew"`
	WebCertMode      string   `json:"webCertMode"`
	WebCertProvider  string   `json:"webCertProvider"`
	Warnings         []string `json:"warnings"`
}

type DomainCheckResult struct {
	Domain    string   `json:"domain"`
	Records   []string `json:"records"`
	ServerIPs []string `json:"serverIps"`
	Matched   bool     `json:"matched"`
	Warnings  []string `json:"warnings"`
}
