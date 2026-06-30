package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
	"x-ui/database"
)

func initCertTestDB(t *testing.T) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "cert-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("init test db failed: %v", err)
	}
}

func TestParseCertificatePair(t *testing.T) {
	certPEM, keyPEM, err := generateSelfSignedCert([]string{"panel.example.com"}, time.Now().Add(48*time.Hour))
	if err != nil {
		t.Fatalf("generate cert failed: %v", err)
	}

	info, err := parseCertificatePair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("parse certificate pair failed: %v", err)
	}
	if info.subject != "panel.example.com" {
		t.Fatalf("unexpected subject: %s", info.subject)
	}
	if !containsString(info.dnsNames, "panel.example.com") {
		t.Fatalf("expected dnsNames to include panel.example.com")
	}
}

func TestParseCertificatePairRejectsMismatchedKey(t *testing.T) {
	certPEM, _, err := generateSelfSignedCert([]string{"panel.example.com"}, time.Now().Add(48*time.Hour))
	if err != nil {
		t.Fatalf("generate cert failed: %v", err)
	}
	_, wrongKeyPEM, err := generateSelfSignedCert([]string{"other.example.com"}, time.Now().Add(48*time.Hour))
	if err != nil {
		t.Fatalf("generate wrong key failed: %v", err)
	}

	if _, err := parseCertificatePair(certPEM, wrongKeyPEM); err == nil {
		t.Fatalf("expected mismatched key validation to fail")
	}
}

func TestReadCertificateInfoParsesMetadata(t *testing.T) {
	certPEM, keyPEM, err := generateSelfSignedCert([]string{"panel.example.com", "alt.example.com"}, time.Now().Add(72*time.Hour))
	if err != nil {
		t.Fatalf("generate cert failed: %v", err)
	}

	info, err := readCertificateInfo(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("read certificate info failed: %v", err)
	}
	if info.issuer == "" {
		t.Fatalf("expected issuer to be populated")
	}
	if !containsString(info.dnsNames, "alt.example.com") {
		t.Fatalf("expected alt.example.com in dnsNames")
	}
	if info.expireAt <= info.notBefore {
		t.Fatalf("expected expireAt > notBefore")
	}
}

func TestUploadCertificateWritesFilesAndPermissions(t *testing.T) {
	initCertTestDB(t)

	certPEM, keyPEM, err := generateSelfSignedCert([]string{"panel.example.com"}, time.Now().Add(48*time.Hour))
	if err != nil {
		t.Fatalf("generate cert failed: %v", err)
	}

	service := CertService{
		panelCertDir: t.TempDir(),
	}
	if err := service.UploadCertificate(string(certPEM), string(keyPEM)); err != nil {
		t.Fatalf("upload certificate failed: %v", err)
	}

	certFile, keyFile := service.managedPanelPaths()
	if !fileExists(certFile) || !fileExists(keyFile) {
		t.Fatalf("expected uploaded files to exist")
	}

	certInfo, err := os.Stat(certFile)
	if err != nil {
		t.Fatalf("stat cert file failed: %v", err)
	}
	keyInfo, err := os.Stat(keyFile)
	if err != nil {
		t.Fatalf("stat key file failed: %v", err)
	}
	if certInfo.Mode().Perm() != 0644 {
		t.Fatalf("unexpected cert file perm: %o", certInfo.Mode().Perm())
	}
	if keyInfo.Mode().Perm() != 0600 {
		t.Fatalf("unexpected key file perm: %o", keyInfo.Mode().Perm())
	}

	settingService := &SettingService{}
	webCertFile, err := settingService.GetCertFile()
	if err != nil {
		t.Fatalf("get cert file failed: %v", err)
	}
	if webCertFile != certFile {
		t.Fatalf("expected webCertFile=%s, got %s", certFile, webCertFile)
	}
	status, err := settingService.GetWebCertStatus()
	if err != nil {
		t.Fatalf("get cert status failed: %v", err)
	}
	if status != "uploaded" {
		t.Fatalf("expected uploaded status, got %s", status)
	}
}

