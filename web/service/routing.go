package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"x-ui/database"
	"x-ui/database/model"

	"gorm.io/gorm"
)

type RoutingService struct{}

type RoutingReorderItem struct {
	Id   int `json:"id"`
	Sort int `json:"sort"`
}

type aiTemplateSpec struct {
	Name   string
	Sort   int
	Domain []string
}

func (s *RoutingService) GetRules(includeSystem bool) ([]model.RoutingRule, error) {
	return s.getRulesByDB(database.GetDB(), includeSystem)
}

func (s *RoutingService) getRulesByDB(db *gorm.DB, includeSystem bool) ([]model.RoutingRule, error) {
	rules := make([]model.RoutingRule, 0)
	query := db.Order("sort asc, id asc")
	if !includeSystem {
		query = query.Where("is_system = ?", false)
	}
	if err := query.Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (s *RoutingService) GetRule(id int) (*model.RoutingRule, error) {
	return s.getRuleByDB(database.GetDB(), id)
}

func (s *RoutingService) getRuleByDB(db *gorm.DB, id int) (*model.RoutingRule, error) {
	rule := &model.RoutingRule{}
	err := db.First(rule, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ROUTING_RULE_NOT_FOUND: routing rule not found")
		}
		return nil, err
	}
	return rule, nil
}

func (s *RoutingService) AddRule(rule *model.RoutingRule) (bool, error) {
	if rule.IsSystem {
		return false, errors.New("ROUTING_RULE_IS_SYSTEM_FORBIDDEN: user cannot create system routing rule")
	}
	if err := s.ValidateRule(rule); err != nil {
		return false, err
	}
	tx := database.GetDB().Begin()
	if err := tx.Select("*").Create(rule).Error; err != nil {
		tx.Rollback()
		return false, err
	}
	return s.commitManagedChange(tx)
}

func (s *RoutingService) UpdateRule(rule *model.RoutingRule) (bool, error) {
	if rule.Id <= 0 {
		return false, errors.New("ROUTING_RULE_NOT_FOUND: routing rule not found")
	}

	tx := database.GetDB().Begin()
	current, err := s.getRuleByDB(tx, rule.Id)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	if current.IsSystem {
		if (strings.TrimSpace(rule.Type) != "" && strings.TrimSpace(rule.Type) != current.Type) ||
			(strings.TrimSpace(rule.Domain) != "" && strings.TrimSpace(rule.Domain) != current.Domain) ||
			(strings.TrimSpace(rule.IP) != "" && strings.TrimSpace(rule.IP) != current.IP) ||
			(strings.TrimSpace(rule.Port) != "" && strings.TrimSpace(rule.Port) != current.Port) ||
			(strings.TrimSpace(rule.Protocol) != "" && strings.TrimSpace(rule.Protocol) != current.Protocol) ||
			(strings.TrimSpace(rule.Network) != "" && strings.TrimSpace(rule.Network) != current.Network) ||
			(strings.TrimSpace(rule.Source) != "" && strings.TrimSpace(rule.Source) != current.Source) ||
			(strings.TrimSpace(rule.InboundTag) != "" && strings.TrimSpace(rule.InboundTag) != current.InboundTag) ||
			(strings.TrimSpace(rule.OutboundTag) != "" && strings.TrimSpace(rule.OutboundTag) != current.OutboundTag) {
			tx.Rollback()
			return false, errors.New("ROUTING_RULE_SYSTEM_UPDATE_RESTRICTED: system routing rule only allows remark and sort changes")
		}
		current.Remark = strings.TrimSpace(rule.Remark)
		if len(current.Remark) > 255 {
			tx.Rollback()
			return false, errors.New("ROUTING_RULE_REMARK_TOO_LONG: remark is too long")
		}
		if rule.Sort != 0 || current.Sort == 0 {
			current.Sort = rule.Sort
		}
		if err := tx.Model(&model.RoutingRule{}).Where("id = ?", current.Id).Updates(map[string]interface{}{
			"remark": current.Remark,
			"sort":   current.Sort,
		}).Error; err != nil {
			tx.Rollback()
			return false, err
		}
		return s.commitManagedChange(tx)
	}

	if rule.IsSystem {
		tx.Rollback()
		return false, errors.New("ROUTING_RULE_IS_SYSTEM_FORBIDDEN: user cannot create system routing rule")
	}
	if err := s.ValidateRule(rule); err != nil {
		tx.Rollback()
		return false, err
	}

	current.Type = rule.Type
	current.Domain = rule.Domain
	current.IP = rule.IP
	current.Port = rule.Port
	current.Protocol = rule.Protocol
	current.Network = rule.Network
	current.Source = rule.Source
	current.InboundTag = rule.InboundTag
	current.OutboundTag = rule.OutboundTag
	current.Enabled = rule.Enabled
	current.Remark = rule.Remark
	current.Sort = rule.Sort

	if err := tx.Model(&model.RoutingRule{}).Where("id = ?", current.Id).Updates(map[string]interface{}{
		"type":         current.Type,
		"domain":       current.Domain,
		"ip":           current.IP,
		"port":         current.Port,
		"protocol":     current.Protocol,
		"network":      current.Network,
		"source":       current.Source,
		"inbound_tag":  current.InboundTag,
		"outbound_tag": current.OutboundTag,
		"enabled":      current.Enabled,
		"remark":       current.Remark,
		"sort":         current.Sort,
	}).Error; err != nil {
		tx.Rollback()
		return false, err
	}
	return s.commitManagedChange(tx)
}

