package config

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGetAdminAPIKeyPrefersActiveKey(t *testing.T) {
	cm := &ConfigManager{}
	upstream := &UpstreamConfig{
		Name:    "test-channel",
		APIKeys: []string{"sk-active"},
		DisabledAPIKeys: []DisabledKeyInfo{{
			Key: "sk-disabled",
		}},
	}

	got, fallback, err := cm.GetAdminAPIKey(upstream, nil, "Messages")
	if err != nil {
		t.Fatalf("GetAdminAPIKey() error = %v", err)
	}
	if fallback {
		t.Fatal("fallback = true, want false")
	}
	if got != "sk-active" {
		t.Fatalf("apiKey = %q, want sk-active", got)
	}
}

func TestGetAdminAPIKeyDoesNotFallbackToDisabledKey(t *testing.T) {
	cm := &ConfigManager{}
	upstream := &UpstreamConfig{
		Name:    "test-channel",
		APIKeys: nil,
		DisabledAPIKeys: []DisabledKeyInfo{{
			Key: "sk-disabled",
		}},
	}

	got, fallback, err := cm.GetAdminAPIKey(upstream, nil, "Messages")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if fallback {
		t.Fatal("fallback = true, want false")
	}
	if got != "" {
		t.Fatalf("apiKey = %q, want empty", got)
	}
}

func TestGetAdminAPIKeySkipsDisabledKeyEvenIfStillActive(t *testing.T) {
	cm := &ConfigManager{}
	upstream := &UpstreamConfig{
		Name:    "test-channel",
		APIKeys: []string{"sk-disabled", "sk-active"},
		DisabledAPIKeys: []DisabledKeyInfo{{
			Key: "sk-disabled",
		}},
	}

	got, fallback, err := cm.GetAdminAPIKey(upstream, nil, "Messages")
	if err != nil {
		t.Fatalf("GetAdminAPIKey() error = %v", err)
	}
	if fallback {
		t.Fatal("fallback = true, want false")
	}
	if got != "sk-active" {
		t.Fatalf("apiKey = %q, want sk-active", got)
	}
}