func TestEnableHTTPSRollsBackSettingsOnRestartFailure(t *testing.T) {
	initCertTestDB(t)

	certPEM, keyPEM, err := generateSelfSignedCert([]string{"panel.example.com"}, time.Now().Add(48*time.Hour))
	if err != nil {
		t.Fatalf("generate cert failed: %v", err)
	}

	service := CertService{
		panelCertDir: t.TempDir(),
		restartPanel: func(delay time.Duration) error {
			return errors.New("restart failed")
		},
		panelPort: func() (int, error) {
			return 54321, nil
		},
	}
	if err := service.UploadCertificate(string(certPEM), string(keyPEM)); err != nil {
		t.Fatalf("upload certificate failed: %v", err)
	}

	settingService := &SettingService{}
	if err := settingService.SetCertFile(""); err != nil {
		t.Fatalf("clear cert file failed: %v", err)
	}
	if err := settingService.SetKeyFile(""); err != nil {
		t.Fatalf("clear key file failed: %v", err)
	}
	if err := settingService.SetWebCertStatus("uploaded"); err != nil {
		t.Fatalf("set status failed: %v", err)
	}

	if _, err := service.EnableHTTPS(); err == nil {
		t.Fatalf("expected enable https to fail on restart error")
	}

	certFile, err := settingService.GetCertFile()
	if err != nil {
		t.Fatalf("get cert file failed: %v", err)
	}
	keyFile, err := settingService.GetKeyFile()
	if err != nil {
		t.Fatalf("get key file failed: %v", err)
	}
	if certFile != "" || keyFile != "" {
		t.Fatalf("expected settings rollback to restore empty cert paths")
	}
	status, err := settingService.GetWebCertStatus()
	if err != nil {
		t.Fatalf("get cert status failed: %v", err)
	}
	if status != "uploaded" {
		t.Fatalf("expected uploaded status after rollback, got %s", status)
	}
}

func TestDisableHTTPSRollsBackSettingsOnRestartFailure(t *testing.T) {
	initCertTestDB(t)

	certPEM, keyPEM, err := generateSelfSignedCert([]string{"panel.example.com"}, time.Now().Add(48*time.Hour))
	if err != nil {
		t.Fatalf("generate cert failed: %v", err)
	}

	service := CertService{
		panelCertDir: t.TempDir(),
		restartPanel: func(delay time.Duration) error {
			return errors.New("restart failed")
		},
		panelPort: func() (int, error) {
			return 54321, nil
		},
	}
	if err := service.UploadCertificate(string(certPEM), string(keyPEM)); err != nil {
		t.Fatalf("upload certificate failed: %v", err)
	}

	settingService := &SettingService{}
	if err := settingService.SetWebCertStatus("enabled"); err != nil {
		t.Fatalf("set status failed: %v", err)
	}

	if _, err := service.DisableHTTPS(); err == nil {
		t.Fatalf("expected disable https to fail on restart error")
	}

	certFile, keyFile := service.managedPanelPaths()
	currentCertFile, err := settingService.GetCertFile()
	if err != nil {
		t.Fatalf("get cert file failed: %v", err)
	}
	currentKeyFile, err := settingService.GetKeyFile()
	if err != nil {
		t.Fatalf("get key file failed: %v", err)
	}
	if currentCertFile != certFile || currentKeyFile != keyFile {
		t.Fatalf("expected settings rollback to restore managed cert paths")
	}
	status, err := settingService.GetWebCertStatus()
	if err != nil {
		t.Fatalf("get cert status failed: %v", err)
	}
	if status != "enabled" {
		t.Fatalf("expected enabled status after rollback, got %s", status)
	}
}

func TestEnableHTTPSRollsBackSettingsWhenHTTPSListenerNotReady(t *testing.T) {
	initCertTestDB(t)

	certPEM, keyPEM, err := generateSelfSignedCert([]string{"panel.example.com"}, time.Now().Add(48*time.Hour))
	if err != nil {
		t.Fatalf("generate cert failed: %v", err)
	}

	restartCount := 0
	service := CertService{
		panelCertDir: t.TempDir(),
		restartPanel: func(delay time.Duration) error {
			restartCount++
			return nil
		},
		panelPort: func() (int, error) {
			return 54321, nil
		},
		waitHTTPSReadyFn: func(port int, timeout time.Duration) error {
			return errors.New("tls probe failed")
		},
		waitHTTPReadyFn: func(port int, timeout time.Duration) error {
			return nil
		},
	}
	if err := service.UploadCertificate(string(certPEM), string(keyPEM)); err != nil {
		t.Fatalf("upload certificate failed: %v", err)
	}

	settingService := &SettingService{}
	if err := settingService.SetCertFile(""); err != nil {
		t.Fatalf("clear cert file failed: %v", err)
	}
	if err := settingService.SetKeyFile(""); err != nil {
		t.Fatalf("clear key file failed: %v", err)
	}
	if err := settingService.SetWebCertStatus("uploaded"); err != nil {
		t.Fatalf("set status failed: %v", err)
	}

	if _, err := service.EnableHTTPS(); err == nil {
		t.Fatalf("expected enable https to fail on listener wait error")
	}
	if restartCount != 2 {
		t.Fatalf("expected two restart attempts, got %d", restartCount)
	}

	certFile, err := settingService.GetCertFile()
	if err != nil {
		t.Fatalf("get cert file failed: %v", err)
	}
	keyFile, err := settingService.GetKeyFile()
	if err != nil {
		t.Fatalf("get key file failed: %v", err)
	}
	if certFile != "" || keyFile != "" {
		t.Fatalf("expected settings rollback to restore empty cert paths")
	}
	status, err := settingService.GetWebCertStatus()
	if err != nil {
		t.Fatalf("get cert status failed: %v", err)
	}
	if status != "uploaded" {
		t.Fatalf("expected uploaded status after rollback, got %s", status)
	}
}