func (s *RoutingService) DeleteRule(id int) (bool, error) {
	tx := database.GetDB().Begin()
	rule, err := s.getRuleByDB(tx, id)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	if rule.IsSystem {
		tx.Rollback()
		return false, errors.New("ROUTING_RULE_SYSTEM_DELETE_FORBIDDEN: system routing rule cannot be deleted")
	}
	if err := tx.Delete(&model.RoutingRule{}, id).Error; err != nil {
		tx.Rollback()
		return false, err
	}
	return s.commitManagedChange(tx)
}

func (s *RoutingService) ToggleRule(id int, enabled bool) (bool, error) {
	tx := database.GetDB().Begin()
	rule, err := s.getRuleByDB(tx, id)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	if !enabled && rule.IsSystem {
		tx.Rollback()
		return false, errors.New("ROUTING_RULE_SYSTEM_DISABLE_FORBIDDEN: system routing rule cannot be disabled")
	}
	if err := tx.Model(&model.RoutingRule{}).Where("id = ?", id).Update("enabled", enabled).Error; err != nil {
		tx.Rollback()
		return false, err
	}
	return s.commitManagedChange(tx)
}

func (s *RoutingService) ReorderRules(items []RoutingReorderItem) (bool, error) {
	if len(items) == 0 {
		return false, nil
	}
	tx := database.GetDB().Begin()
	for _, item := range items {
		rule, err := s.getRuleByDB(tx, item.Id)
		if err != nil {
			tx.Rollback()
			return false, err
		}
		rule.Sort = item.Sort
		if rule.Sort == 0 {
			rule.Sort = 1000
		}
		if err := tx.Model(&model.RoutingRule{}).Where("id = ?", item.Id).Update("sort", rule.Sort).Error; err != nil {
			tx.Rollback()
			return false, err
		}
	}
	return s.commitManagedChange(tx)
}

