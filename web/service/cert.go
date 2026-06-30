package service

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"x-ui/util/common"
	"x-ui/web/entity"
)

const (
	defaultPanelCertDir        = "/usr/local/x-ui/cert"
	defaultPanelCertFile       = defaultPanelCertDir + "/panel.crt"
	defaultPanelKeyFile        = defaultPanelCertDir + "/panel.key"
	defaultAcmePanelCertFile   = defaultPanelCertDir + "/acme-panel.crt"
	defaultAcmePanelKeyFile    = defaultPanelCertDir + "/acme-panel.key"
	defaultRestartDelay        = 3 * time.Second
	defaultListenerWaitTimeout = 15 * time.Second
	defaultListenerRetryDelay  = 500 * time.Millisecond
	defaultProbeTimeout        = 2 * time.Second
)

type CertService struct {
	panelCertDir     string
	restartPanel     func(time.Duration) error
	publicIPs        func() []string
	panelPort        func() (int, error)
	waitHTTPSReadyFn func(int, time.Duration) error
	waitHTTPReadyFn  func(int, time.Duration) error
	runCommand       func(name string, args ...string) ([]byte, error)
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
	domain    string
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
	certFile, keyFile, err := s.resolveEnablePaths(settingService)
	if err != nil {
		return false, err
	}
	mode, err := settingService.GetWebCertMode()
	if err != nil {
		return false, err
	}
	provider, err := settingService.GetWebCertProvider()
	if err != nil {
		return false, err
	}
	if mode == "" || mode == "none" {
		mode = "manual"
	}
	if provider == "" {
		provider = "manual"
	}
	return s.applyHTTPSCertificate(certFile, keyFile, mode, provider, "enabled", false)
}

