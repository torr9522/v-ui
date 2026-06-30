package model

import (
	"fmt"
	"x-ui/util/json_util"
	"x-ui/xray"
)

type Protocol string

const (
	VMess       Protocol = "vmess"
	VLESS       Protocol = "vless"
	Dokodemo    Protocol = "dokodemo-door"
	Http        Protocol = "http"
	Trojan      Protocol = "trojan"
	Shadowsocks Protocol = "shadowsocks"
	Socks       Protocol = "socks"
	Mixed       Protocol = "mixed"
	Tunnel      Protocol = "tunnel"
)

type User struct {
	Id       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Inbound struct {
	Id         int    `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	UserId     int    `json:"-"`
	Up         int64  `json:"up" form:"up"`
	Down       int64  `json:"down" form:"down"`
	Total      int64  `json:"total" form:"total"`
	Remark     string `json:"remark" form:"remark"`
	Enable     bool   `json:"enable" form:"enable"`
	ExpiryTime int64  `json:"expiryTime" form:"expiryTime"`
	IPLimit    int    `json:"ipLimit" form:"ipLimit" gorm:"default:0"`
	IPTimeout  int    `json:"ipTimeout" form:"ipTimeout" gorm:"default:5"`
	PortRate   string `json:"portRate" form:"portRate" gorm:"default:''"`
	IPRate     string `json:"ipRate" form:"ipRate" gorm:"default:''"`

	// config part
	Listen         string   `json:"listen" form:"listen"`
	Port           int      `json:"port" form:"port" gorm:"unique"`
	Protocol       Protocol `json:"protocol" form:"protocol"`
	Settings       string   `json:"settings" form:"settings"`
	StreamSettings string   `json:"streamSettings" form:"streamSettings"`
	Tag            string   `json:"tag" form:"tag" gorm:"unique"`
	Sniffing       string   `json:"sniffing" form:"sniffing"`
}

func (i *Inbound) GenXrayInboundConfig() *xray.InboundConfig {
	listen := i.Listen
	if listen != "" {
		listen = fmt.Sprintf("\"%v\"", listen)
	}
	return &xray.InboundConfig{
		Listen:         json_util.RawMessage(listen),
		Port:           i.Port,
		Protocol:       string(i.Protocol),
		Settings:       json_util.RawMessage(i.Settings),
		StreamSettings: json_util.RawMessage(i.StreamSettings),
		Tag:            i.Tag,
		Sniffing:       json_util.RawMessage(i.Sniffing),
	}
}

type Outbound struct {
	Id             int    `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Tag            string `json:"tag" form:"tag" gorm:"size:64;not null;uniqueIndex:uk_outbound_tag"`
	Protocol       string `json:"protocol" form:"protocol" gorm:"size:32;not null;index:idx_outbound_protocol"`
	Settings       string `json:"settings" form:"settings" gorm:"type:text;not null"`
	StreamSettings string `json:"streamSettings" form:"streamSettings" gorm:"type:text;default:''"`
	Mux            string `json:"mux" form:"mux" gorm:"type:text;default:''"`
	Enabled        bool   `json:"enabled" form:"enabled" gorm:"default:true;index:idx_outbound_enabled_sort,priority:1"`
	Remark         string `json:"remark" form:"remark" gorm:"size:255;default:''"`
	Sort           int    `json:"sort" form:"sort" gorm:"default:1000;index:idx_outbound_enabled_sort,priority:2;index:idx_outbound_sort_id,priority:1"`
	IsSystem       bool   `json:"isSystem" form:"isSystem" gorm:"default:false;index:idx_outbound_system"`
	CreatedAt      int64  `json:"createdAt" gorm:"autoCreateTime:milli;index:idx_outbound_created_at"`
	UpdatedAt      int64  `json:"updatedAt" gorm:"autoUpdateTime:milli;index:idx_outbound_updated_at"`
}

type RoutingRule struct {
	Id          int    `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Type        string `json:"type" form:"type" gorm:"size:16;not null;default:'field'"`
	Domain      string `json:"domain" form:"domain" gorm:"type:text;default:''"`
	IP          string `json:"ip" form:"ip" gorm:"type:text;default:''"`
	Port        string `json:"port" form:"port" gorm:"size:128;default:''"`
	Protocol    string `json:"protocol" form:"protocol" gorm:"type:text;default:''"`
	Network     string `json:"network" form:"network" gorm:"size:32;default:''"`
	Source      string `json:"source" form:"source" gorm:"type:text;default:''"`
	InboundTag  string `json:"inboundTag" form:"inboundTag" gorm:"type:text;default:''"`
	OutboundTag string `json:"outboundTag" form:"outboundTag" gorm:"size:64;not null;index:idx_routing_outbound_tag"`
	Enabled     bool   `json:"enabled" form:"enabled" gorm:"index:idx_routing_enabled_sort,priority:1"`
	Remark      string `json:"remark" form:"remark" gorm:"size:255;default:''"`
	Sort        int    `json:"sort" form:"sort" gorm:"default:1000;index:idx_routing_enabled_sort,priority:2;index:idx_routing_sort_id,priority:1"`
	IsSystem    bool   `json:"isSystem" form:"isSystem" gorm:"default:false;index:idx_routing_system"`
	CreatedAt   int64  `json:"createdAt" gorm:"autoCreateTime:milli;index:idx_routing_created_at"`
	UpdatedAt   int64  `json:"updatedAt" gorm:"autoUpdateTime:milli;index:idx_routing_updated_at"`
}

type Setting struct {
	Id    int    `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Key   string `json:"key" form:"key"`
	Value string `json:"value" form:"value"`
}

type AccessIPRecord struct {
	Id        int    `json:"id" gorm:"primaryKey;autoIncrement"`
	SourceIP  string `json:"sourceIp" gorm:"uniqueIndex:idx_access_ip_port;size:64"`
	LastPort  int    `json:"lastPort" gorm:"uniqueIndex:idx_access_ip_port"`
	HitCount  int64  `json:"hitCount"`
	FirstSeen int64  `json:"firstSeen"`
	LastSeen  int64  `json:"lastSeen" gorm:"index"`
}
