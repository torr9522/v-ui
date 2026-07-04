package service

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"x-ui/database"
	"x-ui/database/model"
)

func initInboundTestDB(t *testing.T) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "inbound-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("init test db failed: %v", err)
	}
}

func newRealityTestInbound(protocol model.Protocol, network string, realitySettings string) *model.Inbound {
	return &model.Inbound{
		Port:     24443,
		Protocol: protocol,
		Settings: `{"clients":[{"id":"11111111-1111-1111-1111-111111111111","flow":"xtls-rprx-vision"}],"decryption":"none"}`,
		StreamSettings: `{
			"network":"` + network + `",
			"security":"reality",
			"realitySettings":` + realitySettings + `,
			"realityShare":{"publicKey":"share-public-key","shortId":"share-short-id","fingerprint":"firefox"}
		}`,
	}
}

func decodeJSONMap(t *testing.T, raw string) map[string]interface{} {
	t.Helper()
	value := make(map[string]interface{})
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		t.Fatalf("unmarshal json failed: %v", err)
	}
	return value
}

func TestAddInboundPersistsNormalizedReality(t *testing.T) {
	initInboundTestDB(t)

	service := InboundService{}
	inbound := newRealityTestInbound(model.VLESS, "tcp", `{
		"target":"target.example.com:443",
		"dest":"ignored.example.com:443",
		"serverNames":"one.example.com\ntwo.example.com",
		"privateKey":"private-key",
		"shortIds":"6ba85179e30d4fc2\n8f2a1b2c3d4e5f60",
		"publicKey":"public-key",
		"shortId":"6ba85179e30d4fc2",
		"fingerprint":"chrome",
		"spiderX":"/news"
	}`)

	if err := service.AddInbound(inbound); err != nil {
		t.Fatalf("add inbound failed: %v", err)
	}

	current, err := service.GetInbound(inbound.Id)
	if err != nil {
		t.Fatalf("get inbound failed: %v", err)
	}

	stream := decodeJSONMap(t, current.StreamSettings)
	if stream["security"] != "reality" {
		t.Fatalf("expected security=reality, got %v", stream["security"])
	}
	if stream["network"] != "tcp" {
		t.Fatalf("expected network=tcp, got %v", stream["network"])
	}

	reality, ok := stream["realitySettings"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected realitySettings object")
	}
	if reality["target"] != "target.example.com:443" {
		t.Fatalf("expected target to win over dest, got %v", reality["target"])
	}
	if _, ok := reality["dest"]; ok {
		t.Fatalf("expected dest to be removed from server realitySettings")
	}
	if reality["privateKey"] != "private-key" {
		t.Fatalf("unexpected privateKey: %v", reality["privateKey"])
	}
	if reality["show"] != false {
		t.Fatalf("expected show=false default, got %v", reality["show"])
	}
	if _, ok := reality["maxTimeDiff"]; ok {
		t.Fatalf("expected empty maxTimeDiff to be omitted")
	}
	if _, ok := reality["publicKey"]; ok {
		t.Fatalf("expected publicKey to be omitted from server realitySettings")
	}
	if _, ok := reality["shortId"]; ok {
		t.Fatalf("expected shortId to be omitted from server realitySettings")
	}
	if _, ok := reality["fingerprint"]; ok {
		t.Fatalf("expected fingerprint to be omitted from server realitySettings")
	}

	realityShare, ok := stream["realityShare"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected realityShare object to be preserved")
	}
	if realityShare["publicKey"] != "share-public-key" {
		t.Fatalf("unexpected realityShare publicKey: %v", realityShare["publicKey"])
	}
	if realityShare["shortId"] != "share-short-id" {
		t.Fatalf("unexpected realityShare shortId: %v", realityShare["shortId"])
	}
	if realityShare["fingerprint"] != "firefox" {
		t.Fatalf("unexpected realityShare fingerprint: %v", realityShare["fingerprint"])
	}

	serverNames, ok := reality["serverNames"].([]interface{})
	if !ok || len(serverNames) != 2 {
		t.Fatalf("expected 2 serverNames, got %v", reality["serverNames"])
	}
	shortIDs, ok := reality["shortIds"].([]interface{})
	if !ok || len(shortIDs) != 2 {
		t.Fatalf("expected 2 shortIds, got %v", reality["shortIds"])
	}
}