func (s *CertService) DisableHTTPS() (bool, error) {
	settingService := &SettingService{}
	snapshot, err := s.captureSettings(settingService)
	if err != nil {
		return false, err
	}
	port, err := s.getPanelPort(settingService)
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
	acmeCertFile, acmeKeyFile := s.acmePanelPaths()
	if snapshot.mode == "acme_http" && fileExists(acmeCertFile) && fileExists(acmeKeyFile) {
		if err := settingService.SetWebCertStatus("issued"); err != nil {
			return false, err
		}
		if err := settingService.SetWebCertMode("acme_http"); err != nil {
			return false, err
		}
		if err := settingService.SetWebCertProvider(snapshot.provider); err != nil {
			return false, err
		}
	} else if fileExists(managedCertFile) && fileExists(managedKeyFile) {
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
	if err := s.waitForHTTPReady(port, defaultListenerWaitTimeout); err != nil {
		return false, s.rollbackListenerState(
			settingService,
			snapshot,
			port,
			errors.New("HTTP listener not ready after restart"),
			err,
		)
	}

	return true, nil
}

func (s *CertService) IssueHTTP(domain string, email string, staging bool) (*entity.AcmeIssueResult, error) {
	settingService := &SettingService{}

	resolvedDomain, err := s.resolveIssueDomain(settingService, domain)
	if err != nil {
		return nil, err
	}
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	checkResult, err := s.CheckDomain(resolvedDomain)
	if err != nil {
		return nil, err
	}
	if !checkResult.Matched {
		return nil, errors.New("domain does not resolve to current server IP")
	}
	if err := s.ensurePortAvailable(80); err != nil {
		return nil, err
	}
	if err := s.ensureAcmeInstalled(email); err != nil {
		return nil, err
	}
	if err := s.ensureSocatInstalled(); err != nil {
		return nil, err
	}

	certFile, keyFile := s.acmePanelPaths()
	if err := os.MkdirAll(filepath.Dir(certFile), 0755); err != nil {
		return nil, err
	}
	if err := s.issueAcmeCertificate(resolvedDomain, staging); err != nil {
		return nil, err
	}
	if err := s.installAcmeCertificate(resolvedDomain, certFile, keyFile); err != nil {
		return nil, err
	}
	if err := os.Chmod(certFile, 0644); err != nil {
		return nil, err
	}
	if err := os.Chmod(keyFile, 0600); err != nil {
		return nil, err
	}

	if err := settingService.SetWebDomain(resolvedDomain); err != nil {
		return nil, err
	}
	applied, err := s.applyHTTPSCertificate(certFile, keyFile, "acme_http", "letsencrypt", "enabled", false)
	if err != nil {
		return nil, err
	}
	status, err := s.GetStatus()
	if err != nil {
		return nil, err
	}
	return &entity.AcmeIssueResult{
		Domain:  resolvedDomain,
		Staging: staging,
		Applied: applied,
		Status:  status,
	}, nil
}

func (s *CertService) managedPanelPaths() (string, string) {
	dir := strings.TrimSpace(s.panelCertDir)
	if dir == "" {
		dir = defaultPanelCertDir
	}
	return filepath.Join(dir, "panel.crt"), filepath.Join(dir, "panel.key")
}

func (s *CertService) acmePanelPaths() (string, string) {
	dir := strings.TrimSpace(s.panelCertDir)
	if dir == "" {
		dir = defaultPanelCertDir
	}
	return filepath.Join(dir, "acme-panel.crt"), filepath.Join(dir, "acme-panel.key")
}

func (s *CertService) resolveStatusPaths(webCertFile string, webKeyFile string, webCertMode string) (string, string) {
	if webCertFile != "" || webKeyFile != "" {
		return webCertFile, webKeyFile
	}
	if webCertMode == "manual" {
		return s.managedPanelPaths()
	}
	if webCertMode == "acme_http" {
		return s.acmePanelPaths()
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
	if mode != "manual" && mode != "acme_http" {
		return "", "", errors.New("webCertFile and webKeyFile are not configured")
	}
	if mode == "acme_http" {
		certFile, keyFile = s.acmePanelPaths()
	} else {
		certFile, keyFile = s.managedPanelPaths()
	}
	if !fileExists(certFile) || !fileExists(keyFile) {
		if mode == "acme_http" {
			return "", "", errors.New("managed acme certificate files do not exist")
		}
		return "", "", errors.New("managed manual certificate files do not exist")
	}
	return certFile, keyFile, nil
}

func (s *CertService) captureSettings(settingService *SettingService) (*certSettingsSnapshot, error) {
	domain, err := settingService.GetWebDomain()
	if err != nil {
		return nil, err
	}
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
		domain:    domain,
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
		settingService.SetWebDomain(snapshot.domain),
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

func (s *CertService) applyHTTPSCertificate(certFile string, keyFile string, mode string, provider string, status string, autoRenew bool) (bool, error) {
	settingService := &SettingService{}
	snapshot, err := s.captureSettings(settingService)
	if err != nil {
		return false, err
	}
	port, err := s.getPanelPort(settingService)
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
	if err := settingService.SetWebCertStatus(status); err != nil {
		return false, err
	}
	if err := settingService.SetWebCertMode(mode); err != nil {
		return false, err
	}
	if err := settingService.SetWebCertProvider(provider); err != nil {
		return false, err
	}
	if err := settingService.SetWebCertIssuer(info.issuer); err != nil {
		return false, err
	}
	if err := settingService.SetWebCertExpireAt(info.expireAt); err != nil {
		return false, err
	}
	if err := settingService.SetWebCertAutoRenew(autoRenew); err != nil {
		return false, err
	}

	if err := s.restart(defaultRestartDelay); err != nil {
		restoreErr := s.restoreSettings(settingService, snapshot)
		restartRestoreErr := s.restart(defaultRestartDelay)
		return false, common.Combine(err, restoreErr, restartRestoreErr)
	}
	if err := s.waitForHTTPSReady(port, defaultListenerWaitTimeout); err != nil {
		return false, s.rollbackListenerState(
			settingService,
			snapshot,
			port,
			errors.New("HTTPS listener not ready after restart"),
			err,
		)
	}

	return true, nil
}

func (s *CertService) restart(delay time.Duration) error {
	if s.restartPanel != nil {
		return s.restartPanel(delay)
	}
	return (&PanelService{}).RestartPanel(delay)
}

func (s *CertService) resolveIssueDomain(settingService *SettingService, domain string) (string, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		var err error
		domain, err = settingService.GetWebDomain()
		if err != nil {
			return "", err
		}
		domain = strings.TrimSpace(domain)
	}
	if domain == "" {
		return "", errors.New("domain is required")
	}
	if err := validateDomain(domain); err != nil {
		return "", err
	}
	return domain, nil
}

func (s *CertService) getPanelPort(settingService *SettingService) (int, error) {
	if s.panelPort != nil {
		return s.panelPort()
	}
	port, err := settingService.GetPort()
	if err != nil {
		return 0, err
	}
	if port <= 0 {
		return 0, errors.New("invalid panel port")
	}
	return port, nil
}

func (s *CertService) waitForHTTPSReady(port int, timeout time.Duration) error {
	if s.waitHTTPSReadyFn != nil {
		return s.waitHTTPSReadyFn(port, timeout)
	}
	addr, err := loopbackAddr(port)
	if err != nil {
		return err
	}
	dialer := &net.Dialer{Timeout: defaultProbeTimeout}
	return waitForListener(timeout, func() error {
		conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return err
		}
		return conn.Close()
	})
}

func (s *CertService) waitForHTTPReady(port int, timeout time.Duration) error {
	if s.waitHTTPReadyFn != nil {
		return s.waitHTTPReadyFn(port, timeout)
	}
	addr, err := loopbackAddr(port)
	if err != nil {
		return err
	}
	client := &http.Client{
		Timeout: defaultProbeTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	url := "http://" + addr + "/"
	return waitForListener(timeout, func() error {
		// If a TLS handshake still succeeds, the panel has not switched back to HTTP-only.
		if err := s.waitForHTTPSReady(port, defaultProbeTimeout); err == nil {
			return errors.New("https listener still active")
		}

		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if isHTTPSRedirectResponse(resp) {
			return errors.New("http endpoint still redirects to https")
		}
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1))
		return nil
	})
}

