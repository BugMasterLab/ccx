package common

import (
	"github.com/BenedictKing/ccx/internal/config"
	"github.com/gin-gonic/gin"
)

// BuildChannelViewWithCooldown 构建渠道视图（含冷却 key 列表）。
// apiType 用于查询 cfgManager 中的运行时冷却数据，传 "Messages"/"Responses"/"Chat"/"Gemini"/"Images"。
func BuildChannelViewWithCooldown(up config.UpstreamConfig, index int, cfgManager *config.ConfigManager, apiType string) gin.H {
	view := BuildChannelView(up, index)
	if cfgManager != nil && apiType != "" {
		view["cooldownApiKeys"] = cfgManager.GetCooldownKeys(apiType, index)
	}
	return view
}

func BuildChannelView(up config.UpstreamConfig, index int) gin.H {
	status := config.GetChannelStatus(&up)
	priority := config.GetChannelPriority(&up, index)
	return gin.H{
		"index":                            index,
		"name":                             up.Name,
		"serviceType":                      up.ServiceType,
		"baseUrl":                          up.BaseURL,
		"baseUrls":                         up.BaseURLs,
		"apiKeys":                          up.APIKeys,
		"description":                      up.Description,
		"website":                          up.Website,
		"insecureSkipVerify":               up.InsecureSkipVerify,
		"modelMapping":                     up.ModelMapping,
		"reasoningMapping":                 up.ReasoningMapping,
		"textVerbosity":                    up.TextVerbosity,
		"fastMode":                         up.FastMode,
		"latency":                          nil,
		"status":                           status,
		"adminState":                       config.GetChannelAdminState(&up),
		"effectiveState":                   config.GetChannelEffectiveState(&up),
		"runtimeState":                     config.GetChannelRuntimeState(&up),
		"priority":                         priority,
		"promotionUntil":                   up.PromotionUntil,
		"lowQuality":                       up.LowQuality,
		"customHeaders":                    up.CustomHeaders,
		"proxyUrl":                         up.ProxyURL,
		"supportedModels":                  up.SupportedModels,
		"routePrefix":                      up.RoutePrefix,
		"disabledApiKeys":                  up.DisabledAPIKeys,
		"autoBlacklistBalance":             up.IsAutoBlacklistBalanceEnabled(),
		"autoBlacklistEmptyStream":         up.IsAutoBlacklistEmptyStreamEnabled(),
		"normalizeMetadataUserId":          up.IsNormalizeMetadataUserIDEnabled(),
		"stripResponsesUser":               up.IsStripResponsesUserEnabled(),
		"modelsResponseMode":               up.GetModelsResponseMode(),
		"manualModels":                     up.ManualModels,
		"streamPassthroughEnabled":         up.IsStreamPassthroughEnabled(),
		"sub2apiPassthroughEnabled":        up.IsSub2APIPassthroughEnabled(),
		"keyAffinityEnabled":               up.IsKeyAffinityEnabled(),
		"strictRequestPassthroughEnabled":  up.IsStrictRequestPassthroughEnabled(),
		"modelsHealthCheckEnabled":         up.IsModelsHealthCheckEnabled(),
		"modelsHealthCheckIntervalMinutes": up.GetModelsHealthCheckIntervalMinutes(),
		"failoverRules":                    up.GetEffectiveFailoverRules(),
	}
}
