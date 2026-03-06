package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BenedictKing/ccx/internal/config"
	"github.com/gin-gonic/gin"
)

func TestBuildProviderRequest_InjectsReasoningBeforeModelRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil).WithContext(context.Background())

	bodyBytes := []byte(`{"model":"gpt-5.1-codex","messages":[{"role":"user","content":"hi"}]}`)
	upstream := &config.UpstreamConfig{
		ServiceType: "openai",
		ModelMapping: map[string]string{
			"gpt-5.1-codex": "gpt-5.2-codex",
		},
		ReasoningMapping: map[string]string{
			"gpt-5.1-codex": "xhigh",
		},
		TextVerbosity: "low",
		FastMode:      true,
	}

	req, err := buildProviderRequest(c, upstream, "https://api.example.com", "sk-test", bodyBytes, "gpt-5.1-codex", false)
	if err != nil {
		t.Fatalf("buildProviderRequest() err = %v", err)
	}

	var got map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&got); err != nil {
		t.Fatalf("decode request body: %v", err)
	}

	if got["model"] != "gpt-5.2-codex" {
		t.Fatalf("model = %v, want gpt-5.2-codex", got["model"])
	}

	reasoning, ok := got["reasoning"].(map[string]interface{})
	if !ok || reasoning["effort"] != "xhigh" {
		t.Fatalf("reasoning = %#v, want effort=xhigh", got["reasoning"])
	}

	text, ok := got["text"].(map[string]interface{})
	if !ok || text["verbosity"] != "low" {
		t.Fatalf("text = %#v, want verbosity=low", got["text"])
	}

	if got["service_tier"] != "priority" {
		t.Fatalf("service_tier = %v, want priority", got["service_tier"])
	}
}
