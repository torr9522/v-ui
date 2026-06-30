package service

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"x-ui/database"
	"x-ui/database/model"

	"gorm.io/gorm"
)

var outboundTagRegexp = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

var allowedOutboundProtocols = map[string]struct{}{
	"freedom":     {},
	"blackhole":   {},
	"socks":       {},
	"vmess":       {},
	"vless":       {},
	"trojan":      {},
	"shadowsocks": {},
}

type OutboundService struct{}

func (s *OutboundService) GetAllOutbounds() ([]model.Outbound, error) {
	return s.GetOutbounds(true)
}

func (s *OutboundService) GetOutbounds(includeSystem bool) ([]model.Outbound, error) {
	return s.getOutboundsByDB(database.GetDB(), includeSystem)
}

func (s *OutboundService) getOutboundsByDB(db *gorm.DB, includeSystem bool) ([]model.Outbound, error) {
	outbounds := make([]model.Outbound, 0)
	query := db.Order("sort asc, id asc")
	if !includeSystem {
		query = query.Where("is_system = ?", false)
	}
	err := query.Find(&outbounds).Error
	if err != nil {
		return nil, err
	}
	return outbounds, nil
}

func (s *OutboundService) GetOutbound(id int) (*model.Outbound, error) {
	return s.getOutboundByDB(database.GetDB(), id)
}

func (s *OutboundService) getOutboundByDB(db *gorm.DB, id int) (*model.Outbound, error) {
	outbound := &model.Outbound{}
	err := db.First(outbound, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("OUTBOUND_NOT_FOUND: outbound not found")
		}
		return nil, err
	}
	return outbound, nil
}

func (s *OutboundService) IsOutboundTagExists(tag string, excludeId int) (bool, error) {
	return s.isOutboundTagExistsByDB(database.GetDB(), tag, excludeId)
}

