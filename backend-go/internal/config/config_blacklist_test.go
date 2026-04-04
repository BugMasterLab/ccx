package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddAPIKeyRemovesDisabledEntryAndRestoreAvoidsDuplicate(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	initialConfig := `{
		"upstream": [{
			"name": "test-channel",
			"baseUrl": "https://example.com",
			"apiKeys": ["sk-active"],
			"disabledApiKeys": [{
				"key": "sk-disabled",
				"reason": "authentication_error",
				"message": "invalid key",
				"disabledAt": "2026-04-04T00:00:00Z"
			}],
			"serviceType": "claude"
		}]
	}`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("写入初始配置失败: %v", err)
	}

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("NewConfigManager() error = %v", err)
	}
	defer cm.Close()

	if err := cm.AddAPIKey(0, "sk-disabled"); err != nil {
		t.Fatalf("AddAPIKey() error = %v", err)
	}

	cfg := cm.GetConfig()
	if len(cfg.Upstream[0].DisabledAPIKeys) != 0 {
		t.Fatalf("DisabledAPIKeys = %+v, want empty after AddAPIKey", cfg.Upstream[0].DisabledAPIKeys)
	}

	cm.mu.Lock()
	cm.config.Upstream[0].DisabledAPIKeys = append(cm.config.Upstream[0].DisabledAPIKeys, DisabledKeyInfo{
		Key:        "sk-disabled",
		Reason:     "authentication_error",
		Message:    "invalid key",
		DisabledAt: "2026-04-04T00:00:00Z",
	})
	cm.mu.Unlock()

	if err := cm.RestoreKey("Messages", 0, "sk-disabled"); err != nil {
		t.Fatalf("RestoreKey() error = %v", err)
	}

	finalCfg := cm.GetConfig()
	count := 0
	for _, key := range finalCfg.Upstream[0].APIKeys {
		if key == "sk-disabled" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("restored key count = %d, want 1; keys=%v", count, finalCfg.Upstream[0].APIKeys)
	}
}

func TestUpdateUpstreamCanSetAutoBlacklistBalance(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	initialConfig := `{
		"upstream": [{
			"name": "test-channel",
			"baseUrl": "https://example.com",
			"apiKeys": ["sk-active"],
			"serviceType": "claude"
		}]
	}`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("写入初始配置失败: %v", err)
	}

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("NewConfigManager() error = %v", err)
	}
	defer cm.Close()

	disabled := false
	if _, err := cm.UpdateUpstream(0, UpstreamUpdate{AutoBlacklistBalance: &disabled}); err != nil {
		t.Fatalf("UpdateUpstream() error = %v", err)
	}

	cfg := cm.GetConfig()
	if cfg.Upstream[0].AutoBlacklistBalance == nil || *cfg.Upstream[0].AutoBlacklistBalance != false {
		t.Fatalf("AutoBlacklistBalance = %v, want false", cfg.Upstream[0].AutoBlacklistBalance)
	}
}
