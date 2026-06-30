package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"x-ui/web/entity"
)

const defaultCloudflareAPIBaseURL = "https://api.cloudflare.com/client/v4"

type CloudflareService struct {
	settingService *SettingService
	baseURL        string
	httpClient     *http.Client
}

type cloudflareAPIError struct {
	StatusCode int
	Message    string
}

func (e *cloudflareAPIError) Error() string {
	return e.Message
}

type cloudflareEnvelope struct {
	Success    bool                `json:"success"`
	Errors     []cloudflareMessage `json:"errors"`
	Messages   []cloudflareMessage `json:"messages"`
	Result     json.RawMessage     `json:"result"`
	ResultInfo *cloudflarePageInfo `json:"result_info"`
}

type cloudflareMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type cloudflarePageInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

type cloudflareTokenVerifyResponse struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Policies []struct {
		PermissionGroups []struct {
			Name string `json:"name"`
		} `json:"permission_groups"`
	} `json:"policies"`
}

type cloudflareZoneResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Account struct {
		ID string `json:"id"`
	} `json:"account"`
}

type cloudflareDNSRecordResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

type cloudflareUserResponse struct {
	ID string `json:"id"`
}

func (s *CloudflareService) GetStatus() (*entity.CloudflareStatus, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	return &entity.CloudflareStatus{
		CloudflareConfig: *cfg,
		HasAPIToken:      strings.TrimSpace(cfg.APIToken) != "",
		HasAPIKey:        strings.TrimSpace(cfg.APIKey) != "",
	}, nil
}

func (s *CloudflareService) SaveConfig(cfg *entity.CloudflareConfig) error {
	if cfg == nil {
		return errors.New("cloudflare config is required")
	}
	settingService := s.getSettingService()
	return combineCloudflareSettingErrors(
		settingService.SetCloudflareAPIToken(cfg.APIToken),
		settingService.SetCloudflareZoneID(cfg.ZoneID),
		settingService.SetCloudflareAccountID(cfg.AccountID),
		settingService.SetCloudflareEmail(cfg.Email),
		settingService.SetCloudflareAPIKey(cfg.APIKey),
		settingService.SetCloudflareEnabled(cfg.Enabled),
	)
}

func (s *CloudflareService) VerifyToken() (*entity.CloudflareVerifyResult, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}

	switch cfg.AuthMode {
	case "api_token":
		result := &cloudflareTokenVerifyResponse{}
		if err := s.doRequest(cfg, http.MethodGet, "/user/tokens/verify", nil, nil, result); err != nil {
			return nil, err
		}
		permissions := make([]string, 0)
		for _, policy := range result.Policies {
			for _, group := range policy.PermissionGroups {
				if group.Name != "" {
					permissions = append(permissions, group.Name)
				}
			}
		}
		sort.Strings(permissions)
		return &entity.CloudflareVerifyResult{
			Valid:       strings.EqualFold(result.Status, "active"),
			AuthMode:    cfg.AuthMode,
			TokenID:     result.ID,
			Status:      result.Status,
			AccountID:   cfg.AccountID,
			Permissions: uniqueStrings(permissions),
			Warnings:    []string{},
		}, nil
	case "global_api_key":
		user := &cloudflareUserResponse{}
		if err := s.doRequest(cfg, http.MethodGet, "/user", nil, nil, user); err != nil {
			return nil, err
		}
		return &entity.CloudflareVerifyResult{
			Valid:       user.ID != "",
			AuthMode:    cfg.AuthMode,
			AccountID:   cfg.AccountID,
			Status:      "legacy",
			Permissions: []string{"Global API Key"},
			Warnings:    []string{"推荐使用 API Token 而不是 Global API Key。"},
		}, nil
	default:
		return nil, errors.New("cloudflare credentials are not configured")
	}
}

func (s *CloudflareService) GetZones() ([]*entity.CloudflareZone, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	return s.getZones(cfg)
}

func (s *CloudflareService) DetectZoneByDomain(domain string) (*entity.CloudflareZone, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	return s.detectZoneByDomain(cfg, domain)
}

func (s *CloudflareService) GetDNSRecords(zoneID string, name string, recordType string) ([]*entity.CloudflareDNSRecord, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	zoneID = strings.TrimSpace(zoneID)
	if zoneID == "" {
		zoneID = cfg.ZoneID
	}
	if zoneID == "" {
		return nil, errors.New("cloudflare zone id is required")
	}

	records := make([]*entity.CloudflareDNSRecord, 0)
	page := 1
	for {
		query := url.Values{}
		query.Set("page", fmt.Sprint(page))
		query.Set("per_page", "100")
		if name = strings.TrimSpace(name); name != "" {
			query.Set("name", name)
		}
		if recordType = strings.TrimSpace(recordType); recordType != "" {
			query.Set("type", recordType)
		}

		rawRecords := make([]*cloudflareDNSRecordResponse, 0)
		resultInfo, err := s.doRequestWithResultInfo(cfg, http.MethodGet, fmt.Sprintf("/zones/%s/dns_records", zoneID), query, nil, &rawRecords)
		if err != nil {
			return nil, err
		}
		for _, record := range rawRecords {
			records = append(records, &entity.CloudflareDNSRecord{
				ID:      record.ID,
				Type:    record.Type,
				Name:    record.Name,
				Content: record.Content,
				TTL:     record.TTL,
			})
		}
		if resultInfo == nil || page >= resultInfo.TotalPages || resultInfo.TotalPages == 0 {
			break
		}
		page++
	}
	return records, nil
}

