package service

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"x-ui/database"
	"x-ui/database/model"
)

func initRoutingTestDB(t *testing.T) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "routing-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("init test db failed: %v", err)
	}
}

func TestAddRulePersistsExplicitDisabled(t *testing.T) {
	initRoutingTestDB(t)

	service := RoutingService{}
	rule := &model.RoutingRule{
		Type:        "field",
		Domain:      `["domain:disabled.example"]`,
		OutboundTag: "direct",
		Enabled:     false,
		Remark:      "Disabled Example",
		Sort:        100,
	}

	applied, err := service.AddRule(rule)
	if err != nil {
		t.Fatalf("add rule failed: %v", err)
	}
	if applied {
		t.Fatalf("expected applied=false when managed mode is disabled")
	}

	current, err := service.GetRule(rule.Id)
	if err != nil {
		t.Fatalf("get rule failed: %v", err)
	}
	if current.Enabled {
		t.Fatalf("expected enabled=false after add, got true")
	}
}

func TestUpdateRulePersistsExplicitDisabled(t *testing.T) {
	initRoutingTestDB(t)

	service := RoutingService{}
	rule := &model.RoutingRule{
		Type:        "field",
		Domain:      `["domain:update.example"]`,
		OutboundTag: "direct",
		Enabled:     true,
		Remark:      "Update Example",
		Sort:        110,
	}
	if _, err := service.AddRule(rule); err != nil {
		t.Fatalf("add initial rule failed: %v", err)
	}

	update := &model.RoutingRule{
		Id:          rule.Id,
		Type:        "field",
		Domain:      `["domain:update.example"]`,
		OutboundTag: "direct",
		Enabled:     false,
		Remark:      "Update Example Disabled",
		Sort:        120,
	}
	if _, err := service.UpdateRule(update); err != nil {
		t.Fatalf("update rule failed: %v", err)
	}

	current, err := service.GetRule(rule.Id)
	if err != nil {
		t.Fatalf("get updated rule failed: %v", err)
	}
	if current.Enabled {
		t.Fatalf("expected enabled=false after update, got true")
	}
}

func TestToggleRulePersistsExplicitDisabled(t *testing.T) {
	initRoutingTestDB(t)

	service := RoutingService{}
	rule := &model.RoutingRule{
		Type:        "field",
		Domain:      `["domain:toggle.example"]`,
		OutboundTag: "direct",
		Enabled:     true,
		Remark:      "Toggle Example",
		Sort:        130,
	}
	if _, err := service.AddRule(rule); err != nil {
		t.Fatalf("add initial rule failed: %v", err)
	}

	if _, err := service.ToggleRule(rule.Id, false); err != nil {
		t.Fatalf("toggle rule failed: %v", err)
	}

	current, err := service.GetRule(rule.Id)
	if err != nil {
		t.Fatalf("get toggled rule failed: %v", err)
	}
	if current.Enabled {
		t.Fatalf("expected enabled=false after toggle, got true")
	}
}

func TestManagedRoutingSkipsDisabledRules(t *testing.T) {
	initRoutingTestDB(t)

	service := RoutingService{}
	enabledRule := &model.RoutingRule{
		Type:        "field",
		Domain:      `["domain:enabled.example"]`,
		OutboundTag: "direct",
		Enabled:     true,
		Remark:      "Enabled Example",
		Sort:        100,
	}
	disabledRule := &model.RoutingRule{
		Type:        "field",
		Domain:      `["domain:disabled.example"]`,
		OutboundTag: "direct",
		Enabled:     false,
		Remark:      "Disabled Example",
		Sort:        110,
	}

	if _, err := service.AddRule(enabledRule); err != nil {
		t.Fatalf("add enabled rule failed: %v", err)
	}
	if _, err := service.AddRule(disabledRule); err != nil {
		t.Fatalf("add disabled rule failed: %v", err)
	}

	xrayService := XrayService{}
	raw, err := xrayService.buildManagedRoutingForDB(database.GetDB())
	if err != nil {
		t.Fatalf("build managed routing failed: %v", err)
	}

	cfg := &routingConfig{}
	if err := json.Unmarshal(raw, cfg); err != nil {
		t.Fatalf("unmarshal managed routing failed: %v", err)
	}

	hasEnabled := false
	hasDisabled := false
	for _, rule := range cfg.Rules {
		for _, domain := range rule.Domain {
			if domain == "domain:enabled.example" {
				hasEnabled = true
			}
			if domain == "domain:disabled.example" {
				hasDisabled = true
			}
		}
	}

	if !hasEnabled {
		t.Fatalf("expected enabled rule to appear in managed routing config")
	}
	if hasDisabled {
		t.Fatalf("expected disabled rule to be filtered from managed routing config")
	}
}