func TestNormalizeStreamSettingsMigratesLegacyRealityShareFields(t *testing.T) {
	service := InboundService{}
	inbound := &model.Inbound{
		Port:     24443,
		Protocol: model.VLESS,
		Settings: `{"clients":[{"id":"11111111-1111-1111-1111-111111111111","flow":"xtls-rprx-vision"}],"decryption":"none"}`,
		StreamSettings: `{
			"network":"tcp",
			"security":"reality",
			"realitySettings":{
				"target":"target.example.com:443",
				"serverNames":["one.example.com"],
				"privateKey":"private-key",
				"shortIds":["6ba85179e30d4fc2"],
				"publicKey":"legacy-public-key",
				"shortId":"legacy-short-id",
				"fingerprint":"chrome"
			}
		}`,
	}

	if err := service.normalizeStreamSettings(inbound); err != nil {
		t.Fatalf("normalize streamSettings failed: %v", err)
	}

	stream := decodeJSONMap(t, inbound.StreamSettings)
	realityShare, ok := stream["realityShare"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected migrated realityShare object")
	}
	if realityShare["publicKey"] != "legacy-public-key" {
		t.Fatalf("expected legacy publicKey migration, got %v", realityShare["publicKey"])
	}
	if realityShare["shortId"] != "legacy-short-id" {
		t.Fatalf("expected legacy shortId migration, got %v", realityShare["shortId"])
	}
	if realityShare["fingerprint"] != "chrome" {
		t.Fatalf("expected legacy fingerprint migration, got %v", realityShare["fingerprint"])
	}
}

func TestNormalizeStreamSettingsRejectsRealityOverWS(t *testing.T) {
	service := InboundService{}
	inbound := newRealityTestInbound(model.VLESS, "ws", `{
		"dest":"target.example.com:443",
		"serverNames":["one.example.com"],
		"privateKey":"private-key",
		"shortIds":["6ba85179e30d4fc2"]
	}`)

	if err := service.normalizeStreamSettings(inbound); err == nil {
		t.Fatalf("expected VLESS + WS + REALITY to fail")
	}
}

func TestNormalizeStreamSettingsRejectsRealityOverTrojan(t *testing.T) {
	service := InboundService{}
	inbound := newRealityTestInbound(model.Trojan, "tcp", `{
		"dest":"target.example.com:443",
		"serverNames":["one.example.com"],
		"privateKey":"private-key",
		"shortIds":["6ba85179e30d4fc2"]
	}`)

	if err := service.normalizeStreamSettings(inbound); err == nil {
		t.Fatalf("expected Trojan + REALITY to fail")
	}
}

func TestNormalizeStreamSettingsRejectsMissingPrivateKey(t *testing.T) {
	service := InboundService{}
	inbound := newRealityTestInbound(model.VLESS, "tcp", `{
		"dest":"target.example.com:443",
		"serverNames":["one.example.com"],
		"shortIds":["6ba85179e30d4fc2"]
	}`)

	if err := service.normalizeStreamSettings(inbound); err == nil {
		t.Fatalf("expected missing privateKey to fail")
	}
}

func TestNormalizeStreamSettingsRejectsMissingServerNames(t *testing.T) {
	service := InboundService{}
	inbound := newRealityTestInbound(model.VLESS, "tcp", `{
		"dest":"target.example.com:443",
		"privateKey":"private-key",
		"shortIds":["6ba85179e30d4fc2"]
	}`)

	if err := service.normalizeStreamSettings(inbound); err == nil {
		t.Fatalf("expected missing serverNames to fail")
	}
}

func TestNormalizeStreamSettingsLeavesTLSUntouched(t *testing.T) {
	service := InboundService{}
	inbound := &model.Inbound{
		Protocol: model.VLESS,
		StreamSettings: `{
			"network":"tcp",
			"security":"tls",
			"tlsSettings":{"serverName":"panel.example.com"}
		}`,
	}
	original := inbound.StreamSettings

	if err := service.normalizeStreamSettings(inbound); err != nil {
		t.Fatalf("expected tls streamSettings to pass unchanged: %v", err)
	}
	if inbound.StreamSettings != original {
		t.Fatalf("expected tls streamSettings to remain unchanged")
	}
}
