package entity

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"strings"
	"time"
	"x-ui/util/common"
	"x-ui/xray"
)

type Msg struct {
	Success bool        `json:"success"`
	Msg     string      `json:"msg"`
	Obj     interface{} `json:"obj"`
}

type Pager struct {
	Current  int         `json:"current"`
	PageSize int         `json:"page_size"`
	Total    int         `json:"total"`
	OrderBy  string      `json:"order_by"`
	Desc     bool        `json:"desc"`
	Key      string      `json:"key"`
	List     interface{} `json:"list"`
}

type AllSetting struct {
	WebListen   string `json:"webListen" form:"webListen"`
	WebPort     int    `json:"webPort" form:"webPort"`
	WebCertFile string `json:"webCertFile" form:"webCertFile"`
	WebKeyFile  string `json:"webKeyFile" form:"webKeyFile"`
	WebBasePath string `json:"webBasePath" form:"webBasePath"`
	WebDomain   string `json:"webDomain" form:"webDomain"`

	WebCertStatus    string `json:"webCertStatus" form:"webCertStatus"`
	WebCertExpireAt  int64  `json:"webCertExpireAt" form:"webCertExpireAt"`
	WebCertIssuer    string `json:"webCertIssuer" form:"webCertIssuer"`
	WebCertAutoRenew bool   `json:"webCertAutoRenew" form:"webCertAutoRenew"`
	WebCertMode      string `json:"webCertMode" form:"webCertMode"`
	WebCertProvider  string `json:"webCertProvider" form:"webCertProvider"`

	XrayTemplateConfig      string `json:"xrayTemplateConfig" form:"xrayTemplateConfig"`
	ManagedOutboundsRouting bool   `json:"managedOutboundsRouting" form:"managedOutboundsRouting"`

	TimeLocation string `json:"timeLocation" form:"timeLocation"`
}

func (s *AllSetting) CheckValid() error {
	if s.WebListen != "" {
		ip := net.ParseIP(s.WebListen)
		if ip == nil {
			return common.NewError("web listen is not valid ip:", s.WebListen)
		}
	}

	if s.WebPort <= 0 || s.WebPort > 65535 {
		return common.NewError("web port is not a valid port:", s.WebPort)
	}

	if s.WebCertFile != "" || s.WebKeyFile != "" {
		_, err := tls.LoadX509KeyPair(s.WebCertFile, s.WebKeyFile)
		if err != nil {
			return common.NewErrorf("cert file <%v> or key file <%v> invalid: %v", s.WebCertFile, s.WebKeyFile, err)
		}
	}

	if !strings.HasPrefix(s.WebBasePath, "/") {
		s.WebBasePath = "/" + s.WebBasePath
	}
	if !strings.HasSuffix(s.WebBasePath, "/") {
		s.WebBasePath += "/"
	}

	s.WebDomain = strings.TrimSpace(s.WebDomain)
	if s.WebDomain != "" {
		if strings.Contains(s.WebDomain, "://") || strings.Contains(s.WebDomain, "/") || strings.Contains(s.WebDomain, " ") {
			return common.NewError("web domain is invalid:", s.WebDomain)
		}
	}

	xrayConfig := &xray.Config{}
	err := json.Unmarshal([]byte(s.XrayTemplateConfig), xrayConfig)
	if err != nil {
		return common.NewError("xray template config invalid:", err)
	}

	_, err = time.LoadLocation(s.TimeLocation)
	if err != nil {
		return common.NewError("time location not exist:", s.TimeLocation)
	}

	return nil
}
