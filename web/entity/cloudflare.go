package entity

type CloudflareConfig struct {
	APIToken  string `json:"apiToken" form:"apiToken"`
	ZoneID    string `json:"zoneId" form:"zoneId"`
	AccountID string `json:"accountId" form:"accountId"`
	Email     string `json:"email" form:"email"`
	APIKey    string `json:"apiKey" form:"apiKey"`
	Enabled   bool   `json:"enabled" form:"enabled"`
	WebDomain string `json:"webDomain" form:"webDomain"`
	AuthMode  string `json:"authMode" form:"authMode"`
}

type CloudflareStatus struct {
	CloudflareConfig
	HasAPIToken bool `json:"hasApiToken"`
	HasAPIKey   bool `json:"hasApiKey"`
}

type CloudflareVerifyResult struct {
	Valid       bool     `json:"valid"`
	AuthMode    string   `json:"authMode"`
	TokenID     string   `json:"tokenId"`
	Status      string   `json:"status"`
	AccountID   string   `json:"accountId"`
	Permissions []string `json:"permissions"`
	Warnings    []string `json:"warnings"`
}

type CloudflareZone struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	AccountID string `json:"accountId"`
}

type CloudflareDNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

type CloudflareTestResult struct {
	Valid          bool                    `json:"valid"`
	AuthMode       string                  `json:"authMode"`
	ZoneID         string                  `json:"zoneId"`
	ZoneName       string                  `json:"zoneName"`
	CanReadZones   bool                    `json:"canReadZones"`
	CanReadRecords bool                    `json:"canReadRecords"`
	Permissions    []string                `json:"permissions"`
	Warnings       []string                `json:"warnings"`
	Verify         *CloudflareVerifyResult `json:"verify"`
}