func TestGetAdminAPIKeyReturnsErrorWhenNoKeysAvailable(t *testing.T) {
	cm := &ConfigManager{}
	upstream := &UpstreamConfig{Name: "test-channel"}

	_, _, err := cm.GetAdminAPIKey(upstream, nil, "Messages")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidateAdminProbeKeyRejectsDisabledAndCooldownKeys(t *testing.T) {
	cm := newConfigManagerForTestJSON(t, `{
		"upstream": [{
			"name": "test-channel",
			"baseUrl": "https://example.com",
			"apiKeys": ["sk-cooldown", "sk-active"],
			"disabledApiKeys": [{"key": "sk-disabled"}],
			"serviceType": "claude"
		}]
	}`)
	cm.MarkKeyAsFailedWithDuration("sk-cooldown", "Messages", time.Hour)

	if err := cm.ValidateAdminProbeKey("Messages", 0, "sk-disabled"); err == nil {
		t.Fatal("disabled key should be rejected")
	}
	if err := cm.ValidateAdminProbeKey("Messages", 0, "sk-cooldown"); err == nil {
		t.Fatal("cooldown key should be rejected")
	}
	if err := cm.ValidateAdminProbeKey("Messages", 0, "sk-active"); err != nil {
		t.Fatalf("active key should be accepted: %v", err)
	}
	if err := cm.ValidateAdminProbeKey("Messages", 0, "sk-new"); err != nil {
		t.Fatalf("unknown temporary key should be accepted: %v", err)
	}
}

func TestGetUsableAPIKeyForChannelSkipsDisabledAndCooldownKeys(t *testing.T) {
	cm := newConfigManagerForTestJSON(t, `{
		"upstream": [{
			"name": "test-channel",
			"baseUrl": "https://example.com",
			"apiKeys": ["sk-disabled", "sk-cooldown", "sk-active"],
			"disabledApiKeys": [{"key": "sk-disabled"}],
			"serviceType": "claude"
		}]
	}`)
	cm.MarkKeyAsFailedWithDuration("sk-cooldown", "Messages", time.Hour)

	got, err := cm.GetUsableAPIKeyForChannel("Messages", 0)
	if err != nil {
		t.Fatalf("GetUsableAPIKeyForChannel() error = %v", err)
	}
	if got != "sk-active" {
		t.Fatalf("apiKey = %q, want sk-active", got)
	}
}

func TestGetNextAPIKeySkipsSingleCooldownKey(t *testing.T) {
	cm := &ConfigManager{
		failedKeysCache: map[string]*FailedKey{
			failedKeyCacheKey("Messages", "sk-only"): {
				Timestamp:    time.Now(),
				FailureCount: 1,
			},
		},
		keyBackoffDurations: []time.Duration{time.Minute},
	}
	upstream := &UpstreamConfig{
		Name:    "test-channel",
		APIKeys: []string{"sk-only"},
	}

	got, err := cm.GetNextAPIKey(upstream, nil, "Messages")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got != "" {
		t.Fatalf("apiKey = %q, want empty", got)
	}
}

func TestGetNextAPIKeyDoesNotReuseOldestCooldownKey(t *testing.T) {
	now := time.Now()
	cm := &ConfigManager{
		failedKeysCache: map[string]*FailedKey{
			failedKeyCacheKey("Messages", "sk-old"): {
				Timestamp:    now.Add(-30 * time.Second),
				FailureCount: 1,
			},
			failedKeyCacheKey("Messages", "sk-new"): {
				Timestamp:    now,
				FailureCount: 1,
			},
		},
		keyBackoffDurations: []time.Duration{time.Minute},
	}
	upstream := &UpstreamConfig{
		Name:    "test-channel",
		APIKeys: []string{"sk-old", "sk-new"},
	}

	got, err := cm.GetNextAPIKey(upstream, nil, "Messages")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got != "" {
		t.Fatalf("apiKey = %q, want empty", got)
	}
}

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
	defer func() { _ = cm.Close() }()

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

func TestMarkKeyAsFailedCoolingWindowAndRecoveryLog(t *testing.T) {
	cm := &ConfigManager{
		failedKeysCache: make(map[string]*FailedKey),
		keyRecoveryTime: 50 * time.Millisecond,
		maxFailureCount: 2,
	}

	cm.MarkKeyAsFailed("sk-test", "Messages")
	if !cm.IsKeyFailed("sk-test", "Messages") {
		t.Fatal("IsKeyFailed() = false, want true immediately after failure")
	}

	cacheKey := failedKeyCacheKey("Messages", "sk-test")
	cm.mu.Lock()
	cm.failedKeysCache[cacheKey].Timestamp = time.Now().Add(-100 * time.Millisecond)
	cm.mu.Unlock()

	if cm.IsKeyFailed("sk-test", "Messages") {
		t.Fatal("IsKeyFailed() = true, want false after recovery window elapsed")
	}

	var buf bytes.Buffer
	origWriter := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(origWriter)

	cm.mu.Lock()
	cm.failedKeysCache[cacheKey] = &FailedKey{Timestamp: time.Now().Add(-100 * time.Millisecond), FailureCount: 1}
	cm.mu.Unlock()

	cm.mu.Lock()
	now := time.Now()
	for key, failure := range cm.failedKeysCache {
		recoveryTime := cm.keyRecoveryTime
		if failure.FailureCount > cm.maxFailureCount {
			recoveryTime = cm.keyRecoveryTime * 2
		}
		if now.Sub(failure.Timestamp) > recoveryTime {
			delete(cm.failedKeysCache, key)
			log.Printf("[Config-Key] API密钥 %s 已从失败列表中恢复", key)
		}
	}
	cm.mu.Unlock()

	if _, exists := cm.failedKeysCache[cacheKey]; exists {
		t.Fatal("failed key cache entry still exists after simulated cleanup")
	}
	if !strings.Contains(buf.String(), "已从失败列表中恢复") {
		t.Fatalf("expected recovery log, got %q", buf.String())
	}
}

func TestBlacklistAndRestoreLogsIncludeTransitionFields(t *testing.T) {
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

	var buf bytes.Buffer
	origWriter := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(origWriter)

	if err := cm.BlacklistKey("Messages", 0, "sk-active", "insufficient_balance", "no balance"); err != nil {
		t.Fatalf("BlacklistKey() error = %v", err)
	}
	if err := cm.RestoreKey("Messages", 0, "sk-active"); err != nil {
		t.Fatalf("RestoreKey() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "from=active") || !strings.Contains(output, "to=disabled") || !strings.Contains(output, "cause=insufficient_balance") {
		t.Fatalf("blacklist transition fields missing: %q", output)
	}
	if !strings.Contains(output, "from=disabled") || !strings.Contains(output, "to=active") || !strings.Contains(output, "cause=manual_restore") {
		t.Fatalf("restore transition fields missing: %q", output)
	}
}

func TestValidateChannelKeysSuspendsChatChannelWithoutKeys(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	initialConfig := `{
		"chatUpstream": [{
			"name": "chat-channel",
			"baseUrl": "https://example.com",
			"apiKeys": [],
			"status": "active",
			"serviceType": "openai"
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

	cfg := cm.GetConfig()
	if len(cfg.ChatUpstream) != 1 {
		t.Fatalf("len(ChatUpstream) = %d, want 1", len(cfg.ChatUpstream))
	}
	if got := cfg.ChatUpstream[0].Status; got != "suspended" {
		t.Fatalf("Chat status = %s, want suspended", got)
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
	defer func() { _ = cm.Close() }()

	disabled := false
	if _, err := cm.UpdateUpstream(0, UpstreamUpdate{AutoBlacklistBalance: &disabled}); err != nil {
		t.Fatalf("UpdateUpstream() error = %v", err)
	}

	cfg := cm.GetConfig()
	if cfg.Upstream[0].AutoBlacklistBalance == nil || *cfg.Upstream[0].AutoBlacklistBalance != false {
		t.Fatalf("AutoBlacklistBalance = %v, want false", cfg.Upstream[0].AutoBlacklistBalance)
	}
}
