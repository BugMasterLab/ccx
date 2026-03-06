package chat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/BenedictKing/ccx/internal/config"
	"github.com/gin-gonic/gin"
)

func setupChatConfigManager(t *testing.T, upstream []config.UpstreamConfig) *config.ConfigManager {
	t.Helper()
	cfg := config.Config{ChatUpstream: upstream}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("序列化配置失败: %v", err)
	}
	tmpFile := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}
	cm, err := config.NewConfigManager(tmpFile)
	if err != nil {
		t.Fatalf("创建配置管理器失败: %v", err)
	}
	t.Cleanup(func() { cm.Close() })
	return cm
}

func TestGetUpstreams_IncludesAdvancedOptionFields(t *testing.T) {
	cm := setupChatConfigManager(t, []config.UpstreamConfig{{
		Name:             "chat-ch",
		ServiceType:      "openai",
		BaseURL:          "https://api.example.com",
		APIKeys:          []string{"sk-1"},
		ModelMapping:     map[string]string{"gpt-5": "gpt-5.2"},
		ReasoningMapping: map[string]string{"gpt-5": "xhigh"},
		TextVerbosity:    "low",
		FastMode:         true,
	}})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/chat/channels", GetUpstreams(cm))

	req := httptest.NewRequest(http.MethodGet, "/chat/channels", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp struct {
		Channels []map[string]interface{} `json:"channels"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	ch := resp.Channels[0]
	if ch["textVerbosity"] != "low" {
		t.Fatalf("textVerbosity = %v, want low", ch["textVerbosity"])
	}
	if ch["fastMode"] != true {
		t.Fatalf("fastMode = %v, want true", ch["fastMode"])
	}
	rm, ok := ch["reasoningMapping"].(map[string]interface{})
	if !ok || rm["gpt-5"] != "xhigh" {
		t.Fatalf("reasoningMapping = %#v, want gpt-5=xhigh", ch["reasoningMapping"])
	}
}