func (s *CertService) rollbackListenerState(settingService *SettingService, snapshot *certSettingsSnapshot, port int, reason error, cause error) error {
	restoreErr := s.restoreSettings(settingService, snapshot)
	restartRestoreErr := s.restart(defaultRestartDelay)
	var waitRestoreErr error
	if restartRestoreErr == nil {
		waitRestoreErr = s.waitForSnapshotReady(snapshot, port, defaultListenerWaitTimeout)
	}
	return common.Combine(reason, cause, restoreErr, restartRestoreErr, waitRestoreErr)
}

func (s *CertService) waitForSnapshotReady(snapshot *certSettingsSnapshot, port int, timeout time.Duration) error {
	if snapshot != nil && snapshot.certFile != "" && snapshot.keyFile != "" {
		return s.waitForHTTPSReady(port, timeout)
	}
	return s.waitForHTTPReady(port, timeout)
}

func (s *CertService) lookupPublicServerIPs() []string {
	if s.publicIPs != nil {
		return s.publicIPs()
	}
	return getPublicServerIPs()
}

func (s *CertService) ensurePortAvailable(port int) error {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("port %d is not available: %w", port, err)
	}
	return listener.Close()
}

func (s *CertService) ensureAcmeInstalled(email string) error {
	acmePath := s.acmeBinaryPath()
	if fileExists(acmePath) {
		return nil
	}
	scriptPath := "/usr/local/x-ui/scripts/acme_install.sh"
	if !fileExists(scriptPath) {
		return errors.New("acme install script not found")
	}
	args := []string{scriptPath}
	if email != "" {
		args = append(args, "email="+email)
	}
	if _, err := s.command("sh", args...); err != nil {
		return fmt.Errorf("install acme.sh failed: %w", err)
	}
	if !fileExists(acmePath) {
		return errors.New("acme.sh installation did not produce executable")
	}
	return nil
}