func (s *RoutingService) ValidateRule(rule *model.RoutingRule) error {
	rule.Type = strings.TrimSpace(rule.Type)
	rule.Domain = strings.TrimSpace(rule.Domain)
	rule.IP = strings.TrimSpace(rule.IP)
	rule.Port = strings.TrimSpace(rule.Port)
	rule.Protocol = strings.TrimSpace(rule.Protocol)
	rule.Network = strings.TrimSpace(rule.Network)
	rule.Source = strings.TrimSpace(rule.Source)
	rule.InboundTag = strings.TrimSpace(rule.InboundTag)
	rule.OutboundTag = strings.TrimSpace(rule.OutboundTag)
	rule.Remark = strings.TrimSpace(rule.Remark)

	if rule.Type == "" {
		rule.Type = "field"
	}
	if rule.Type != "field" {
		return errors.New("ROUTING_RULE_TYPE_INVALID: type must be field")
	}
	if rule.OutboundTag == "" {
		return errors.New("ROUTING_RULE_OUTBOUND_TAG_REQUIRED: outboundTag is required")
	}
	exists, err := outboundTagExists(rule.OutboundTag)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("ROUTING_RULE_OUTBOUND_TAG_NOT_FOUND: outboundTag not found")
	}
	if rule.Domain == "" &&
		rule.IP == "" &&
		rule.Port == "" &&
		rule.Protocol == "" &&
		rule.Network == "" &&
		rule.Source == "" &&
		rule.InboundTag == "" {
		return errors.New("ROUTING_RULE_MATCH_REQUIRED: at least one match condition is required")
	}
	if err := validateJSONArrayString(rule.Domain); err != nil {
		return errors.New("ROUTING_RULE_DOMAIN_INVALID: domain must be a valid JSON array")
	}
	if err := validateJSONArrayString(rule.IP); err != nil {
		return errors.New("ROUTING_RULE_IP_INVALID: ip must be a valid JSON array")
	}
	if err := validateJSONArrayString(rule.Protocol); err != nil {
		return errors.New("ROUTING_RULE_PROTOCOL_INVALID: protocol must be a valid JSON array")
	}
	if err := validateJSONArrayString(rule.Source); err != nil {
		return errors.New("ROUTING_RULE_SOURCE_INVALID: source must be a valid JSON array")
	}
	if err := validateJSONArrayString(rule.InboundTag); err != nil {
		return errors.New("ROUTING_RULE_INBOUND_TAG_INVALID: inboundTag must be a valid JSON array")
	}
	if len(rule.Remark) > 255 {
		return errors.New("ROUTING_RULE_REMARK_TOO_LONG: remark is too long")
	}
	if rule.Sort == 0 {
		rule.Sort = 1000
	}
	return nil
}

func outboundTagExists(tag string) (bool, error) {
	return outboundTagExistsByDB(database.GetDB(), tag)
}

