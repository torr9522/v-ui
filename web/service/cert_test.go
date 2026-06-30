package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
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
