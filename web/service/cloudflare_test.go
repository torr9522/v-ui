package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCloudflareVerifyTokenSuccess(t *testing.T) {
	initCertTestDB(t)
	setCloudflareTokenConfig(t, "token-123", "", "", false)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/tokens/verify" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		writeCloudflareJSON(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "token-id",
				"status": "active",
				"policies": [
					{"permission_groups": [{"name": "Zone:Read"}, {"name": "DNS:Read"}]}
				]
			}
		}`)
	}))
	defer server.Close()

	service := CloudflareService{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}
	result, err := service.VerifyToken()
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid token")
	}
	if result.AuthMode != "api_token" {
		t.Fatalf("expected api_token auth mode, got %s", result.AuthMode)
	}
	if !containsString(result.Permissions, "Zone:Read") {
		t.Fatalf("expected Zone:Read permission, got %v", result.Permissions)
	}
}

func TestCloudflareTestConnectionPermissionDenied(t *testing.T) {
	initCertTestDB(t)
	setCloudflareTokenConfig(t, "token-123", "", "", false)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user/tokens/verify":
			writeCloudflareJSON(w, `{"success": true, "errors": [], "messages": [], "result": {"id": "token-id", "status": "active"}}`)
		case "/zones":
			w.WriteHeader(http.StatusForbidden)
			writeCloudflareJSON(w, `{"success": false, "errors": [{"code": 9109, "message": "permission denied"}], "messages": [], "result": []}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	service := CloudflareService{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}
	_, err := service.TestConnection("panel.example.com")
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected permission denied error, got %v", err)
	}
}

func TestCloudflareTestConnectionZoneNotFound(t *testing.T) {
	initCertTestDB(t)
	setCloudflareTokenConfig(t, "token-123", "missing-zone", "", false)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user/tokens/verify":
			writeCloudflareJSON(w, `{"success": true, "errors": [], "messages": [], "result": {"id": "token-id", "status": "active"}}`)
		case "/zones":
			writeCloudflareJSON(w, `{"success": true, "errors": [], "messages": [], "result": [{"id": "zone-1", "name": "example.com", "status": "active", "account": {"id": "acc-1"}}], "result_info": {"page": 1, "per_page": 50, "total_pages": 1}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	service := CloudflareService{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}
	_, err := service.TestConnection("panel.example.com")
	if err == nil || !strings.Contains(err.Error(), "configured Cloudflare zone id was not found") {
		t.Fatalf("expected zone not found error, got %v", err)
	}
}

func TestCloudflareDetectZoneByDomainMismatch(t *testing.T) {
	initCertTestDB(t)
	setCloudflareTokenConfig(t, "token-123", "", "", false)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/zones" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		writeCloudflareJSON(w, `{"success": true, "errors": [], "messages": [], "result": [{"id": "zone-1", "name": "example.com", "status": "active", "account": {"id": "acc-1"}}], "result_info": {"page": 1, "per_page": 50, "total_pages": 1}}`)
	}))
	defer server.Close()

	service := CloudflareService{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}
	_, err := service.DetectZoneByDomain("panel.other.net")
	if err == nil || !strings.Contains(err.Error(), "domain does not belong to any Cloudflare zone") {
		t.Fatalf("expected domain mismatch error, got %v", err)
	}
}

func TestCloudflareGetDNSRecords(t *testing.T) {
	initCertTestDB(t)
	setCloudflareTokenConfig(t, "token-123", "zone-1", "", false)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/zones/zone-1/dns_records":
			if got := r.URL.Query().Get("type"); got != "TXT" {
				t.Fatalf("expected TXT query, got %s", got)
			}
			writeCloudflareJSON(w, `{"success": true, "errors": [], "messages": [], "result": [{"id": "rec-1", "type": "TXT", "name": "_acme-challenge.example.com", "content": "token", "ttl": 120}], "result_info": {"page": 1, "per_page": 100, "total_pages": 1}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	service := CloudflareService{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}
	records, err := service.GetDNSRecords("zone-1", "", "TXT")
	if err != nil {
		t.Fatalf("GetDNSRecords failed: %v", err)
	}
	if len(records) != 1 || records[0].ID != "rec-1" {
		t.Fatalf("unexpected records: %+v", records)
	}
}

func setCloudflareTokenConfig(t *testing.T, token string, zoneID string, webDomain string, enabled bool) {
	t.Helper()
	settingService := &SettingService{}
	if err := settingService.SetCloudflareAPIToken(token); err != nil {
		t.Fatalf("set token failed: %v", err)
	}
	if err := settingService.SetCloudflareZoneID(zoneID); err != nil {
		t.Fatalf("set zone id failed: %v", err)
	}
	if err := settingService.SetWebDomain(webDomain); err != nil {
		t.Fatalf("set web domain failed: %v", err)
	}
	if err := settingService.SetCloudflareEnabled(enabled); err != nil {
		t.Fatalf("set enabled failed: %v", err)
	}
	if err := settingService.SetCloudflareEmail(""); err != nil {
		t.Fatalf("clear email failed: %v", err)
	}
	if err := settingService.SetCloudflareAPIKey(""); err != nil {
		t.Fatalf("clear api key failed: %v", err)
	}
}

func writeCloudflareJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprint(w, body)
}
