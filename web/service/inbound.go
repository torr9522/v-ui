package service

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"time"
	"x-ui/database"
	"x-ui/database/model"
	"x-ui/util/common"
	"x-ui/xray"
)

type InboundService struct {
}

type trojanInboundSettings struct {
	Clients   []map[string]interface{} `json:"clients"`
	Fallbacks []map[string]interface{} `json:"fallbacks"`
}

type vlessInboundSettings struct {
	Vlesses    []map[string]interface{} `json:"clients"`
	Decryption string                   `json:"decryption"`
	Fallbacks  []map[string]interface{} `json:"fallbacks"`
}

type socksInboundSettings struct {
	Auth     string                   `json:"auth"`
	Accounts []map[string]interface{} `json:"accounts"`
	Udp      interface{}              `json:"udp"`
	IP       string                   `json:"ip"`
}

var rateLimitRegexp = regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)(kbit|mbit|gbit|kbps|mbps|gbps|kbyte(?:/s)?|mbyte(?:/s)?|gbyte(?:/s)?)?$`)
var realityShortIDRegexp = regexp.MustCompile(`^[0-9a-fA-F]+$`)

func normalizeRate(rate string) string {
	rate = strings.TrimSpace(strings.ToLower(rate))
	rate = strings.ReplaceAll(rate, " ", "")
	if rate == "" {
		return ""
	}

	matches := rateLimitRegexp.FindStringSubmatch(rate)
	if len(matches) == 0 {
		return rate
	}

	unit := matches[2]
	if unit == "" {
		unit = "mbit"
	}
	switch unit {
	case "kbyte", "mbyte", "gbyte":
		unit += "/s"
	}

	return matches[1] + unit
}

func (s *InboundService) normalizeLimit(inbound *model.Inbound) {
	if inbound.IPLimit < 0 {
		inbound.IPLimit = 0
	}
	if inbound.IPTimeout <= 0 {
		inbound.IPTimeout = 5
	}
	if inbound.IPTimeout > 1440 {
		inbound.IPTimeout = 1440
	}
	inbound.PortRate = normalizeRate(inbound.PortRate)
	inbound.IPRate = normalizeRate(inbound.IPRate)
	if inbound.PortRate != "" && !rateLimitRegexp.MatchString(inbound.PortRate) {
		inbound.PortRate = ""
	}
	if inbound.IPRate != "" && !rateLimitRegexp.MatchString(inbound.IPRate) {
		inbound.IPRate = ""
	}
}

func (s *InboundService) normalizeProtocolSettings(inbound *model.Inbound) {
	settings := strings.TrimSpace(inbound.Settings)
	if settings == "" {
		return
	}

	switch inbound.Protocol {
	case model.Trojan:
		var trojan trojanInboundSettings
		if err := json.Unmarshal([]byte(settings), &trojan); err != nil {
			return
		}
		for _, client := range trojan.Clients {
			delete(client, "flow")
		}
		data, err := json.Marshal(trojan)
		if err != nil {
			return
		}
		inbound.Settings = string(data)
	case model.VLESS:
		var vless vlessInboundSettings
		if err := json.Unmarshal([]byte(settings), &vless); err != nil {
			return
		}
		for _, client := range vless.Vlesses {
			if flow, ok := client["flow"].(string); ok {
				if flow == "xtls-rprx-direct" || flow == "xtls-rprx-origin" {
					client["flow"] = ""
				}
			}
		}
		data, err := json.Marshal(vless)
		if err != nil {
			return
		}
		inbound.Settings = string(data)
	case model.Socks, model.Mixed:
		var socks socksInboundSettings
		if err := json.Unmarshal([]byte(settings), &socks); err != nil {
			return
		}
		socks.Auth = normalizeSocksAuth(socks.Auth, len(socks.Accounts) > 0)
		socks.Udp = normalizeSocksUDPValue(socks.Udp)
		if strings.TrimSpace(socks.IP) == "" {
			socks.IP = "127.0.0.1"
		}
		if socks.Auth != "password" {
			socks.Accounts = nil
		}
		data, err := json.Marshal(socks)
		if err != nil {
			return
		}
		inbound.Settings = string(data)
	default:
		return
	}
}

func normalizeSocksAuth(auth string, hasAccounts bool) string {
	auth = strings.TrimSpace(strings.ToLower(auth))
	switch auth {
	case "password":
		return "password"
	case "noauth":
		return "noauth"
	default:
		if hasAccounts {
			return "password"
		}
		return "noauth"
	}
}

func normalizeSocksUDPValue(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		switch strings.TrimSpace(strings.ToLower(v)) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off", "":
			return false
		}
	case float64:
		return v != 0
	case int:
		return v != 0
	case int64:
		return v != 0
	case nil:
		return true
	}
	return true
}

func trimJSONStringValue(value interface{}) string {
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return ""
}

func normalizeRealityStringList(value interface{}) []string {
	items := make([]string, 0)
	appendValue := func(text string) {
		text = strings.TrimSpace(text)
		if text != "" {
			items = append(items, text)
		}
	}

	switch v := value.(type) {
	case []string:
		for _, item := range v {
			appendValue(item)
		}
	case []interface{}:
		for _, item := range v {
			appendValue(trimJSONStringValue(item))
		}
	case string:
		normalized := strings.ReplaceAll(v, "\r\n", "\n")
		normalized = strings.ReplaceAll(normalized, `\n`, "\n")
		for _, item := range strings.Split(normalized, "\n") {
			appendValue(item)
		}
	}
	return items
}

func normalizeRealityMaxTimeDiff(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case float64:
		if v > 0 {
			return int64(v), true
		}
	case float32:
		if v > 0 {
			return int64(v), true
		}
	case int:
		if v > 0 {
			return int64(v), true
		}
	case int32:
		if v > 0 {
			return int64(v), true
		}
	case int64:
		if v > 0 {
			return v, true
		}
	case json.Number:
		if num, err := v.Int64(); err == nil && num > 0 {
			return num, true
		}
	}
	return 0, false
}

func normalizeRealityShow(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "on":
			return true
		}
	case float64:
		return v != 0
	case int:
		return v != 0
	case int64:
		return v != 0
	}
	return false
}

func normalizeRealityShareMetadata(stream map[string]interface{}, realitySettings map[string]interface{}) map[string]interface{} {
	realityShare, _ := stream["realityShare"].(map[string]interface{})

	publicKey := trimJSONStringValue(nil)
	shortID := trimJSONStringValue(nil)
	fingerprint := trimJSONStringValue(nil)

	if realityShare != nil {
		publicKey = trimJSONStringValue(realityShare["publicKey"])
		shortID = trimJSONStringValue(realityShare["shortId"])
		fingerprint = trimJSONStringValue(realityShare["fingerprint"])
	}

	if publicKey == "" {
		publicKey = trimJSONStringValue(realitySettings["publicKey"])
	}
	if shortID == "" {
		shortID = trimJSONStringValue(realitySettings["shortId"])
	}
	if fingerprint == "" {
		fingerprint = trimJSONStringValue(realitySettings["fingerprint"])
	}

	normalized := map[string]interface{}{}
	if publicKey != "" {
		normalized["publicKey"] = publicKey
	}
	if shortID != "" {
		normalized["shortId"] = shortID
	}
	if fingerprint != "" {
		normalized["fingerprint"] = fingerprint
	}
	return normalized
}

func validateRealityShortIDs(items []string) error {
	for _, item := range items {
		if len(item) == 0 {
			return common.NewError("REALITY shortIds 不能为空")
		}
		if len(item) > 16 {
			return common.NewError("REALITY shortId 长度不能超过 16:", item)
		}
		if len(item)%2 != 0 {
			return common.NewError("REALITY shortId 长度必须为偶数:", item)
		}
		if !realityShortIDRegexp.MatchString(item) {
			return common.NewError("REALITY shortId 必须是十六进制:", item)
		}
	}
	return nil
}

func (s *InboundService) normalizeStreamSettings(inbound *model.Inbound) error {
	raw := strings.TrimSpace(inbound.StreamSettings)
	if raw == "" {
		return nil
	}
	if !strings.Contains(strings.ToLower(raw), "reality") {
		return nil
	}

	stream := make(map[string]interface{})
	if err := json.Unmarshal([]byte(raw), &stream); err != nil {
		return err
	}

	security := strings.ToLower(trimJSONStringValue(stream["security"]))
	if security != "reality" {
		return nil
	}
	if inbound.Protocol != model.VLESS {
		return common.NewError("REALITY 第一轮仅支持 VLESS 入站")
	}

	network := strings.ToLower(trimJSONStringValue(stream["network"]))
	if network != "tcp" {
		return common.NewError("REALITY 第一轮仅支持 RAW(TCP) 传输")
	}

	if tcpSettings, ok := stream["tcpSettings"].(map[string]interface{}); ok {
		if header, ok := tcpSettings["header"].(map[string]interface{}); ok {
			headerType := strings.ToLower(trimJSONStringValue(header["type"]))
			if headerType != "" && headerType != "none" {
				return common.NewError("REALITY 第一轮不支持 TCP HTTP camouflage")
			}
		}
	}

	realitySettings, ok := stream["realitySettings"].(map[string]interface{})
	if !ok {
		return common.NewError("REALITY 缺少 realitySettings")
	}

	target := trimJSONStringValue(realitySettings["target"])
	if target == "" {
		target = trimJSONStringValue(realitySettings["dest"])
	}
	if target == "" {
		return common.NewError("REALITY target/dest 不能为空")
	}

	serverNames := normalizeRealityStringList(realitySettings["serverNames"])
	if len(serverNames) == 0 {
		return common.NewError("REALITY serverNames 至少需要一个值")
	}

	privateKey := trimJSONStringValue(realitySettings["privateKey"])
	if privateKey == "" {
		return common.NewError("REALITY privateKey 不能为空")
	}

	shortIDs := normalizeRealityStringList(realitySettings["shortIds"])
	if len(shortIDs) == 0 {
		return common.NewError("REALITY shortIds 至少需要一个值")
	}
	if err := validateRealityShortIDs(shortIDs); err != nil {
		return err
	}

	normalizedReality := map[string]interface{}{
		"target":      target,
		"serverNames": serverNames,
		"privateKey":  privateKey,
		"shortIds":    shortIDs,
		"show":        normalizeRealityShow(realitySettings["show"]),
	}
	if maxTimeDiff, ok := normalizeRealityMaxTimeDiff(realitySettings["maxTimeDiff"]); ok {
		normalizedReality["maxTimeDiff"] = maxTimeDiff
	}

	normalizedStream := map[string]interface{}{
		"network":         "tcp",
		"security":        "reality",
		"realitySettings": normalizedReality,
	}
	if realityShare := normalizeRealityShareMetadata(stream, realitySettings); len(realityShare) > 0 {
		normalizedStream["realityShare"] = realityShare
	}
	if tcpSettings, ok := stream["tcpSettings"].(map[string]interface{}); ok && len(tcpSettings) > 0 {
		normalizedStream["tcpSettings"] = tcpSettings
	}

	data, err := json.Marshal(normalizedStream)
	if err != nil {
		return err
	}
	inbound.StreamSettings = string(data)
	return nil
}

func (s *InboundService) CleanupLegacyTrojanSettings() error {
	db := database.GetDB()
	inbounds := make([]*model.Inbound, 0)
	err := db.Model(model.Inbound{}).Where("protocol = ?", model.Trojan).Find(&inbounds).Error
	if err != nil {
		return err
	}

	for _, inbound := range inbounds {
		oldSettings := inbound.Settings
		s.normalizeProtocolSettings(inbound)
		if inbound.Settings == oldSettings {
			continue
		}
		if err := db.Model(&model.Inbound{}).Where("id = ?", inbound.Id).Update("settings", inbound.Settings).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *InboundService) CleanupLegacyVlessSettings() error {
	db := database.GetDB()
	inbounds := make([]*model.Inbound, 0)
	err := db.Model(model.Inbound{}).Where("protocol = ?", model.VLESS).Find(&inbounds).Error
	if err != nil {
		return err
	}

	for _, inbound := range inbounds {
		oldSettings := inbound.Settings
		s.normalizeProtocolSettings(inbound)
		if inbound.Settings == oldSettings {
			continue
		}
		if err := db.Model(&model.Inbound{}).Where("id = ?", inbound.Id).Update("settings", inbound.Settings).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *InboundService) CleanupLegacySocksSettings() error {
	db := database.GetDB()
	inbounds := make([]*model.Inbound, 0)
	err := db.Model(model.Inbound{}).Where("protocol in ?", []model.Protocol{model.Socks, model.Mixed}).Find(&inbounds).Error
	if err != nil {
		return err
	}

	for _, inbound := range inbounds {
		oldSettings := inbound.Settings
		s.normalizeProtocolSettings(inbound)
		if inbound.Settings == oldSettings {
			continue
		}
		if err := db.Model(&model.Inbound{}).Where("id = ?", inbound.Id).Update("settings", inbound.Settings).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *InboundService) GetInbounds(userId int) ([]*model.Inbound, error) {
	db := database.GetDB()
	var inbounds []*model.Inbound
	err := db.Model(model.Inbound{}).Where("user_id = ?", userId).Find(&inbounds).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return inbounds, nil
}

func (s *InboundService) GetAllInbounds() ([]*model.Inbound, error) {
	db := database.GetDB()
	var inbounds []*model.Inbound
	err := db.Model(model.Inbound{}).Find(&inbounds).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return inbounds, nil
}

func (s *InboundService) checkPortExist(port int, ignoreId int) (bool, error) {
	db := database.GetDB()
	db = db.Model(model.Inbound{}).Where("port = ?", port)
	if ignoreId > 0 {
		db = db.Where("id != ?", ignoreId)
	}
	var count int64
	err := db.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *InboundService) AddInbound(inbound *model.Inbound) error {
	s.normalizeLimit(inbound)
	s.normalizeProtocolSettings(inbound)
	if err := s.normalizeStreamSettings(inbound); err != nil {
		return err
	}
	exist, err := s.checkPortExist(inbound.Port, 0)
	if err != nil {
		return err
	}
	if exist {
		return common.NewError("端口已存在:", inbound.Port)
	}
	db := database.GetDB()
	return db.Save(inbound).Error
}

func (s *InboundService) AddInbounds(inbounds []*model.Inbound) error {
	for _, inbound := range inbounds {
		s.normalizeLimit(inbound)
		s.normalizeProtocolSettings(inbound)
		if err := s.normalizeStreamSettings(inbound); err != nil {
			return err
		}
		exist, err := s.checkPortExist(inbound.Port, 0)
		if err != nil {
			return err
		}
		if exist {
			return common.NewError("端口已存在:", inbound.Port)
		}
	}

	db := database.GetDB()
	tx := db.Begin()
	var err error
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	for _, inbound := range inbounds {
		err = tx.Save(inbound).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *InboundService) DelInbound(id int) error {
	db := database.GetDB()
	return db.Delete(model.Inbound{}, id).Error
}

func (s *InboundService) GetInbound(id int) (*model.Inbound, error) {
	db := database.GetDB()
	inbound := &model.Inbound{}
	err := db.Model(model.Inbound{}).First(inbound, id).Error
	if err != nil {
		return nil, err
	}
	return inbound, nil
}

func (s *InboundService) UpdateInbound(inbound *model.Inbound) error {
	s.normalizeLimit(inbound)
	s.normalizeProtocolSettings(inbound)
	if err := s.normalizeStreamSettings(inbound); err != nil {
		return err
	}
	exist, err := s.checkPortExist(inbound.Port, inbound.Id)
	if err != nil {
		return err
	}
	if exist {
		return common.NewError("端口已存在:", inbound.Port)
	}

	oldInbound, err := s.GetInbound(inbound.Id)
	if err != nil {
		return err
	}
	oldInbound.Up = inbound.Up
	oldInbound.Down = inbound.Down
	oldInbound.Total = inbound.Total
	oldInbound.Remark = inbound.Remark
	oldInbound.Enable = inbound.Enable
	oldInbound.ExpiryTime = inbound.ExpiryTime
	oldInbound.IPLimit = inbound.IPLimit
	oldInbound.IPTimeout = inbound.IPTimeout
	oldInbound.PortRate = inbound.PortRate
	oldInbound.IPRate = inbound.IPRate
	oldInbound.Listen = inbound.Listen
	oldInbound.Port = inbound.Port
	oldInbound.Protocol = inbound.Protocol
	oldInbound.Settings = inbound.Settings
	oldInbound.StreamSettings = inbound.StreamSettings
	oldInbound.Sniffing = inbound.Sniffing
	oldInbound.Tag = fmt.Sprintf("inbound-%v", inbound.Port)

	db := database.GetDB()
	return db.Save(oldInbound).Error
}

func (s *InboundService) AddTraffic(traffics []*xray.Traffic) (err error) {
	if len(traffics) == 0 {
		return nil
	}
	db := database.GetDB()
	db = db.Model(model.Inbound{})
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	for _, traffic := range traffics {
		if traffic.IsInbound {
			err = tx.Where("tag = ?", traffic.Tag).
				UpdateColumn("up", gorm.Expr("up + ?", traffic.Up)).
				UpdateColumn("down", gorm.Expr("down + ?", traffic.Down)).
				Error
			if err != nil {
				return
			}
		}
	}
	return
}

func (s *InboundService) DisableInvalidInbounds() (int64, error) {
	db := database.GetDB()
	now := time.Now().Unix() * 1000
	result := db.Model(model.Inbound{}).
		Where("((total > 0 and up + down >= total) or (expiry_time > 0 and expiry_time <= ?)) and enable = ?", now, true).
		Update("enable", false)
	err := result.Error
	count := result.RowsAffected
	return count, err
}
