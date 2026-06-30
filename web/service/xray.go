package service

import (
	"encoding/json"
	"errors"
	"go.uber.org/atomic"
	"strings"
	"sync"
	"x-ui/database"
	"x-ui/database/model"
	"x-ui/logger"
	"x-ui/xray"

	"gorm.io/gorm"
)

var p *xray.Process
var lock sync.Mutex
var isNeedXrayRestart atomic.Bool
var result string

type XrayService struct {
	inboundService  InboundService
	outboundService OutboundService
	routingService  RoutingService
	settingService  SettingService
}

type routingConfig struct {
	DomainStrategy string        `json:"domainStrategy,omitempty"`
	Rules          []routingRule `json:"rules,omitempty"`
}

type routingRule struct {
	Type        string   `json:"type,omitempty"`
	Domain      []string `json:"domain,omitempty"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	IP          []string `json:"ip,omitempty"`
	Port        string   `json:"port,omitempty"`
	Protocol    []string `json:"protocol,omitempty"`
	Network     string   `json:"network,omitempty"`
	Source      []string `json:"source,omitempty"`
	OutboundTag string   `json:"outboundTag,omitempty"`
}

type outboundConfig struct {
	Tag            string      `json:"tag,omitempty"`
	Protocol       string      `json:"protocol"`
	Settings       interface{} `json:"settings"`
	StreamSettings interface{} `json:"streamSettings,omitempty"`
	Mux            interface{} `json:"mux,omitempty"`
}

const dokodemoFreedomTag = "dokodemo-freedom"

func (s *XrayService) IsXrayRunning() bool {
	return p != nil && p.IsRunning()
}

func (s *XrayService) GetXrayErr() error {
	if p == nil {
		return nil
	}
	return p.GetErr()
}

func (s *XrayService) GetXrayResult() string {
	if result != "" {
		return result
	}
	if s.IsXrayRunning() {
		return ""
	}
	if p == nil {
		return ""
	}
	result = p.GetResult()
	return result
}

func (s *XrayService) GetXrayVersion() string {
	if p == nil {
		return "Unknown"
	}
	return p.GetVersion()
}

func (s *XrayService) GetXrayConfig() (*xray.Config, error) {
	managed, err := s.settingService.GetManagedOutboundsRouting()
	if err != nil {
		return nil, err
	}
	if managed {
		return s.getManagedXrayConfig()
	}
	return s.getLegacyXrayConfig()
}

func (s *XrayService) getBaseXrayConfig() (*xray.Config, error) {
	templateConfig, err := s.settingService.GetXrayConfigTemplate()
	if err != nil {
		return nil, err
	}

	xrayConfig := &xray.Config{}
	if err := json.Unmarshal([]byte(templateConfig), xrayConfig); err != nil {
		return nil, err
	}

	accessIPService := AccessIPService{}
	if err := accessIPService.ApplyAccessLogSetting(xrayConfig); err != nil {
		return nil, err
	}
	return xrayConfig, nil
}

func (s *XrayService) getLegacyXrayConfig() (*xray.Config, error) {
	xrayConfig, err := s.getBaseXrayConfig()
	if err != nil {
		return nil, err
	}

	inbounds, err := s.inboundService.GetAllInbounds()
	if err != nil {
		return nil, err
	}
	allowPrivateOutboundNeeded := false
	for _, inbound := range inbounds {
		if !inbound.Enable {
			continue
		}
		inboundConfig := inbound.GenXrayInboundConfig()
		xrayConfig.InboundConfigs = append(xrayConfig.InboundConfigs, *inboundConfig)
		if inbound.Protocol == model.Dokodemo || inbound.Protocol == model.Tunnel {
			allowPrivateOutboundNeeded = true
		}
	}
	if allowPrivateOutboundNeeded {
		if err := s.ensureDokodemoTunnelRouting(xrayConfig, inbounds); err != nil {
			return nil, err
		}
	}
	return xrayConfig, nil
}

func (s *XrayService) getManagedXrayConfig() (*xray.Config, error) {
	return s.getManagedXrayConfigForDB(database.GetDB())
}

func (s *XrayService) getManagedXrayConfigForDB(db *gorm.DB) (*xray.Config, error) {
	xrayConfig, err := s.getBaseXrayConfig()
	if err != nil {
		return nil, err
	}

	inbounds, err := s.inboundService.GetAllInbounds()
	if err != nil {
		return nil, err
	}

	allowPrivateOutboundNeeded := false
	for _, inbound := range inbounds {
		if !inbound.Enable {
			continue
		}
		inboundConfig := inbound.GenXrayInboundConfig()
		xrayConfig.InboundConfigs = append(xrayConfig.InboundConfigs, *inboundConfig)
		if inbound.Protocol == model.Dokodemo || inbound.Protocol == model.Tunnel {
			allowPrivateOutboundNeeded = true
		}
	}

	outboundsRaw, err := s.buildManagedOutboundsForDB(db)
	if err != nil {
		return nil, err
	}
	xrayConfig.OutboundConfigs = outboundsRaw

	routingRaw, err := s.buildManagedRoutingForDB(db)
	if err != nil {
		return nil, err
	}
	xrayConfig.RouterConfig = routingRaw

	if allowPrivateOutboundNeeded {
		if err := s.ensureManagedDokodemoTunnelRouting(xrayConfig, inbounds); err != nil {
			return nil, err
		}
	}

	return xrayConfig, nil
}

func (s *XrayService) buildManagedOutbounds() ([]byte, error) {
	return s.buildManagedOutboundsForDB(database.GetDB())
}

func (s *XrayService) buildManagedOutboundsForDB(db *gorm.DB) ([]byte, error) {
	outbounds, err := s.outboundService.getOutboundsByDB(db, true)
	if err != nil {
		return nil, err
	}

	var direct *model.Outbound
	var blocked *model.Outbound
	userOutbounds := make([]model.Outbound, 0)
	for _, outbound := range outbounds {
		switch outbound.Tag {
		case "direct":
			tmp := outbound
			direct = &tmp
		case "blocked":
			tmp := outbound
			blocked = &tmp
		default:
			if !outbound.Enabled {
				continue
			}
			if outbound.Tag == "" || outbound.Tag == dokodemoFreedomTag {
				continue
			}
			userOutbounds = append(userOutbounds, outbound)
		}
	}

	if direct == nil {
		direct = &model.Outbound{
			Tag:      "direct",
			Protocol: "freedom",
			Settings: "{}",
		}
	}
	if blocked == nil {
		blocked = &model.Outbound{
			Tag:      "blocked",
			Protocol: "blackhole",
			Settings: "{}",
		}
	}

	configs := make([]outboundConfig, 0, 2+len(userOutbounds))
	directConfig, err := buildOutboundConfig(direct)
	if err != nil {
		return nil, err
	}
	blockedConfig, err := buildOutboundConfig(blocked)
	if err != nil {
		return nil, err
	}
	configs = append(configs, directConfig, blockedConfig)

	for _, outbound := range userOutbounds {
		if outbound.Tag == "direct" || outbound.Tag == "blocked" {
			continue
		}
		cfg, err := buildOutboundConfig(&outbound)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}

	return json.Marshal(configs)
}

func (s *XrayService) buildManagedRouting() ([]byte, error) {
	return s.buildManagedRoutingForDB(database.GetDB())
}

func (s *XrayService) buildManagedRoutingForDB(db *gorm.DB) ([]byte, error) {
	baseRouting := &routingConfig{}
	routingRaw := []byte(nil)

	xrayConfig, err := s.getBaseXrayConfig()
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(string(xrayConfig.RouterConfig))
	if raw != "" && raw != "null" {
		if err := json.Unmarshal(xrayConfig.RouterConfig, baseRouting); err != nil {
			return nil, err
		}
	}

	rules := make([]routingRule, 0)
	rules = append(rules, routingRule{
		Type:        "field",
		InboundTag:  []string{"api"},
		OutboundTag: "api",
	})

	dbRules, err := s.routingService.getRulesByDB(db, true)
	if err != nil {
		return nil, err
	}
	for _, rule := range dbRules {
		if !rule.Enabled {
			continue
		}
		cfg, err := buildRoutingRuleConfig(&rule)
		if err != nil {
			return nil, err
		}
		rules = append(rules, cfg)
	}

	routing := routingConfig{
		DomainStrategy: baseRouting.DomainStrategy,
		Rules:          rules,
	}
	routingRaw, err = json.Marshal(routing)
	if err != nil {
		return nil, err
	}
	return routingRaw, nil
}

func buildOutboundConfig(outbound *model.Outbound) (outboundConfig, error) {
	settings, err := decodeJSONObject(outbound.Settings)
	if err != nil {
		return outboundConfig{}, err
	}

	cfg := outboundConfig{
		Tag:      outbound.Tag,
		Protocol: outbound.Protocol,
		Settings: settings,
	}
	if strings.TrimSpace(outbound.StreamSettings) != "" {
		streamSettings, err := decodeJSONObject(outbound.StreamSettings)
		if err != nil {
			return outboundConfig{}, err
		}
		cfg.StreamSettings = streamSettings
	}
	if strings.TrimSpace(outbound.Mux) != "" {
		mux, err := decodeJSONObject(outbound.Mux)
		if err != nil {
			return outboundConfig{}, err
		}
		cfg.Mux = mux
	}
	return cfg, nil
}

func buildRoutingRuleConfig(rule *model.RoutingRule) (routingRule, error) {
	cfg := routingRule{
		Type:        defaultRoutingRuleType(rule.Type),
		OutboundTag: rule.OutboundTag,
	}

	if values, err := decodeJSONArrayStrings(rule.Domain); err != nil {
		return routingRule{}, err
	} else if len(values) > 0 {
		cfg.Domain = values
	}
	if values, err := decodeJSONArrayStrings(rule.IP); err != nil {
		return routingRule{}, err
	} else if len(values) > 0 {
		cfg.IP = values
	}
	if values, err := decodeJSONArrayStrings(rule.Protocol); err != nil {
		return routingRule{}, err
	} else if len(values) > 0 {
		cfg.Protocol = values
	}
	if values, err := decodeJSONArrayStrings(rule.Source); err != nil {
		return routingRule{}, err
	} else if len(values) > 0 {
		cfg.Source = values
	}
	if values, err := decodeJSONArrayStrings(rule.InboundTag); err != nil {
		return routingRule{}, err
	} else if len(values) > 0 {
		cfg.InboundTag = values
	}
	if strings.TrimSpace(rule.Port) != "" {
		cfg.Port = strings.TrimSpace(rule.Port)
	}
	if strings.TrimSpace(rule.Network) != "" {
		cfg.Network = strings.TrimSpace(rule.Network)
	}
	return cfg, nil
}

func defaultRoutingRuleType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "field"
	}
	return value
}

func decodeJSONObject(raw string) (map[string]interface{}, error) {
	value := make(map[string]interface{})
	text := strings.TrimSpace(raw)
	if text == "" {
		text = "{}"
	}
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		return nil, err
	}
	return value, nil
}

func decodeJSONArrayStrings(raw string) ([]string, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil, nil
	}
	values := make([]string, 0)
	if err := json.Unmarshal([]byte(text), &values); err != nil {
		return nil, err
	}
	return values, nil
}

func (s *XrayService) ensureDokodemoTunnelRouting(xrayConfig *xray.Config, inbounds []*model.Inbound) error {
	routing := &routingConfig{}
	raw := string(xrayConfig.RouterConfig)
	if raw != "" && raw != "null" {
		if err := json.Unmarshal(xrayConfig.RouterConfig, routing); err != nil {
			return err
		}
	}

	tags := make([]string, 0)
	for _, inbound := range inbounds {
		if !inbound.Enable {
			continue
		}
		if inbound.Protocol == model.Dokodemo || inbound.Protocol == model.Tunnel {
			tags = append(tags, inbound.Tag)
		}
	}
	if len(tags) == 0 {
		return nil
	}

	if err := ensureTaggedFreedomOutbound(xrayConfig); err != nil {
		return err
	}

	rule := routingRule{
		Type:        "field",
		InboundTag:  tags,
		OutboundTag: dokodemoFreedomTag,
	}

	rules := make([]routingRule, 0, len(routing.Rules)+1)
	rules = append(rules, rule)
	rules = append(rules, routing.Rules...)
	routing.Rules = rules

	data, err := json.Marshal(routing)
	if err != nil {
		return err
	}
	xrayConfig.RouterConfig = data
	return nil
}

func (s *XrayService) ensureManagedDokodemoTunnelRouting(xrayConfig *xray.Config, inbounds []*model.Inbound) error {
	routing := &routingConfig{}
	raw := string(xrayConfig.RouterConfig)
	if raw != "" && raw != "null" {
		if err := json.Unmarshal(xrayConfig.RouterConfig, routing); err != nil {
			return err
		}
	}

	tags := make([]string, 0)
	for _, inbound := range inbounds {
		if !inbound.Enable {
			continue
		}
		if inbound.Protocol == model.Dokodemo || inbound.Protocol == model.Tunnel {
			tags = append(tags, inbound.Tag)
		}
	}
	if len(tags) == 0 {
		return nil
	}

	if err := ensureTaggedFreedomOutbound(xrayConfig); err != nil {
		return err
	}

	rule := routingRule{
		Type:        "field",
		InboundTag:  tags,
		OutboundTag: dokodemoFreedomTag,
	}

	rules := make([]routingRule, 0, len(routing.Rules)+1)
	if len(routing.Rules) > 0 {
		rules = append(rules, routing.Rules[0])
		rules = append(rules, rule)
		rules = append(rules, routing.Rules[1:]...)
	} else {
		rules = append(rules, rule)
	}
	routing.Rules = rules

	data, err := json.Marshal(routing)
	if err != nil {
		return err
	}
	xrayConfig.RouterConfig = data
	return nil
}

func ensureTaggedFreedomOutbound(xrayConfig *xray.Config) error {
	outbounds := make([]map[string]interface{}, 0)
	raw := string(xrayConfig.OutboundConfigs)
	if raw != "" && raw != "null" {
		if err := json.Unmarshal(xrayConfig.OutboundConfigs, &outbounds); err != nil {
			return err
		}
	}

	for _, outbound := range outbounds {
		if tag, ok := outbound["tag"].(string); ok && tag == dokodemoFreedomTag {
			return nil
		}
	}

	outbounds = append(outbounds, map[string]interface{}{
		"protocol": "freedom",
		"settings": map[string]interface{}{},
		"tag":      dokodemoFreedomTag,
	})

	data, err := json.Marshal(outbounds)
	if err != nil {
		return err
	}
	xrayConfig.OutboundConfigs = data
	return nil
}

func (s *XrayService) ValidateManagedXrayConfig() error {
	xrayConfig, err := s.GetXrayConfig()
	if err != nil {
		return err
	}
	return xray.TestConfig(xrayConfig)
}

func (s *XrayService) validateManagedXrayConfigForDB(db *gorm.DB) error {
	xrayConfig, err := s.getManagedXrayConfigForDB(db)
	if err != nil {
		return err
	}
	return xray.TestConfig(xrayConfig)
}

func (s *XrayService) ApplyManagedXrayIfEnabled() error {
	managed, err := s.settingService.GetManagedOutboundsRouting()
	if err != nil {
		return err
	}
	if !managed {
		return nil
	}
	if err := s.ValidateManagedXrayConfig(); err != nil {
		return err
	}
	return s.RestartXray(false)
}

func (s *XrayService) GetXrayTraffic() ([]*xray.Traffic, error) {
	if !s.IsXrayRunning() {
		return nil, errors.New("xray is not running")
	}
	return p.GetTraffic(true)
}

func (s *XrayService) RestartXray(isForce bool) error {
	lock.Lock()
	defer lock.Unlock()
	logger.Debug("restart xray, force:", isForce)

	xrayConfig, err := s.GetXrayConfig()
	if err != nil {
		return err
	}

	err = xray.TestConfig(xrayConfig)
	if err != nil {
		return err
	}

	var oldConfig *xray.Config
	if p != nil && p.IsRunning() {
		if !isForce && p.GetConfig().Equals(xrayConfig) {
			logger.Debug("not need to restart xray")
			return nil
		}
		oldConfig = p.GetConfig()
		p.Stop()
	}

	p = xray.NewProcess(xrayConfig)
	result = ""
	err = p.Start()
	if err == nil {
		return nil
	}
	if oldConfig != nil {
		logger.Warning("new xray start failed, trying rollback:", err)
		rollbackProcess := xray.NewProcess(oldConfig)
		rollbackErr := rollbackProcess.Start()
		if rollbackErr == nil {
			p = rollbackProcess
			return err
		}
		logger.Error("rollback xray start failed:", rollbackErr)
	}
	return err
}

func (s *XrayService) StopXray() error {
	lock.Lock()
	defer lock.Unlock()
	logger.Debug("stop xray")
	if s.IsXrayRunning() {
		return p.Stop()
	}
	return errors.New("xray is not running")
}

func (s *XrayService) SetToNeedRestart() {
	isNeedXrayRestart.Store(true)
}

func (s *XrayService) IsNeedRestartAndSetFalse() bool {
	return isNeedXrayRestart.CAS(true, false)
}