func TestDisableHTTPSRollsBackSettingsWhenHTTPListenerNotReady(t *testing.T) {
	initCertTestDB(t)

	certPEM, keyPEM, err := generateSelfSignedCert([]string{"panel.example.com"}, time.Now().Add(48*time.Hour))
	if err != nil {
		t.Fatalf("generate cert failed: %v", err)
	}

	restartCount := 0
	service := CertService{
		panelCertDir: t.TempDir(),
		restartPanel: func(delay time.Duration) error {
			restartCount++
			return nil
		},
		panelPort: func() (int, error) {
			return 54321, nil
		},
		waitHTTPSReadyFn: func(port int, timeout time.Duration) error {
			return nil
		},
		waitHTTPReadyFn: func(port int, timeout time.Duration) error {
			return errors.New("http probe failed")
		},
	}
	if err := service.UploadCertificate(string(certPEM), string(keyPEM)); err != nil {
		t.Fatalf("upload certificate failed: %v", err)
	}

	settingService := &SettingService{}
	if err := settingService.SetWebCertStatus("enabled"); err != nil {
		t.Fatalf("set status failed: %v", err)
	}

	if _, err := service.DisableHTTPS(); err == nil {
		t.Fatalf("expected disable https to fail on listener wait error")
	}
	if restartCount != 2 {
		t.Fatalf("expected two restart attempts, got %d", restartCount)
	}

	certFile, keyFile := service.managedPanelPaths()
	currentCertFile, err := settingService.GetCertFile()
	if err != nil {
		t.Fatalf("get cert file failed: %v", err)
	}
	currentKeyFile, err := settingService.GetKeyFile()
	if err != nil {
		t.Fatalf("get key file failed: %v", err)
	}
	if currentCertFile != certFile || currentKeyFile != keyFile {
		t.Fatalf("expected settings rollback to restore managed cert paths")
	}
	status, err := settingService.GetWebCertStatus()
	if err != nil {
		t.Fatalf("get cert status failed: %v", err)
	}
	if status != "enabled" {
		t.Fatalf("expected enabled status after rollback, got %s", status)
	}
}

func TestWaitForHTTPReady(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	port := mustTestURLPort(t, server.URL)
	service := CertService{}
	if err := service.waitForHTTPReady(port, 2*time.Second); err != nil {
		t.Fatalf("wait for http ready failed: %v", err)
	}
}

func TestWaitForHTTPReadyRejectsHTTPSRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://127.0.0.1"+r.URL.Path, http.StatusTemporaryRedirect)
	}))
	defer server.Close()

	port := mustTestURLPort(t, server.URL)
	service := CertService{}
	if err := service.waitForHTTPReady(port, time.Second); err == nil {
		t.Fatalf("expected https redirect response to fail http readiness")
	}
}

func TestWaitForHTTPSReady(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	port := mustTestURLPort(t, server.URL)
	service := CertService{}
	if err := service.waitForHTTPSReady(port, 2*time.Second); err != nil {
		t.Fatalf("wait for https ready failed: %v", err)
	}
}

func TestWaitForHTTPReadyRejectsTLSServer(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	port := mustTestURLPort(t, server.URL)
	service := CertService{}
	if err := service.waitForHTTPReady(port, time.Second); err == nil {
		t.Fatalf("expected tls server to fail http readiness")
	}
}

