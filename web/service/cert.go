package service

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"x-ui/util/common"
	"x-ui/web/entity"
)

const (
	defaultPanelCertDir  = "/usr/local/x-ui/cert"
	defaultPanelCertFile = defaultPanelCertDir + "/panel.crt"
	defaultPanelKeyFile  = defaultPanelCertDir + "/panel.key"
	defaultRestartDelay  = 3 * time.Second
)

type CertService struct {
	panelCertDir string
	restartPanel func(time.Duration) error
	publicIPs    func() []string
}

type certInfo struct {
	issuer    string
	subject   string
	dnsNames  []string
	notBefore int64
	notAfter  int64
	expireAt  int64
	daysLeft  int64
	isExpired bool
}

type certSettingsSnapshot struct {
	certFile  string
	keyFile   string
	status    string
	expireAt  int64
	issuer    string
	autoRenew bool
	mode      string
	provider  string
}

func (s *CertService) GetStatus() (*entity.CertStatus, error) {
	settingService := &SettingService{}

	webDomain, err := settingService.GetWebDomain()
	if err != nil {
		return nil, err
	}
	webCertFile, err := settingService.GetCertFile()
	if err != nil {
		return nil, err
	}
	webKeyFile, err := settingService.GetKeyFile()
	if err != nil {
		return nil, err
	}
	webCertStatus, err := settingService.GetWebCertStatus()
	if err != nil {
		return nil, err
	}
	webCertExpireAt, err := settingService.GetWebCertExpireAt()
	if err != nil {
		return nil, err
	}
	webCertIssuer, err := settingService.GetWebCertIssuer()
	if err != nil {
		return nil, err
	}
	webCertAutoRenew, err := settingService.GetWebCertAutoRenew()
	if err != nil {
		return nil, err
	}
	webCertMode, err := settingService.GetWebCertMode()
	if err != nil {
		return nil, err
	}
	webCertProvider, err := settingService.GetWebCertProvider()
	if err != nil {
		return nil, err
	}

	displayCertFile, displayKeyFile := s.resolveStatusPaths(webCertFile, webKeyFile, webCertMode)
	status := &entity.CertStatus{
		WebDomain:        webDomain,
		HTTPSActive:      webCertFile != "" && webKeyFile != "",
		WebCertFile:      displayCertFile,
		WebKeyFile:       displayKeyFile,
		CertExists:       fileExists(displayCertFile),
		KeyExists:        fileExists(displayKeyFile),
		WebCertStatus:    webCertStatus,
		WebCertExpireAt:  webCertExpireAt,
		WebCertIssuer:    webCertIssuer,
		WebCertAutoRenew: webCertAutoRenew,
		WebCertMode:      webCertMode,
		WebCertProvider:  webCertProvider,
		DNSNames:         []string{},
		Warnings:         []string{},
	}

	if !status.HTTPSActive && status.CertExists && status.KeyExists {
		status.Warnings = append(status.Warnings, "managed certificate files exist, but HTTPS is currently disabled")
	}
	if status.CertExists != status.KeyExists {
		status.WebCertStatus = "unknown"
		status.Warnings = append(status.Warnings, "certificate file and key file are not both present")
	}
	if !status.CertExists {
		return status, nil
	}

	info, err := readCertificateInfoFromFiles(displayCertFile, displayKeyFile)
	if err != nil {
		status.WebCertStatus = "unknown"
		status.Warnings = append(status.Warnings, err.Error())
		return status, nil
	}

	status.WebCertIssuer = info.issuer
	status.Subject = info.subject
	status.DNSNames = info.dnsNames
	status.NotBefore = info.notBefore
	status.NotAfter = info.notAfter
	status.WebCertExpireAt = info.expireAt
	status.DaysLeft = info.daysLeft
	status.IsExpired = info.isExpired

	if status.WebCertStatus == "" || status.WebCertStatus == "none" {
		status.WebCertStatus = "ready"
	}
	if info.isExpired {
		status.WebCertStatus = "expired"
	}

	return status, nil
}

func (s *CertService) SetDomain(domain string) error {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return (&SettingService{}).SetWebDomain("")
	}
	if err := validateDomain(domain); err != nil {
		return err
	}
	return (&SettingService{}).SetWebDomain(domain)
}