func (s *OutboundService) isOutboundTagExistsByDB(db *gorm.DB, tag string, excludeId int) (bool, error) {
	query := db.Model(&model.Outbound{}).Where("tag = ?", tag)
	if excludeId > 0 {
		query = query.Where("id != ?", excludeId)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *OutboundService) ValidateOutbound(ob *model.Outbound) error {
	ob.Tag = strings.TrimSpace(ob.Tag)
	ob.Protocol = strings.TrimSpace(strings.ToLower(ob.Protocol))
	ob.Settings = strings.TrimSpace(ob.Settings)
	ob.StreamSettings = strings.TrimSpace(ob.StreamSettings)
	ob.Mux = strings.TrimSpace(ob.Mux)
	ob.Remark = strings.TrimSpace(ob.Remark)

	if ob.Tag == "" {
		return errors.New("OUTBOUND_TAG_REQUIRED: tag is required")
	}
	if len(ob.Tag) > 64 {
		return errors.New("OUTBOUND_TAG_INVALID: tag format is invalid")
	}
	if !outboundTagRegexp.MatchString(ob.Tag) {
		return errors.New("OUTBOUND_TAG_INVALID: tag format is invalid")
	}
	if strings.EqualFold(ob.Tag, "api") {
		return errors.New("OUTBOUND_TAG_FORBIDDEN: tag api is reserved")
	}
	if _, ok := allowedOutboundProtocols[ob.Protocol]; !ok {
		return errors.New("OUTBOUND_PROTOCOL_INVALID: protocol is not supported")
	}
	if err := validateJSONObjectString(ob.Settings); err != nil {
		return errors.New("OUTBOUND_SETTINGS_INVALID: settings must be a valid JSON object")
	}
	if ob.StreamSettings != "" {
		if err := validateJSONObjectString(ob.StreamSettings); err != nil {
			return errors.New("OUTBOUND_STREAM_SETTINGS_INVALID: streamSettings must be a valid JSON object")
		}
	}
	if ob.Mux != "" {
		if err := validateJSONObjectString(ob.Mux); err != nil {
			return errors.New("OUTBOUND_MUX_INVALID: mux must be a valid JSON object")
		}
	}
	if len(ob.Remark) > 255 {
		return errors.New("OUTBOUND_REMARK_TOO_LONG: remark is too long")
	}
	if ob.Sort == 0 {
		ob.Sort = 1000
	}
	return nil
}

func (s *OutboundService) AddOutbound(ob *model.Outbound) (bool, error) {
	if err := s.ValidateOutbound(ob); err != nil {
		return false, err
	}
	if isReservedSystemTag(ob.Tag) {
		return false, errors.New("OUTBOUND_TAG_RESERVED: reserved system tag")
	}
	if ob.IsSystem {
		return false, errors.New("OUTBOUND_IS_SYSTEM_FORBIDDEN: user cannot create system outbound")
	}

	tx := database.GetDB().Begin()
	exists, err := s.isOutboundTagExistsByDB(tx, ob.Tag, 0)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	if exists {
		tx.Rollback()
		return false, errors.New("OUTBOUND_TAG_DUPLICATE: tag already exists")
	}

	if err := tx.Create(ob).Error; err != nil {
		tx.Rollback()
		return false, err
	}
	return s.commitManagedChange(tx)
}

func (s *OutboundService) UpdateOutbound(ob *model.Outbound) (bool, error) {
	if ob.Id <= 0 {
		return false, errors.New("OUTBOUND_NOT_FOUND: outbound not found")
	}

	tx := database.GetDB().Begin()
	current, err := s.getOutboundByDB(tx, ob.Id)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	if current.IsSystem {
		if (ob.Tag != "" && ob.Tag != current.Tag) ||
			(strings.TrimSpace(ob.Protocol) != "" && strings.ToLower(strings.TrimSpace(ob.Protocol)) != current.Protocol) ||
			(strings.TrimSpace(ob.Settings) != "" && strings.TrimSpace(ob.Settings) != current.Settings) ||
			(strings.TrimSpace(ob.StreamSettings) != "" && strings.TrimSpace(ob.StreamSettings) != current.StreamSettings) ||
			(strings.TrimSpace(ob.Mux) != "" && strings.TrimSpace(ob.Mux) != current.Mux) {
			tx.Rollback()
			return false, errors.New("OUTBOUND_SYSTEM_UPDATE_RESTRICTED: system outbound only allows remark and sort changes")
		}

		current.Remark = strings.TrimSpace(ob.Remark)
		if len(current.Remark) > 255 {
			tx.Rollback()
			return false, errors.New("OUTBOUND_REMARK_TOO_LONG: remark is too long")
		}
		if ob.Sort != 0 || current.Sort == 0 {
			current.Sort = ob.Sort
		}
		if err := tx.Save(current).Error; err != nil {
			tx.Rollback()
			return false, err
		}
		return s.commitManagedChange(tx)
	}

	if ob.IsSystem {
		tx.Rollback()
		return false, errors.New("OUTBOUND_IS_SYSTEM_FORBIDDEN: user cannot create system outbound")
	}
	if err := s.ValidateOutbound(ob); err != nil {
		tx.Rollback()
		return false, err
	}
	if isReservedSystemTag(ob.Tag) {
		tx.Rollback()
		return false, errors.New("OUTBOUND_TAG_RESERVED: reserved system tag")
	}

	exists, err := s.isOutboundTagExistsByDB(tx, ob.Tag, ob.Id)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	if exists {
		tx.Rollback()
		return false, errors.New("OUTBOUND_TAG_DUPLICATE: tag already exists")
	}

	current.Tag = ob.Tag
	current.Protocol = ob.Protocol
	current.Settings = ob.Settings
	current.StreamSettings = ob.StreamSettings
	current.Mux = ob.Mux
	current.Enabled = ob.Enabled
	current.Remark = ob.Remark
	current.Sort = ob.Sort

	if err := tx.Save(current).Error; err != nil {
		tx.Rollback()
		return false, err
	}
	return s.commitManagedChange(tx)
}

func (s *OutboundService) DeleteOutbound(id int) (bool, error) {
	tx := database.GetDB().Begin()
	outbound, err := s.getOutboundByDB(tx, id)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	if outbound.IsSystem || isReservedSystemTag(outbound.Tag) {
		tx.Rollback()
		return false, errors.New("OUTBOUND_SYSTEM_DELETE_FORBIDDEN: system outbound cannot be deleted")
	}
	refCount, err := s.countOutboundRefsByDB(tx, outbound.Tag)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	if refCount > 0 {
		tx.Rollback()
		return false, errors.New("OUTBOUND_IN_USE: outbound is referenced by routing rules")
	}
	if err := tx.Delete(&model.Outbound{}, id).Error; err != nil {
		tx.Rollback()
		return false, err
	}
	return s.commitManagedChange(tx)
}

func (s *OutboundService) ToggleOutbound(id int, enabled bool) (bool, error) {
	tx := database.GetDB().Begin()
	outbound, err := s.getOutboundByDB(tx, id)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	if !enabled && (outbound.IsSystem || isReservedSystemTag(outbound.Tag)) {
		tx.Rollback()
		return false, errors.New("OUTBOUND_SYSTEM_DISABLE_FORBIDDEN: system outbound cannot be disabled")
	}
	if err := tx.Model(&model.Outbound{}).Where("id = ?", id).Update("enabled", enabled).Error; err != nil {
		tx.Rollback()
		return false, err
	}
	return s.commitManagedChange(tx)
}

func isReservedSystemTag(tag string) bool {
	return strings.EqualFold(tag, "direct") || strings.EqualFold(tag, "blocked")
}

func (s *OutboundService) CountOutboundRefs(tag string) (int64, error) {
	return s.countOutboundRefsByDB(database.GetDB(), tag)
}

func (s *OutboundService) countOutboundRefsByDB(db *gorm.DB, tag string) (int64, error) {
	var count int64
	err := db.Model(&model.RoutingRule{}).Where("outbound_tag = ?", tag).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *OutboundService) commitManagedChange(tx *gorm.DB) (bool, error) {
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

func validateJSONObjectString(raw string) error {
	var value interface{}
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return err
	}
	if _, ok := value.(map[string]interface{}); !ok {
		return errors.New("not json object")
	}
	return nil
}
