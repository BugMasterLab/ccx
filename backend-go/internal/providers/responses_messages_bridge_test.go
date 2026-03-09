package providers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/BenedictKing/ccx/internal/config"
	"github.com/BenedictKing/ccx/internal/types"
	"github.com/gin-gonic/gin"
)

func TestResponsesProvider_BuildResponsesRequestFromClaude(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &ResponsesProvider{}
	upstream := &config.UpstreamConfig{
		ServiceType: "responses",
		ModelMapping: map[string]string{
			"gpt-5": "gpt-5.2",
		},
	}

	body := []byte(`{
		"model":"gpt-5",
		"system":"you are helpful",
		"max_tokens":1024,
		"temperature":0.2,
		"stream":true,
		"messages":[
			{"role":"user","content":[{"type":"text","text":"hello"}]},
			{"role":"assistant","content":[{"type":"tool_use","id":"call_1","name":"weather","input":{"city":"shanghai"}}]},
			{"role":"user","content":[{"type":"tool_result","tool_use_id":"call_1","content":"sunny"}]}
		],
		"tools":[{"name":"weather","description":"weather tool","input_schema":{"type":"object"}}]
	}`)

	result, err := provider.buildResponsesRequestFromClaude(body, upstream)
	if err != nil {
		t.Fatalf("buildResponsesRequestFromClaude() err = %v", err)
	}

	if result["model"] != "gpt-5.2" {
		t.Fatalf("model = %v, want gpt-5.2", result["model"])
	}
	if result["instructions"] != "you are helpful" {
		t.Fatalf("instructions = %v, want you are helpful", result["instructions"])
	}
	if result["max_output_tokens"] != float64(1024) && result["max_output_tokens"] != 1024 {
		t.Fatalf("max_output_tokens = %v, want 1024", result["max_output_tokens"])
	}
	if result["stream"] != true {
		t.Fatalf("stream = %v, want true", result["stream"])
	}

	input, ok := result["input"].([]map[string]interface{})
	if !ok {
		// marshal/unmarshal fallback for interface dynamic shape
		b, _ := json.Marshal(result["input"])
		var tmp []map[string]interface{}
		if err := json.Unmarshal(b, &tmp); err != nil {
			t.Fatalf("input decode err: %v", err)
		}
		input = tmp
	}

	if len(input) != 3 {
		t.Fatalf("len(input) = %d, want 3", len(input))
	}
	if input[0]["type"] != "message" {
		t.Fatalf("input[0].type = %v, want message", input[0]["type"])
	}
	if input[1]["type"] != "function_call" {
		t.Fatalf("input[1].type = %v, want function_call", input[1]["type"])
	}
	if input[2]["type"] != "function_call_output" {
		t.Fatalf("input[2].type = %v, want function_call_output", input[2]["type"])
	}

	tools, ok := result["tools"].([]map[string]interface{})
	if !ok {
		b, _ := json.Marshal(result["tools"])
		var tmp []map[string]interface{}
		if err := json.Unmarshal(b, &tmp); err != nil {
			t.Fatalf("tools decode err: %v", err)
		}
		tools = tmp
	}
	if len(tools) != 1 || tools[0]["name"] != "weather" {
		t.Fatalf("tools = %#v, want weather tool", tools)
	}
	// 验证 type 字段必须存在且为 "function"
	if tools[0]["type"] != "function" {
		t.Fatalf("tools[0][\"type\"] = %v, want \"function\"", tools[0]["type"])
	}
	// 验证 parameters 字段必须存在
	if tools[0]["parameters"] == nil {
		t.Fatalf("tools[0][\"parameters\"] is nil, want non-nil")
	}
}

func TestResponsesProvider_ConvertToClaudeResponse(t *testing.T) {
	provider := &ResponsesProvider{}
	providerResp := &types.ProviderResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body: []byte(`{
			"id":"resp_123",
			"status":"completed",
			"output":[
				{"type":"message","content":[{"type":"output_text","text":"hello world"}]},
				{"type":"function_call","call_id":"call_1","name":"weather","arguments":"{\"city\":\"shanghai\"}"}
			],
			"usage":{"input_tokens":12,"output_tokens":34}
		}`),
	}

	claudeResp, err := provider.ConvertToClaudeResponse(providerResp)
	if err != nil {
		t.Fatalf("ConvertToClaudeResponse() err = %v", err)
	}
	if claudeResp.ID != "resp_123" {
		t.Fatalf("ID = %s, want resp_123", claudeResp.ID)
	}
	if claudeResp.StopReason != "tool_use" {
		t.Fatalf("StopReason = %s, want tool_use", claudeResp.StopReason)
	}
	if len(claudeResp.Content) != 2 {
		t.Fatalf("len(Content) = %d, want 2", len(claudeResp.Content))
	}
	if claudeResp.Content[0].Type != "text" || claudeResp.Content[0].Text != "hello world" {
		t.Fatalf("content[0] = %#v, want text hello world", claudeResp.Content[0])
	}
	if claudeResp.Content[1].Type != "tool_use" || claudeResp.Content[1].Name != "weather" {
		t.Fatalf("content[1] = %#v, want tool_use weather", claudeResp.Content[1])
	}
	if claudeResp.Usage == nil || claudeResp.Usage.InputTokens != 12 || claudeResp.Usage.OutputTokens != 34 {
		t.Fatalf("usage = %#v, want input=12 output=34", claudeResp.Usage)
	}
}