func TestWaitForReadyTimeout(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve local port failed: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatalf("close reserved listener failed: %v", err)
	}

	service := CertService{}
	if err := service.waitForHTTPReady(port, time.Second); err == nil {
		t.Fatalf("expected http readiness probe to time out")
	}
	if err := service.waitForHTTPSReady(port, time.Second); err == nil {
		t.Fatalf("expected https readiness probe to time out")
	}
}

func TestFindAcmeShPrefersRootHome(t *testing.T) {
	rootHome := filepath.Join(t.TempDir(), "root-acme")
	if err := os.MkdirAll(rootHome, 0755); err != nil {
		t.Fatalf("mkdir acme home failed: %v", err)
	}
	rootBin := filepath.Join(rootHome, "acme.sh")
	if err := os.WriteFile(rootBin, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatalf("write acme bin failed: %v", err)
	}

	service := CertService{
		acmeSearchHomes: []string{rootHome},
	}
	acmeHome, acmeBin, err := service.findAcmeSh()
	if err != nil {
		t.Fatalf("findAcmeSh failed: %v", err)
	}
	if acmeHome != rootHome {
		t.Fatalf("expected acme home %s, got %s", rootHome, acmeHome)
	}
	if acmeBin != rootBin {
		t.Fatalf("expected acme bin %s, got %s", rootBin, acmeBin)
	}
}

func TestFindAcmeShFallsBackToLegacyRootPath(t *testing.T) {
	legacyHome := filepath.Join(t.TempDir(), "legacy-acme")
	if err := os.MkdirAll(legacyHome, 0755); err != nil {
		t.Fatalf("mkdir legacy acme home failed: %v", err)
	}
	legacyBin := filepath.Join(legacyHome, "acme.sh")
	if err := os.WriteFile(legacyBin, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatalf("write legacy acme bin failed: %v", err)
	}

	service := CertService{
		acmeSearchHomes: []string{filepath.Join(t.TempDir(), "missing"), legacyHome},
	}
	acmeHome, acmeBin, err := service.findAcmeSh()
	if err != nil {
		t.Fatalf("findAcmeSh failed: %v", err)
	}
	if acmeHome != legacyHome {
		t.Fatalf("expected legacy acme home %s, got %s", legacyHome, acmeHome)
	}
	if acmeBin != legacyBin {
		t.Fatalf("expected legacy acme bin %s, got %s", legacyBin, acmeBin)
	}
}

func TestFindAcmeShReturnsErrorWhenMissing(t *testing.T) {
	service := CertService{
		acmeSearchHomes: []string{filepath.Join(t.TempDir(), "missing")},
		lookPath: func(string) (string, error) {
			return "", errors.New("not found")
		},
	}
	if _, _, err := service.findAcmeSh(); err == nil {
		t.Fatalf("expected findAcmeSh to fail when acme.sh is missing")
	}
}

func TestInstallAcmeCertificateUsesDiscoveredHomeAndBin(t *testing.T) {
	acmeHome := filepath.Join(t.TempDir(), "acme-home")
	if err := os.MkdirAll(acmeHome, 0755); err != nil {
		t.Fatalf("mkdir acme home failed: %v", err)
	}
	acmeBin := filepath.Join(acmeHome, "acme.sh")
	if err := os.WriteFile(acmeBin, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatalf("write acme bin failed: %v", err)
	}

	var gotName string
	var gotArgs []string
	service := CertService{
		acmeSearchHomes: []string{acmeHome},
		runCommand: func(name string, args ...string) ([]byte, error) {
			gotName = name
			gotArgs = append([]string{}, args...)
			return []byte("ok"), nil
		},
	}
	if err := service.installAcmeCertificate("example.com", "/tmp/panel.crt", "/tmp/panel.key"); err != nil {
		t.Fatalf("installAcmeCertificate failed: %v", err)
	}
	if gotName != acmeBin {
		t.Fatalf("expected command %s, got %s", acmeBin, gotName)
	}
	expectedPrefix := []string{"--home", acmeHome, "--install-cert", "-d", "example.com"}
	for i, want := range expectedPrefix {
		if i >= len(gotArgs) || gotArgs[i] != want {
			t.Fatalf("expected arg[%d]=%s, got %v", i, want, gotArgs)
		}
	}
}

func generateSelfSignedCert(dnsNames []string, notAfter time.Time) ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return nil, nil, err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: dnsNames[0],
		},
		DNSNames:              dnsNames,
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	return certPEM, keyPEM, nil
}

func mustTestURLPort(t *testing.T, rawURL string) int {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse test url failed: %v", err)
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		t.Fatalf("parse test port failed: %v", err)
	}
	return port
}