func (s *CertService) CheckDomain(domain string) (*entity.DomainCheckResult, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return nil, errors.New("domain is required")
	}
	if err := validateDomain(domain); err != nil {
		return nil, err
	}

	records := make([]string, 0)
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}
	recordSet := map[string]struct{}{}
	for _, ip := range ips {
		recordSet[ip.String()] = struct{}{}
	}
	for ip := range recordSet {
		records = append(records, ip)
	}
	sort.Strings(records)

	serverIPs := s.lookupPublicServerIPs()
	serverSet := map[string]struct{}{}
	for _, ip := range serverIPs {
		serverSet[ip] = struct{}{}
	}

	matched := false
	for _, record := range records {
		if _, ok := serverSet[record]; ok {
			matched = true
			break
		}
	}

	result := &entity.DomainCheckResult{
		Domain:    domain,
		Records:   records,
		ServerIPs: serverIPs,
		Matched:   matched,
		Warnings:  []string{},
	}
	if len(serverIPs) == 0 {
		result.Warnings = append(result.Warnings, "failed to detect public server IP")
	}
	if len(records) == 0 {
		result.Warnings = append(result.Warnings, "no A/AAAA record found")
	}

	return result, nil
}

func (s *CertService) UploadCertificate(certPEM string, keyPEM string) error {
	certPEM = strings.TrimSpace(certPEM)
	keyPEM = strings.TrimSpace(keyPEM)
	if certPEM == "" || keyPEM == "" {
		return errors.New("certPem and keyPem are required")
	}

	info, err := parseCertificatePair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return err
	}

	certFile, keyFile := s.managedPanelPaths()
	if err := os.MkdirAll(filepath.Dir(certFile), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(certFile, []byte(certPEM+"\n"), 0644); err != nil {
		return err
	}
	if err := os.Chmod(certFile, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(keyFile, []byte(keyPEM+"\n"), 0600); err != nil {
		return err
	}
	if err := os.Chmod(keyFile, 0600); err != nil {
		return err
	}

	settingService := &SettingService{}
	if err := settingService.SetCertFile(certFile); err != nil {
		return err
	}
	if err := settingService.SetKeyFile(keyFile); err != nil {
		return err
	}
	if err := settingService.SetWebCertStatus("uploaded"); err != nil {
		return err
	}
	if err := settingService.SetWebCertMode("manual"); err != nil {
		return err
	}
	if err := settingService.SetWebCertProvider("manual"); err != nil {
		return err
	}
	if err := settingService.SetWebCertIssuer(info.issuer); err != nil {
		return err
	}
	if err := settingService.SetWebCertExpireAt(info.expireAt); err != nil {
		return err
	}
	if err := settingService.SetWebCertAutoRenew(false); err != nil {
		return err
	}

	return nil
}

func (s *CertService) EnableHTTPS() (bool, error) {
	settingService := &SettingService{}
	snapshot, err := s.captureSettings(settingService)
	if err != nil {
		return false, err
	}

	certFile, keyFile, err := s.resolveEnablePaths(settingService)
	if err != nil {
		return false, err
	}
	info, err := readCertificateInfoFromFiles(certFile, keyFile)
	if err != nil {
		return false, err
	}

	if err := settingService.SetCertFile(certFile); err != nil {
		return false, err
	}
	if err := settingService.SetKeyFile(keyFile); err != nil {
		return false, err
	}
	if err := settingService.SetWebCertStatus("enabled"); err != nil {
		return false, err
	}
	if err := settingService.SetWebCertMode("manual"); err != nil {
		return false, err
	}
	if err := settingService.SetWebCertProvider("manual"); err != nil {
		return false, err
	}
	if err := settingService.SetWebCertIssuer(info.issuer); err != nil {
		return false, err
	}
	if err := settingService.SetWebCertExpireAt(info.expireAt); err != nil {
		return false, err
	}

	if err := s.restart(defaultRestartDelay); err != nil {
		restoreErr := s.restoreSettings(settingService, snapshot)
		restartRestoreErr := s.restart(defaultRestartDelay)
		return false, common.Combine(err, restoreErr, restartRestoreErr)
	}

	return true, nil
}

func (s *CertService) DisableHTTPS() (bool, error) {
	settingService := &SettingService{}
	snapshot, err := s.captureSettings(settingService)
	if err != nil {
		return false, err
	}

	if err := settingService.SetCertFile(""); err != nil {
		return false, err
	}
	if err := settingService.SetKeyFile(""); err != nil {
		return false, err
	}

	managedCertFile, managedKeyFile := s.managedPanelPaths()
	if fileExists(managedCertFile) && fileExists(managedKeyFile) {
		if err := settingService.SetWebCertStatus("uploaded"); err != nil {
			return false, err
		}
		if err := settingService.SetWebCertMode("manual"); err != nil {
			return false, err
		}
		if err := settingService.SetWebCertProvider("manual"); err != nil {
			return false, err
		}
	} else {
		if err := settingService.SetWebCertStatus("none"); err != nil {
			return false, err
		}
		if err := settingService.SetWebCertMode("none"); err != nil {
			return false, err
		}
		if err := settingService.SetWebCertProvider(""); err != nil {
			return false, err
		}
	}

	if err := s.restart(defaultRestartDelay); err != nil {
		restoreErr := s.restoreSettings(settingService, snapshot)
		restartRestoreErr := s.restart(defaultRestartDelay)
		return false, common.Combine(err, restoreErr, restartRestoreErr)
	}

	return true, nil
}

func (s *CertService) managedPanelPaths() (string, string) {
	dir := strings.TrimSpace(s.panelCertDir)
	if dir == "" {
		dir = defaultPanelCertDir
	}
	return filepath.Join(dir, "panel.crt"), filepath.Join(dir, "panel.key")
}

func (s *CertService) resolveStatusPaths(webCertFile string, webKeyFile string, webCertMode string) (string, string) {
	if webCertFile != "" || webKeyFile != "" {
		return webCertFile, webKeyFile
	}
	if webCertMode == "manual" {
		return s.managedPanelPaths()
	}
	return webCertFile, webKeyFile
}

func (s *CertService) resolveEnablePaths(settingService *SettingService) (string, string, error) {
	certFile, err := settingService.GetCertFile()
	if err != nil {
		return "", "", err
	}
	keyFile, err := settingService.GetKeyFile()
	if err != nil {
		return "", "", err
	}
	if certFile != "" && keyFile != "" {
		if !fileExists(certFile) || !fileExists(keyFile) {
			return "", "", errors.New("configured certificate file or key file does not exist")
		}
		return certFile, keyFile, nil
	}

	mode, err := settingService.GetWebCertMode()
	if err != nil {
		return "", "", err
	}
	if mode != "manual" {
		return "", "", errors.New("webCertFile and webKeyFile are not configured")
	}
	certFile, keyFile = s.managedPanelPaths()
	if !fileExists(certFile) || !fileExists(keyFile) {
		return "", "", errors.New("managed manual certificate files do not exist")
	}
	return certFile, keyFile, nil
}

func (s *CertService) captureSettings(settingService *SettingService) (*certSettingsSnapshot, error) {
	certFile, err := settingService.GetCertFile()
	if err != nil {
		return nil, err
	}
	keyFile, err := settingService.GetKeyFile()
	if err != nil {
		return nil, err
	}
	status, err := settingService.GetWebCertStatus()
	if err != nil {
		return nil, err
	}
	expireAt, err := settingService.GetWebCertExpireAt()
	if err != nil {
		return nil, err
	}
	issuer, err := settingService.GetWebCertIssuer()
	if err != nil {
		return nil, err
	}
	autoRenew, err := settingService.GetWebCertAutoRenew()
	if err != nil {
		return nil, err
	}
	mode, err := settingService.GetWebCertMode()
	if err != nil {
		return nil, err
	}
	provider, err := settingService.GetWebCertProvider()
	if err != nil {
		return nil, err
	}

	return &certSettingsSnapshot{
		certFile:  certFile,
		keyFile:   keyFile,
		status:    status,
		expireAt:  expireAt,
		issuer:    issuer,
		autoRenew: autoRenew,
		mode:      mode,
		provider:  provider,
	}, nil
}

func (s *CertService) restoreSettings(settingService *SettingService, snapshot *certSettingsSnapshot) error {
	if snapshot == nil {
		return nil
	}
	return common.Combine(
		settingService.SetCertFile(snapshot.certFile),
		settingService.SetKeyFile(snapshot.keyFile),
		settingService.SetWebCertStatus(snapshot.status),
		settingService.SetWebCertExpireAt(snapshot.expireAt),
		settingService.SetWebCertIssuer(snapshot.issuer),
		settingService.SetWebCertAutoRenew(snapshot.autoRenew),
		settingService.SetWebCertMode(snapshot.mode),
		settingService.SetWebCertProvider(snapshot.provider),
	)
}

func (s *CertService) restart(delay time.Duration) error {
	if s.restartPanel != nil {
		return s.restartPanel(delay)
	}
	return (&PanelService{}).RestartPanel(delay)
}

func (s *CertService) lookupPublicServerIPs() []string {
	if s.publicIPs != nil {
		return s.publicIPs()
	}
	return getPublicServerIPs()
}

func validateDomain(domain string) error {
	if strings.Contains(domain, "://") {
		return errors.New("domain must not include scheme")
	}
	if strings.Contains(domain, "/") || strings.Contains(domain, "?") || strings.Contains(domain, "#") {
		return errors.New("domain must not include path or query")
	}
	if strings.ContainsAny(domain, " \t\r\n") {
		return errors.New("domain must not include spaces")
	}
	if strings.HasSuffix(domain, ".") {
		domain = strings.TrimSuffix(domain, ".")
	}
	if domain == "" {
		return errors.New("domain is required")
	}

	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return errors.New("domain format is invalid")
	}
	for _, label := range labels {
		if label == "" || len(label) > 63 {
			return errors.New("domain format is invalid")
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return errors.New("domain format is invalid")
		}
		for _, r := range label {
			switch {
			case r >= 'a' && r <= 'z':
			case r >= 'A' && r <= 'Z':
			case r >= '0' && r <= '9':
			case r == '-':
			default:
				return errors.New("domain format is invalid")
			}
		}
	}

	return nil
}

