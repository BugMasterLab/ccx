package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BenedictKing/ccx/internal/config"
	"github.com/BenedictKing/ccx/internal/metrics"
	"github.com/BenedictKing/ccx/internal/scheduler"
	"github.com/BenedictKing/ccx/internal/session"
	"github.com/BenedictKing/ccx/internal/warmup"
	"github.com/gin-gonic/gin"
)

// TestGetChannelDashboard_DisabledKeysVisibleForBothServiceTypes 验证：
// Messages 标签下的渠道无论 serviceType=claude 还是 serviceType=responses，
// dashboard 都应返回 disabledApiKeys 字段。复现 Bug 1 的场景。
func TestGetChannelDashboard_DisabledKeysVisibleForBothServiceTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		Upstream: []config.UpstreamConfig{
			{
				Name:        "claude-channel",
				ServiceType: "claude",
				BaseURL:     "https://example.com",
				APIKeys:     []string{"sk-claude-active"},
				DisabledAPIKeys: []config.DisabledKeyInfo{{
					Key:        "sk-claude-blocked",
					Reason:     "authentication_error",
					Message:    "invalid",
					DisabledAt: time.Now().Format(time.RFC3339),
				}},
			},
			{
				Name:        "responses-channel",
				ServiceType: "responses",
				BaseURL:     "https://example.com",
				APIKeys:     []string{"sk-responses-active"},
				DisabledAPIKeys: []config.DisabledKeyInfo{{
					Key:        "sk-responses-blocked",
					Reason:     "authentication_error",
					Message:    "invalid",
					DisabledAt: time.Now().Format(time.RFC3339),
				}},
			},
		},
	}

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("序列化配置失败: %v", err)
	}
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	cfgManager, err := config.NewConfigManager(configFile)
	if err != nil {
		t.Fatalf("创建配置管理器失败: %v", err)
	}
	defer cfgManager.Close()

	messagesMetrics := metrics.NewMetricsManager()
	responsesMetrics := metrics.NewMetricsManager()
	geminiMetrics := metrics.NewMetricsManager()
	chatMetrics := metrics.NewMetricsManager()
	imagesMetrics := metrics.NewMetricsManager()
	defer messagesMetrics.Stop()
	defer responsesMetrics.Stop()
	defer geminiMetrics.Stop()
	defer chatMetrics.Stop()
	defer imagesMetrics.Stop()

	traceAffinity := session.NewTraceAffinityManager()
	defer traceAffinity.Stop()
	urlManager := warmup.NewURLManager(30*time.Second, 3)
	sch := scheduler.NewChannelScheduler(cfgManager, messagesMetrics, responsesMetrics, geminiMetrics, chatMetrics, imagesMetrics, traceAffinity, urlManager)

	r := gin.New()
	r.GET("/messages/channels/dashboard", GetChannelDashboard(cfgManager, sch))

	req := httptest.NewRequest(http.MethodGet, "/messages/channels/dashboard?type=messages", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Channels []map[string]any `json:"channels"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if len(resp.Channels) != 2 {
		t.Fatalf("channels len=%d, want=2", len(resp.Channels))
	}

	for _, ch := range resp.Channels {
		name, _ := ch["name"].(string)
		st, _ := ch["serviceType"].(string)
		dk, ok := ch["disabledApiKeys"]
		if !ok || dk == nil {
			t.Fatalf("channel %s (serviceType=%s) 缺少 disabledApiKeys 字段: %+v", name, st, ch)
		}
		dkSlice, ok := dk.([]any)
		if !ok || len(dkSlice) != 1 {
			t.Fatalf("channel %s (serviceType=%s) disabledApiKeys 应有 1 项，实际=%v", name, st, dk)
		}
		// dashboard 也应返回 cooldownApiKeys 字段（即使为空数组），保证前端能渲染冷却 chip
		if _, ok := ch["cooldownApiKeys"]; !ok {
			t.Fatalf("channel %s (serviceType=%s) 缺少 cooldownApiKeys 字段: %+v", name, st, ch)
		}
		t.Logf("channel %s (serviceType=%s) disabledApiKeys=%v cooldownApiKeys=%v ✓", name, st, dkSlice, ch["cooldownApiKeys"])
	}
}
