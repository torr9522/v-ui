package model

import (
	"encoding/json"
	"testing"
)

func TestGenXrayInboundConfigStripsRealityShare(t *testing.T) {
	inbound := &Inbound{
		Port:     29443,
		Protocol: VLESS,
		Settings: `{"clients":[{"id":"11111111-1111-1111-1111-111111111111","flow":"xtls-rprx-vision"}],"decryption":"none"}`,
		StreamSettings: `{
			"network":"tcp",
			"security":"reality",
			"realitySettings":{
				"target":"www.cloudflare.com:443",
				"serverNames":["www.cloudflare.com"],
				"privateKey":"private-key",
				"shortIds":["0123456789abcdef"],
				"show":false
			},
			"realityShare":{
				"publicKey":"public-key",
				"shortId":"0123456789abcdef",
				"fingerprint":"chrome"
			}
		}`,
	}

	cfg := inbound.GenXrayInboundConfig()
	stream := make(map[string]interface{})
	if err := json.Unmarshal(cfg.StreamSettings, &stream); err != nil {
		t.Fatalf("unmarshal streamSettings failed: %v", err)
	}

	if _, ok := stream["realityShare"]; ok {
		t.Fatalf("expected realityShare to be stripped from xray config")
	}
	realitySettings, ok := stream["realitySettings"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected realitySettings object")
	}
	if realitySettings["privateKey"] != "private-key" {
		t.Fatalf("unexpected privateKey after strip: %v", realitySettings["privateKey"])
	}
}