func (s *CloudflareService) CreateTXTRecord(zoneID string, name string, content string) (*entity.CloudflareDNSRecord, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(zoneID) == "" {
		return nil, errors.New("cloudflare zone id is required")
	}
	body := map[string]interface{}{
		"type":    "TXT",
		"name":    strings.TrimSpace(name),
		"content": strings.TrimSpace(content),
		"ttl":     120,
	}
	result := &cloudflareDNSRecordResponse{}
	if err := s.doRequest(cfg, http.MethodPost, fmt.Sprintf("/zones/%s/dns_records", zoneID), nil, body, result); err != nil {
		return nil, err
	}
	return &entity.CloudflareDNSRecord{
		ID:      result.ID,
		Type:    result.Type,
		Name:    result.Name,
		Content: result.Content,
		TTL:     result.TTL,
	}, nil
}

func (s *CloudflareService) DeleteTXTRecord(zoneID string, recordID string) error {
	cfg, err := s.loadConfig()
	if err != nil {
		return err
	}
	if strings.TrimSpace(zoneID) == "" || strings.TrimSpace(recordID) == "" {
		return errors.New("cloudflare zone id and record id are required")
	}
	return s.doRequest(cfg, http.MethodDelete, fmt.Sprintf("/zones/%s/dns_records/%s", zoneID, recordID), nil, nil, nil)
}

func (s *CloudflareService) TestConnection(domain string) (*entity.CloudflareTestResult, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	verify, err := s.VerifyToken()
	if err != nil {
		return nil, err
	}
	zones, err := s.getZones(cfg)
	if err != nil {
		return nil, err
	}
	result := &entity.CloudflareTestResult{
		Valid:        verify.Valid,
		AuthMode:     cfg.AuthMode,
		CanReadZones: true,
		Permissions:  append([]string{}, verify.Permissions...),
		Warnings:     append([]string{}, verify.Warnings...),
		Verify:       verify,
	}

	var zone *entity.CloudflareZone
	if cfg.ZoneID != "" {
		for _, item := range zones {
			if item.ID == cfg.ZoneID {
				zone = item
				break
			}
		}
		if zone == nil {
			return nil, errors.New("configured Cloudflare zone id was not found")
		}
	} else {
		zone, err = s.detectZoneByDomainWithZones(strings.TrimSpace(domain), zones)
		if err != nil {
			return nil, err
		}
	}

	result.ZoneID = zone.ID
	result.ZoneName = zone.Name

	if _, err := s.GetDNSRecords(zone.ID, "", ""); err != nil {
		return nil, err
	}
	result.CanReadRecords = true
	return result, nil
}

func (s *CloudflareService) loadConfig() (*entity.CloudflareConfig, error) {
	settingService := s.getSettingService()
	apiToken, err := settingService.GetCloudflareAPIToken()
	if err != nil {
		return nil, err
	}
	zoneID, err := settingService.GetCloudflareZoneID()
	if err != nil {
		return nil, err
	}
	accountID, err := settingService.GetCloudflareAccountID()
	if err != nil {
		return nil, err
	}
	email, err := settingService.GetCloudflareEmail()
	if err != nil {
		return nil, err
	}
	apiKey, err := settingService.GetCloudflareAPIKey()
	if err != nil {
		return nil, err
	}
	enabled, err := settingService.GetCloudflareEnabled()
	if err != nil {
		return nil, err
	}
	webDomain, err := settingService.GetWebDomain()
	if err != nil {
		return nil, err
	}

	cfg := &entity.CloudflareConfig{
		APIToken:  strings.TrimSpace(apiToken),
		ZoneID:    strings.TrimSpace(zoneID),
		AccountID: strings.TrimSpace(accountID),
		Email:     strings.TrimSpace(email),
		APIKey:    strings.TrimSpace(apiKey),
		Enabled:   enabled,
		WebDomain: strings.TrimSpace(webDomain),
	}
	cfg.AuthMode = s.resolveAuthMode(cfg)
	return cfg, nil
}

