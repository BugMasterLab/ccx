package converters

import (
	"strings"

	"github.com/BenedictKing/ccx/internal/session"
	"github.com/BenedictKing/ccx/internal/types"
)

// ============== Claude Messages API 转换器 ==============

// MapReasoningEffortToThinking 将 OpenAI/Codex 风格的 reasoning effort 字段
// 映射为 Claude Messages API 的 thinking 配置。
//
// 入参：effort 取值（大小写不敏感）
//   - "none"            → 返回 nil（即不附带 thinking 字段，禁用）
//   - "auto"            → enabled, budget=8000（默认中等）
//   - "minimal" / "low" → enabled, budget=1024
//   - "medium"          → enabled, budget=8000
//   - "high"            → enabled, budget=16000
//   - "xhigh"           → enabled, budget=32000
//   - 其它/空            → 返回 nil（不附带 thinking 字段）
//
// 调用方在请求体已存在 thinking 字段时应跳过本函数，避免覆盖客户端显式配置。
func MapReasoningEffortToThinking(effort string) map[string]interface{} {
	switch strings.ToLower(effort) {
	case "none", "":
		return nil
	case "auto", "medium":
		return map[string]interface{}{"type": "enabled", "budget_tokens": 8000}
	case "minimal", "low":
		return map[string]interface{}{"type": "enabled", "budget_tokens": 1024}
	case "high":
		return map[string]interface{}{"type": "enabled", "budget_tokens": 16000}
	case "xhigh":
		return map[string]interface{}{"type": "enabled", "budget_tokens": 32000}
	default:
		return nil
	}
}

// ClaudeConverter 实现 Responses → Claude Messages API 转换
type ClaudeConverter struct{}

// ToProviderRequest 将 Responses 请求转换为 Claude Messages 格式
func (c *ClaudeConverter) ToProviderRequest(sess *session.Session, req *types.ResponsesRequest) (interface{}, error) {
	// 转换 messages 和 system
	messages, system, err := ResponsesToClaudeMessages(sess, req.Input, req.Instructions)
	if err != nil {
		return nil, err
	}

	// 构建 Claude 请求
	claudeReq := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
		"stream":   req.Stream,
	}

	// Claude 使用独立的 system 参数（不在 messages 中）
	if system != "" {
		claudeReq["system"] = system
	}

	// 复制其他参数
	if req.MaxTokens > 0 {
		claudeReq["max_tokens"] = req.MaxTokens
	}
	if req.Temperature > 0 {
		claudeReq["temperature"] = req.Temperature
	}
	if req.TopP > 0 {
		claudeReq["top_p"] = req.TopP
	}
	if req.Stop != nil {
		claudeReq["stop_sequences"] = req.Stop // Claude 使用 stop_sequences
	}
	if len(req.Tools) > 0 {
		if tools := responsesToolsToClaude(req.Tools); len(tools) > 0 {
			claudeReq["tools"] = tools
		}
	}

	// 转换 Responses 请求中的 reasoning.effort → Claude thinking 配置
	// req.Reasoning 已在 types.ResponsesRequest 中预解析为通用 map
	if req.Reasoning != nil {
		if effortRaw, ok := req.Reasoning["effort"]; ok {
			if effort, ok := effortRaw.(string); ok {
				if thinkingCfg := MapReasoningEffortToThinking(effort); thinkingCfg != nil {
					claudeReq["thinking"] = thinkingCfg
				}
			}
		}
	}

	return claudeReq, nil
}

// FromProviderResponse 将 Claude 响应转换为 Responses 格式
func (c *ClaudeConverter) FromProviderResponse(resp map[string]interface{}, sessionID string) (*types.ResponsesResponse, error) {
	return ClaudeResponseToResponses(resp, sessionID)
}

// GetProviderName 获取上游服务名称
func (c *ClaudeConverter) GetProviderName() string {
	return "Claude Messages API"
}
