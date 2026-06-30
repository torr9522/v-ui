package web

import (
	"path/filepath"
	"testing"
	"x-ui/database"
	"x-ui/web/service"
)

func initWebTestDB(t *testing.T) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "web-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("init test db failed: %v", err)
	}
}

func TestResolveHTTPSConfigSelfHealsMissingCertificateFiles(t *testing.T) {
	initWebTestDB(t)

	settingService := &service.SettingService{}
	if err := settingService.SetCertFile("/tmp/not_exist.crt"); err != nil {
		t.Fatalf("set cert file failed: %v", err)
	}
	if err := settingService.SetKeyFile("/tmp/not_exist.key"); err != nil {
		t.Fatalf("set key file failed: %v", err)
	}
	if err := settingService.SetWebCertStatus("enabled"); err != nil {
		t.Fatalf("set web cert status failed: %v", err)
	}
	if err := settingService.SetWebCertMode("acme_http"); err != nil {
		t.Fatalf("set web cert mode failed: %v", err)
	}
	if err := settingService.SetWebCertProvider("letsencrypt"); err != nil {
		t.Fatalf("set web cert provider failed: %v", err)
	}
	if err := settingService.SetWebDomain("panel.example.com"); err != nil {
		t.Fatalf("set web domain failed: %v", err)
	}

	server := NewServer()
	certFile, keyFile, err := server.resolveHTTPSConfig()
	if err != nil {
		t.Fatalf("resolveHTTPSConfig failed: %v", err)
	}
	if certFile != "" || keyFile != "" {
		t.Fatalf("expected HTTP fallback, got cert=%q key=%q", certFile, keyFile)
	}

	currentCertFile, err := settingService.GetCertFile()
	if err != nil {
		t.Fatalf("get cert file failed: %v", err)
	}
	currentKeyFile, err := settingService.GetKeyFile()
	if err != nil {
		t.Fatalf("get key file failed: %v", err)
	}
	currentStatus, err := settingService.GetWebCertStatus()
	if err != nil {
		t.Fatalf("get web cert status failed: %v", err)
	}
	currentMode, err := settingService.GetWebCertMode()
	if err != nil {
		t.Fatalf("get web cert mode failed: %v", err)
	}
	currentProvider, err := settingService.GetWebCertProvider()
	if err != nil {
		t.Fatalf("get web cert provider failed: %v", err)
	}
	currentDomain, err := settingService.GetWebDomain()
	if err != nil {
		t.Fatalf("get web domain failed: %v", err)
	}

	if currentCertFile != "" || currentKeyFile != "" {
		t.Fatalf("expected certificate paths to be cleared, got cert=%q key=%q", currentCertFile, currentKeyFile)
	}
	if currentStatus != "missing" {
		t.Fatalf("expected webCertStatus=missing, got %q", currentStatus)
	}
	if currentMode != "none" {
		t.Fatalf("expected webCertMode=none, got %q", currentMode)
	}
	if currentProvider != "letsencrypt" {
		t.Fatalf("expected provider preserved, got %q", currentProvider)
	}
	if currentDomain != "panel.example.com" {
		t.Fatalf("expected domain preserved, got %q", currentDomain)
	}
}

func TestResolveHTTPSConfigAllowsHTTPWhenOnlyOnePathConfigured(t *testing.T) {
	initWebTestDB(t)

	settingService := &service.SettingService{}
	if err := settingService.SetCertFile("/tmp/not_exist.crt"); err != nil {
		t.Fatalf("set cert file failed: %v", err)
	}
	if err := settingService.SetKeyFile(""); err != nil {
		t.Fatalf("set key file failed: %v", err)
	}
	if err := settingService.SetWebCertStatus("enabled"); err != nil {
		t.Fatalf("set status failed: %v", err)
	}

	server := NewServer()
	certFile, keyFile, err := server.resolveHTTPSConfig()
	if err != nil {
		t.Fatalf("resolveHTTPSConfig failed: %v", err)
	}
	if certFile != "" || keyFile != "" {
		t.Fatalf("expected HTTP mode, got cert=%q key=%q", certFile, keyFile)
	}

	currentStatus, err := settingService.GetWebCertStatus()
	if err != nil {
		t.Fatalf("get status failed: %v", err)
	}
	if currentStatus != "enabled" {
		t.Fatalf("expected status unchanged for partial config, got %q", currentStatus)
	}
}