func (s *CertService) ensureSocatInstalled() error {
	if _, err := exec.LookPath("socat"); err == nil {
		return nil
	}
	osRelease, _ := os.ReadFile("/etc/os-release")
	content := strings.ToLower(string(osRelease))
	switch {
	case strings.Contains(content, "debian"), strings.Contains(content, "ubuntu"):
		if _, err := s.command("apt-get", "update"); err != nil {
			return fmt.Errorf("install socat failed during apt-get update: %w", err)
		}
		if _, err := s.command("apt-get", "install", "-y", "socat"); err != nil {
			return fmt.Errorf("install socat failed: %w", err)
		}
	case strings.Contains(content, "centos"), strings.Contains(content, "rocky"), strings.Contains(content, "alma"), strings.Contains(content, "rhel"):
		if _, err := s.command("sh", "-c", "command -v dnf >/dev/null 2>&1 && dnf install -y socat || yum install -y socat"); err != nil {
			return fmt.Errorf("install socat failed: %w", err)
		}
	default:
		return errors.New("socat is required for acme standalone mode; please install socat")
	}
	if _, err := exec.LookPath("socat"); err != nil {
		return errors.New("socat install completed but executable not found")
	}
	return nil
}

func (s *CertService) issueAcmeCertificate(domain string, staging bool) error {
	acmePath := s.acmeBinaryPath()
	if !staging {
		if _, err := s.command(acmePath, "--set-default-ca", "--server", "letsencrypt"); err != nil {
			return fmt.Errorf("set default acme ca failed: %w", err)
		}
	}
	args := []string{"--issue", "-d", domain, "--standalone", "--httpport", "80"}
	if staging {
		args = append(args, "--staging")
	}
	if _, err := s.command(acmePath, args...); err != nil {
		return fmt.Errorf("issue acme certificate failed: %w", err)
	}
	return nil
}

func (s *CertService) installAcmeCertificate(domain string, certFile string, keyFile string) error {
	acmePath := s.acmeBinaryPath()
	if _, err := s.command(acmePath, "--install-cert", "-d", domain, "--cert-file", certFile, "--key-file", keyFile); err != nil {
		return fmt.Errorf("install acme certificate failed: %w", err)
	}
	return nil
}

func (s *CertService) acmeBinaryPath() string {
	home := strings.TrimSpace(os.Getenv("HOME"))
	if home == "" {
		home = "/root"
	}
	return filepath.Join(home, ".acme.sh", "acme.sh")
}

func (s *CertService) command(name string, args ...string) ([]byte, error) {
	if s.runCommand != nil {
		return s.runCommand(name, args...)
	}
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("%s: %w", strings.TrimSpace(string(output)), err)
	}
	return output, nil
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

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email is required")
	}
	if strings.ContainsAny(email, " \t\r\n") || !strings.Contains(email, "@") {
		return errors.New("email format is invalid")
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

func waitForListener(timeout time.Duration, probe func() error) error {
	if timeout <= 0 {
		return errors.New("listener wait timeout must be positive")
	}
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		if err := probe(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(defaultListenerRetryDelay)
	}
	if lastErr == nil {
		lastErr = errors.New("listener probe failed")
	}
	return lastErr
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

func isHTTPSRedirectResponse(resp *http.Response) bool {
	if resp == nil {
		return false
	}
	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
	default:
		return false
	}
	location := strings.TrimSpace(resp.Header.Get("Location"))
	return strings.HasPrefix(strings.ToLower(location), "https://")
}

func loopbackAddr(port int) (string, error) {
	if port <= 0 {
		return "", errors.New("invalid panel port")
	}
	return fmt.Sprintf("127.0.0.1:%d", port), nil
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
