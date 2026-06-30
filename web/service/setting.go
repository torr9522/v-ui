package service

import (
	_ "embed"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"x-ui/database"
	"x-ui/database/model"
	"x-ui/logger"
	"x-ui/util/common"
	"x-ui/util/random"
	"x-ui/util/reflect_util"
	"x-ui/web/entity"
)

//go:embed config.json
var xrayTemplateConfig string

var defaultValueMap = map[string]string{
	"xrayTemplateConfig":      xrayTemplateConfig,
	"managedOutboundsRouting": "false",
	"webListen":               "",
	"webPort":                 "54321",
	"webCertFile":             "",
	"webKeyFile":              "",
	"webDomain":               "",
	"webCertStatus":           "none",
	"webCertExpireAt":         "0",
	"webCertIssuer":           "",
	"webCertAutoRenew":        "false",
	"webCertMode":             "none",
	"webCertProvider":         "",
	"cloudflareApiToken":      "",
	"cloudflareZoneID":        "",
	"cloudflareAccountID":     "",
	"cloudflareEmail":         "",
	"cloudflareApiKey":        "",
	"cloudflareEnabled":       "false",
	"secret":                  random.Seq(32),
	"webBasePath":             "/",
	"timeLocation":            "Asia/Shanghai",
}

type SettingService struct {
}

func (s *SettingService) GetAllSetting() (*entity.AllSetting, error) {
	db := database.GetDB()
	settings := make([]*model.Setting, 0)
	err := db.Model(model.Setting{}).Find(&settings).Error
	if err != nil {
		return nil, err
	}
	allSetting := &entity.AllSetting{}
	t := reflect.TypeOf(allSetting).Elem()
	v := reflect.ValueOf(allSetting).Elem()
	fields := reflect_util.GetFields(t)

	setSetting := func(key, value string) (err error) {
		defer func() {
			panicErr := recover()
			if panicErr != nil {
				err = errors.New(fmt.Sprint(panicErr))
			}
		}()

		var found bool
		var field reflect.StructField
		for _, f := range fields {
			if f.Tag.Get("json") == key {
				field = f
				found = true
				break
			}
		}

		if !found {
			// 有些设置自动生成，不需要返回到前端给用户修改
			return nil
		}

		fieldV := v.FieldByName(field.Name)
		switch t := fieldV.Interface().(type) {
		case int:
			n, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return err
			}
			fieldV.SetInt(n)
		case int64:
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			fieldV.SetInt(n)
		case bool:
			b, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			fieldV.SetBool(b)
		case string:
			fieldV.SetString(value)
		default:
			return common.NewErrorf("unknown field %v type %v", key, t)
		}
		return
	}

	keyMap := map[string]bool{}
	for _, setting := range settings {
		err := setSetting(setting.Key, setting.Value)
		if err != nil {
			return nil, err
		}
		keyMap[setting.Key] = true
	}

	for key, value := range defaultValueMap {
		if keyMap[key] {
			continue
		}
		err := setSetting(key, value)
		if err != nil {
			return nil, err
		}
	}

	return allSetting, nil
}

func (s *SettingService) ResetSettings() error {
	db := database.GetDB()
	return db.Where("1 = 1").Delete(model.Setting{}).Error
}

func (s *SettingService) getSetting(key string) (*model.Setting, error) {
	db := database.GetDB()
	setting := &model.Setting{}
	err := db.Model(model.Setting{}).Where("key = ?", key).First(setting).Error
	if err != nil {
		return nil, err
	}
	return setting, nil
}

func (s *SettingService) saveSetting(key string, value string) error {
	setting, err := s.getSetting(key)
	db := database.GetDB()
	if database.IsNotFound(err) {
		return db.Create(&model.Setting{
			Key:   key,
			Value: value,
		}).Error
	} else if err != nil {
		return err
	}
	setting.Key = key
	setting.Value = value
	return db.Save(setting).Error
}