func outboundTagExistsByDB(db *gorm.DB, tag string) (bool, error) {
	var count int64
	err := db.Model(&model.Outbound{}).Where("tag = ?", tag).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *RoutingService) commitManagedChange(tx *gorm.DB) (bool, error) {
	xrayService := XrayService{}
	managed, err := xrayService.settingService.GetManagedOutboundsRouting()
	if err != nil {
		tx.Rollback()
		return false, err
	}
	if managed {
		if err := xrayService.validateManagedXrayConfigForDB(tx); err != nil {
			tx.Rollback()
			return false, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return false, err
	}
	if !managed {
		return false, nil
	}
	if err := xrayService.ApplyManagedXrayIfEnabled(); err != nil {
		return false, err
	}
	return true, nil
}

func (s *RoutingService) ApplyAITemplate(outboundTag string, sortBase int, remarkPrefix string) (created int, skipped int, applied bool, err error) {
	outboundTag = strings.TrimSpace(outboundTag)
	remarkPrefix = strings.TrimSpace(remarkPrefix)
	if outboundTag == "" {
		return 0, 0, false, errors.New("ROUTING_RULE_OUTBOUND_TAG_REQUIRED: outboundTag is required")
	}
	if sortBase == 0 {
		sortBase = 100
	}
	if remarkPrefix == "" {
		remarkPrefix = "AI"
	}

	tx := database.GetDB().Begin()
	exists, err := outboundTagExistsByDB(tx, outboundTag)
	if err != nil {
		tx.Rollback()
		return 0, 0, false, err
	}
	if !exists {
		tx.Rollback()
		return 0, 0, false, errors.New("ROUTING_RULE_OUTBOUND_TAG_NOT_FOUND: outboundTag not found")
	}

	candidates := s.buildAITemplateCandidates(outboundTag, sortBase, remarkPrefix)
	existingRules, err := s.getRulesByDB(tx, true)
	if err != nil {
		tx.Rollback()
		return 0, 0, false, err
	}

	fingerprints := make(map[string]struct{}, len(existingRules))
	for _, rule := range existingRules {
		fingerprints[s.routingRuleFingerprint(rule)] = struct{}{}
	}

	for _, candidate := range candidates {
		fp := s.routingRuleFingerprint(candidate)
		if _, ok := fingerprints[fp]; ok {
			skipped++
			continue
		}
		rule := candidate
		if err := tx.Select("*").Create(&rule).Error; err != nil {
			tx.Rollback()
			return created, skipped, false, err
		}
		fingerprints[fp] = struct{}{}
		created++
	}

	applied, err = s.commitManagedChange(tx)
	if err != nil {
		return 0, 0, false, err
	}
	return created, skipped, applied, nil
}

func (s *RoutingService) buildAITemplateCandidates(outboundTag string, sortBase int, remarkPrefix string) []model.RoutingRule {
	if sortBase == 0 {
		sortBase = 100
	}
	if strings.TrimSpace(remarkPrefix) == "" {
		remarkPrefix = "AI"
	}

	specs := []aiTemplateSpec{
		{
			Name: "OpenAI",
			Sort: sortBase + 10,
			Domain: []string{
				"geosite:openai",
				"domain:openai.com",
				"domain:chatgpt.com",
				"domain:oaistatic.com",
				"domain:oaiusercontent.com",
				"domain:openai.azure.com",
			},
		},
		{
			Name: "Anthropic",
			Sort: sortBase + 20,
			Domain: []string{
				"domain:anthropic.com",
				"domain:claude.ai",
			},
		},
		{
			Name: "Google AI",
			Sort: sortBase + 30,
			Domain: []string{
				"geosite:google",
				"domain:googleapis.com",
				"domain:generativelanguage.googleapis.com",
				"domain:ai.google.dev",
				"domain:makersuite.google.com",
				"domain:gemini.google.com",
			},
		},
		{
			Name: "xAI",
			Sort: sortBase + 40,
			Domain: []string{
				"domain:x.ai",
				"domain:grok.com",
			},
		},
		{
			Name: "Perplexity",
			Sort: sortBase + 50,
			Domain: []string{
				"domain:perplexity.ai",
			},
		},
		{
			Name: "Poe",
			Sort: sortBase + 60,
			Domain: []string{
				"domain:poe.com",
			},
		},
		{
			Name: "Cursor",
			Sort: sortBase + 70,
			Domain: []string{
				"domain:cursor.sh",
				"domain:cursor.com",
			},
		},
		{
			Name: "GitHub Copilot",
			Sort: sortBase + 80,
			Domain: []string{
				"geosite:github",
				"domain:github.com",
				"domain:githubusercontent.com",
				"domain:githubcopilot.com",
				"domain:copilot.microsoft.com",
			},
		},
		{
			Name: "HuggingFace",
			Sort: sortBase + 90,
			Domain: []string{
				"domain:huggingface.co",
				"domain:hf.space",
			},
		},
	}

	rules := make([]model.RoutingRule, 0, len(specs))
	for _, spec := range specs {
		domain, _ := json.Marshal(spec.Domain)
		rules = append(rules, model.RoutingRule{
			Type:        "field",
			Domain:      string(domain),
			OutboundTag: outboundTag,
			Enabled:     true,
			Remark:      fmt.Sprintf("%s %s", strings.TrimSpace(remarkPrefix), spec.Name),
			Sort:        spec.Sort,
		})
	}
	return rules
}

func (s *RoutingService) routingRuleFingerprint(rule model.RoutingRule) string {
	return strings.Join([]string{
		strings.TrimSpace(rule.Domain),
		strings.TrimSpace(rule.IP),
		strings.TrimSpace(rule.Port),
		strings.TrimSpace(rule.Protocol),
		strings.TrimSpace(rule.Network),
		strings.TrimSpace(rule.Source),
		strings.TrimSpace(rule.InboundTag),
		strings.TrimSpace(rule.OutboundTag),
	}, "||")
}

func validateJSONArrayString(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var value interface{}
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return err
	}
	if _, ok := value.([]interface{}); !ok {
		return errors.New("not json array")
	}
	return nil
}