func readCertificateInfoFromFiles(certFile string, keyFile string) (*certInfo, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, errors.New("failed to read certificate file")
	}
	var keyPEM []byte
	if keyFile != "" && fileExists(keyFile) {
		keyPEM, err = os.ReadFile(keyFile)
		if err != nil {
			return nil, errors.New("failed to read key file")
		}
	}
	return readCertificateInfo(certPEM, keyPEM)
}

func readCertificateInfo(certPEM []byte, keyPEM []byte) (*certInfo, error) {
	info, err := parseCertificatePEM(certPEM)
	if err != nil {
		return nil, err
	}
	if len(keyPEM) > 0 {
		if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
			return nil, errors.New("certificate and private key do not match")
		}
	}
	return info, nil
}

func parseCertificatePair(certPEM []byte, keyPEM []byte) (*certInfo, error) {
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		return nil, errors.New("certificate and private key are required")
	}
	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		return nil, errors.New("certificate and private key do not match")
	}
	return parseCertificatePEM(certPEM)
}

func parseCertificatePEM(certPEM []byte) (*certInfo, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.New("failed to parse certificate")
	}

	issuer := cert.Issuer.CommonName
	if issuer == "" {
		issuer = cert.Issuer.String()
	}
	subject := cert.Subject.CommonName
	if subject == "" {
		subject = cert.Subject.String()
	}
	dnsNames := append([]string{}, cert.DNSNames...)
	if cert.Subject.CommonName != "" && !containsString(dnsNames, cert.Subject.CommonName) {
		dnsNames = append([]string{cert.Subject.CommonName}, dnsNames...)
	}

	now := time.Now()
	daysLeft := int64(0)
	isExpired := cert.NotAfter.Before(now)
	if !isExpired {
		daysLeft = int64(cert.NotAfter.Sub(now).Hours() / 24)
	}

	return &certInfo{
		issuer:    issuer,
		subject:   subject,
		dnsNames:  dnsNames,
		notBefore: cert.NotBefore.Unix(),
		notAfter:  cert.NotAfter.Unix(),
		expireAt:  cert.NotAfter.Unix(),
		daysLeft:  daysLeft,
		isExpired: isExpired,
	}, nil
}

func getPublicServerIPs() []string {
	endpoints := []string{
		"https://ifconfig.me/ip",
		"https://api.ip.sb/ip",
	}
	set := map[string]struct{}{}
	client := &http.Client{Timeout: 5 * time.Second}

	for _, endpoint := range endpoints {
		resp, err := client.Get(endpoint)
		if err != nil {
			continue
		}
		func() {
			defer resp.Body.Close()
			body, err := io.ReadAll(io.LimitReader(resp.Body, 128))
			if err != nil {
				return
			}
			value := strings.TrimSpace(string(body))
			if value == "" {
				return
			}
			if ip := net.ParseIP(value); ip != nil {
				set[ip.String()] = struct{}{}
			}
		}()
	}

	if localIPs, err := net.LookupIP(strings.TrimSpace(hostname())); err == nil {
		for _, ip := range localIPs {
			if ip.IsGlobalUnicast() {
				set[ip.String()] = struct{}{}
			}
		}
	}

	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	return name
}

func fileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