func (s *SettingService) getString(key string) (string, error) {
	setting, err := s.getSetting(key)
	if database.IsNotFound(err) {
		value, ok := defaultValueMap[key]
		if !ok {
			return "", common.NewErrorf("key <%v> not in defaultValueMap", key)
		}
		return value, nil
	} else if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (s *SettingService) setString(key string, value string) error {
	return s.saveSetting(key, value)
}

func (s *SettingService) getInt(key string) (int, error) {
	str, err := s.getString(key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(str)
}

func (s *SettingService) setInt(key string, value int) error {
	return s.setString(key, strconv.Itoa(value))
}

func (s *SettingService) getInt64(key string) (int64, error) {
	str, err := s.getString(key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(str, 10, 64)
}

func (s *SettingService) setInt64(key string, value int64) error {
	return s.setString(key, strconv.FormatInt(value, 10))
}

func (s *SettingService) getBool(key string) (bool, error) {
	str, err := s.getString(key)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(str)
}

func (s *SettingService) GetManagedOutboundsRouting() (bool, error) {
	return s.getBool("managedOutboundsRouting")
}

func (s *SettingService) GetXrayConfigTemplate() (string, error) {
	return s.getString("xrayTemplateConfig")
}

func (s *SettingService) GetListen() (string, error) {
	return s.getString("webListen")
}

func (s *SettingService) GetPort() (int, error) {
	return s.getInt("webPort")
}

func (s *SettingService) SetPort(port int) error {
	return s.setInt("webPort", port)
}

func (s *SettingService) GetCertFile() (string, error) {
	return s.getString("webCertFile")
}

func (s *SettingService) SetCertFile(path string) error {
	return s.setString("webCertFile", path)
}

func (s *SettingService) GetKeyFile() (string, error) {
	return s.getString("webKeyFile")
}

func (s *SettingService) SetKeyFile(path string) error {
	return s.setString("webKeyFile", path)
}

func (s *SettingService) GetWebDomain() (string, error) {
	return s.getString("webDomain")
}

func (s *SettingService) SetWebDomain(domain string) error {
	return s.setString("webDomain", domain)
}

func (s *SettingService) GetWebCertStatus() (string, error) {
	return s.getString("webCertStatus")
}

func (s *SettingService) SetWebCertStatus(status string) error {
	return s.setString("webCertStatus", status)
}

func (s *SettingService) GetWebCertExpireAt() (int64, error) {
	return s.getInt64("webCertExpireAt")
}

func (s *SettingService) SetWebCertExpireAt(expireAt int64) error {
	return s.setInt64("webCertExpireAt", expireAt)
}

func (s *SettingService) GetWebCertIssuer() (string, error) {
	return s.getString("webCertIssuer")
}

func (s *SettingService) SetWebCertIssuer(issuer string) error {
	return s.setString("webCertIssuer", issuer)
}

func (s *SettingService) GetWebCertAutoRenew() (bool, error) {
	return s.getBool("webCertAutoRenew")
}

func (s *SettingService) SetWebCertAutoRenew(autoRenew bool) error {
	return s.setString("webCertAutoRenew", strconv.FormatBool(autoRenew))
}

func (s *SettingService) GetWebCertMode() (string, error) {
	return s.getString("webCertMode")
}

func (s *SettingService) SetWebCertMode(mode string) error {
	return s.setString("webCertMode", mode)
}

func (s *SettingService) GetWebCertProvider() (string, error) {
	return s.getString("webCertProvider")
}

func (s *SettingService) SetWebCertProvider(provider string) error {
	return s.setString("webCertProvider", provider)
}

func (s *SettingService) GetCloudflareAPIToken() (string, error) {
	return s.getString("cloudflareApiToken")
}

func (s *SettingService) SetCloudflareAPIToken(value string) error {
	return s.setString("cloudflareApiToken", strings.TrimSpace(value))
}

func (s *SettingService) GetCloudflareZoneID() (string, error) {
	return s.getString("cloudflareZoneID")
}

func (s *SettingService) SetCloudflareZoneID(value string) error {
	return s.setString("cloudflareZoneID", strings.TrimSpace(value))
}

func (s *SettingService) GetCloudflareAccountID() (string, error) {
	return s.getString("cloudflareAccountID")
}

func (s *SettingService) SetCloudflareAccountID(value string) error {
	return s.setString("cloudflareAccountID", strings.TrimSpace(value))
}

func (s *SettingService) GetCloudflareEmail() (string, error) {
	return s.getString("cloudflareEmail")
}

func (s *SettingService) SetCloudflareEmail(value string) error {
	return s.setString("cloudflareEmail", strings.TrimSpace(value))
}

func (s *SettingService) GetCloudflareAPIKey() (string, error) {
	return s.getString("cloudflareApiKey")
}

func (s *SettingService) SetCloudflareAPIKey(value string) error {
	return s.setString("cloudflareApiKey", strings.TrimSpace(value))
}

func (s *SettingService) GetCloudflareEnabled() (bool, error) {
	return s.getBool("cloudflareEnabled")
}

func (s *SettingService) SetCloudflareEnabled(value bool) error {
	return s.setString("cloudflareEnabled", strconv.FormatBool(value))
}

func (s *SettingService) GetSecret() ([]byte, error) {
	secret, err := s.getString("secret")
	if secret == defaultValueMap["secret"] {
		err := s.saveSetting("secret", secret)
		if err != nil {
			logger.Warning("save secret failed:", err)
		}
	}
	return []byte(secret), err
}

func (s *SettingService) GetBasePath() (string, error) {
	basePath, err := s.getString("webBasePath")
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	return basePath, nil
}

func (s *SettingService) GetTimeLocation() (*time.Location, error) {
	l, err := s.getString("timeLocation")
	if err != nil {
		return nil, err
	}
	location, err := time.LoadLocation(l)
	if err != nil {
		defaultLocation := defaultValueMap["timeLocation"]
		logger.Errorf("location <%v> not exist, using default location: %v", l, defaultLocation)
		return time.LoadLocation(defaultLocation)
	}
	return location, nil
}

func (s *SettingService) UpdateAllSetting(allSetting *entity.AllSetting) error {
	if err := allSetting.CheckValid(); err != nil {
		return err
	}

	v := reflect.ValueOf(allSetting).Elem()
	t := reflect.TypeOf(allSetting).Elem()
	fields := reflect_util.GetFields(t)
	errs := make([]error, 0)
	for _, field := range fields {
		key := field.Tag.Get("json")
		fieldV := v.FieldByName(field.Name)
		value := fmt.Sprint(fieldV.Interface())
		err := s.saveSetting(key, value)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return common.Combine(errs...)
}