func (s *CloudflareService) getZones(cfg *entity.CloudflareConfig) ([]*entity.CloudflareZone, error) {
	zones := make([]*entity.CloudflareZone, 0)
	page := 1
	for {
		query := url.Values{}
		query.Set("page", fmt.Sprint(page))
		query.Set("per_page", "50")
		rawZones := make([]*cloudflareZoneResponse, 0)
		resultInfo, err := s.doRequestWithResultInfo(cfg, http.MethodGet, "/zones", query, nil, &rawZones)
		if err != nil {
			return nil, err
		}
		for _, zone := range rawZones {
			zones = append(zones, &entity.CloudflareZone{
				ID:        zone.ID,
				Name:      zone.Name,
				Status:    zone.Status,
				AccountID: zone.Account.ID,
			})
		}
		if resultInfo == nil || page >= resultInfo.TotalPages || resultInfo.TotalPages == 0 {
			break
		}
		page++
	}
	sort.Slice(zones, func(i, j int) bool {
		return zones[i].Name < zones[j].Name
	})
	return zones, nil
}

func (s *CloudflareService) detectZoneByDomain(cfg *entity.CloudflareConfig, domain string) (*entity.CloudflareZone, error) {
	zones, err := s.getZones(cfg)
	if err != nil {
		return nil, err
	}
	return s.detectZoneByDomainWithZones(domain, zones)
}

func (s *CloudflareService) detectZoneByDomainWithZones(domain string, zones []*entity.CloudflareZone) (*entity.CloudflareZone, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	domain = strings.TrimSuffix(domain, ".")
	if domain == "" {
		return nil, errors.New("domain is required for zone detection")
	}

	var matched *entity.CloudflareZone
	for _, zone := range zones {
		name := strings.ToLower(strings.TrimSpace(zone.Name))
		if domain == name || strings.HasSuffix(domain, "."+name) {
			if matched == nil || len(zone.Name) > len(matched.Name) {
				matched = zone
			}
		}
	}
	if matched == nil {
		return nil, errors.New("domain does not belong to any Cloudflare zone")
	}
	return matched, nil
}

func (s *CloudflareService) resolveAuthMode(cfg *entity.CloudflareConfig) string {
	if cfg == nil {
		return "none"
	}
	if strings.TrimSpace(cfg.APIToken) != "" {
		return "api_token"
	}
	if strings.TrimSpace(cfg.Email) != "" && strings.TrimSpace(cfg.APIKey) != "" {
		return "global_api_key"
	}
	return "none"
}

func (s *CloudflareService) getSettingService() *SettingService {
	if s.settingService != nil {
		return s.settingService
	}
	return &SettingService{}
}

func (s *CloudflareService) getHTTPClient() *http.Client {
	if s.httpClient != nil {
		return s.httpClient
	}
	return http.DefaultClient
}

func (s *CloudflareService) getBaseURL() string {
	if strings.TrimSpace(s.baseURL) != "" {
		return strings.TrimRight(strings.TrimSpace(s.baseURL), "/")
	}
	return defaultCloudflareAPIBaseURL
}

func (s *CloudflareService) doRequest(cfg *entity.CloudflareConfig, method string, path string, query url.Values, body interface{}, out interface{}) error {
	_, err := s.doRequestWithResultInfo(cfg, method, path, query, body, out)
	return err
}

func (s *CloudflareService) doRequestWithResultInfo(cfg *entity.CloudflareConfig, method string, path string, query url.Values, body interface{}, out interface{}) (*cloudflarePageInfo, error) {
	if cfg == nil {
		return nil, errors.New("cloudflare config is required")
	}
	if cfg.AuthMode == "none" {
		return nil, errors.New("cloudflare credentials are not configured")
	}

	baseURL := s.getBaseURL()
	requestURL := baseURL + path
	if len(query) > 0 {
		requestURL += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, requestURL, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	switch cfg.AuthMode {
	case "api_token":
		req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	case "global_api_key":
		req.Header.Set("X-Auth-Email", cfg.Email)
		req.Header.Set("X-Auth-Key", cfg.APIKey)
	default:
		return nil, errors.New("cloudflare credentials are not configured")
	}

	resp, err := s.getHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	envelope := &cloudflareEnvelope{}
	if err := json.Unmarshal(bodyBytes, envelope); err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest || !envelope.Success {
		return nil, &cloudflareAPIError{
			StatusCode: resp.StatusCode,
			Message:    formatCloudflareError(envelope, resp.StatusCode),
		}
	}
	if out != nil && len(envelope.Result) > 0 {
		if err := json.Unmarshal(envelope.Result, out); err != nil {
			return nil, err
		}
	}
	return envelope.ResultInfo, nil
}

func formatCloudflareError(envelope *cloudflareEnvelope, statusCode int) string {
	messages := make([]string, 0, len(envelope.Errors)+len(envelope.Messages))
	for _, item := range envelope.Errors {
		if item.Message != "" {
			messages = append(messages, item.Message)
		}
	}
	for _, item := range envelope.Messages {
		if item.Message != "" {
			messages = append(messages, item.Message)
		}
	}
	if len(messages) == 0 {
		return fmt.Sprintf("cloudflare api request failed with status %d", statusCode)
	}
	return strings.Join(uniqueStrings(messages), "; ")
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func combineCloudflareSettingErrors(errs ...error) error {
	filtered := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			filtered = append(filtered, err)
		}
	}
	return errors.Join(filtered...)
}
