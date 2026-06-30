package service

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
	"x-ui/web/entity"
)

type CertService struct{}

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

	status := &entity.CertStatus{
		WebDomain:        webDomain,
		WebCertFile:      webCertFile,
		WebKeyFile:       webKeyFile,
		WebCertStatus:    webCertStatus,
		WebCertExpireAt:  webCertExpireAt,
		WebCertIssuer:    webCertIssuer,
		WebCertAutoRenew: webCertAutoRenew,
		WebCertMode:      webCertMode,
		WebCertProvider:  webCertProvider,
		Warnings:         []string{},
	}

	if webCertFile == "" {
		return status, nil
	}

	certInfo, certErr := readCertificateInfo(webCertFile)
	if certErr != nil {
		status.WebCertStatus = "unknown"
		status.Warnings = append(status.Warnings, certErr.Error())
		return status, nil
	}

	status.WebCertExpireAt = certInfo.NotAfter.Unix()
	if issuer := certInfo.Issuer.CommonName; issuer != "" {
		status.WebCertIssuer = issuer
	} else {
		status.WebCertIssuer = certInfo.Issuer.String()
	}
	if status.WebCertStatus == "" || status.WebCertStatus == "none" {
		status.WebCertStatus = "ready"
	}
	if !certInfo.NotAfter.IsZero() && certInfo.NotAfter.Before(time.Now()) {
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

	serverIPs := getPublicServerIPs()
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

func readCertificateInfo(certFile string) (*x509.Certificate, error) {
	data, err := os.ReadFile(certFile)
	if err != nil {
		return nil, errors.New("failed to read certificate file")
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.New("failed to parse certificate")
	}
	return cert, nil
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
