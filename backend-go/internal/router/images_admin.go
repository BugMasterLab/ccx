package router

import (
	"github.com/BenedictKing/ccx/internal/config"
	"github.com/BenedictKing/ccx/internal/handlers"
	"github.com/BenedictKing/ccx/internal/handlers/images"
	"github.com/BenedictKing/ccx/internal/metrics"
	"github.com/BenedictKing/ccx/internal/scheduler"
	"github.com/gin-gonic/gin"
)

func RegisterImagesAdminRoutes(apiGroup *gin.RouterGroup, cfgManager *config.ConfigManager, channelScheduler *scheduler.ChannelScheduler, imagesMetricsManager *metrics.MetricsManager) {
	apiGroup.GET("/images/channels", images.GetUpstreams(cfgManager))
	apiGroup.POST("/images/channels", images.AddUpstream(cfgManager))
	apiGroup.PUT("/images/channels/:id", images.UpdateUpstream(cfgManager, channelScheduler))
	apiGroup.DELETE("/images/channels/:id", images.DeleteUpstream(cfgManager, channelScheduler))
	apiGroup.POST("/images/channels/:id/keys", images.AddApiKey(cfgManager))
	apiGroup.DELETE("/images/channels/:id/keys/:apiKey", images.DeleteApiKey(cfgManager))
	apiGroup.POST("/images/channels/:id/keys/:apiKey/top", images.MoveApiKeyToTop(cfgManager))
	apiGroup.POST("/images/channels/:id/keys/:apiKey/bottom", images.MoveApiKeyToBottom(cfgManager))
	apiGroup.POST("/images/channels/:id/keys/restore", handlers.RestoreBlacklistedKey(cfgManager, "Images"))
	apiGroup.POST("/images/channels/reorder", images.ReorderChannels(cfgManager))
	apiGroup.PATCH("/images/channels/:id/status", images.SetChannelStatus(cfgManager))
	apiGroup.POST("/images/channels/:id/resume", handlers.ResumeChannelWithKind(channelScheduler, cfgManager, scheduler.ChannelKindImages))
	apiGroup.POST("/images/channels/:id/promotion", images.SetChannelPromotion(cfgManager))
	apiGroup.GET("/images/channels/metrics", handlers.GetImagesChannelMetrics(imagesMetricsManager, cfgManager))
	apiGroup.GET("/images/channels/metrics/history", handlers.GetImagesChannelMetricsHistory(imagesMetricsManager, cfgManager))
	apiGroup.GET("/images/channels/:id/keys/metrics/history", handlers.GetImagesChannelKeyMetricsHistory(imagesMetricsManager, cfgManager))
	apiGroup.GET("/images/global/stats/history", handlers.GetGlobalStatsHistory(imagesMetricsManager))
	apiGroup.GET("/images/ping/:id", images.PingChannel(cfgManager))
	apiGroup.GET("/images/ping", images.PingAllChannels(cfgManager))
	apiGroup.POST("/images/channels/:id/models", images.GetChannelModels(cfgManager))
	apiGroup.GET("/images/models/stats/history", handlers.GetModelStatsHistory(imagesMetricsManager))
	apiGroup.GET("/images/channels/:id/logs", handlers.GetChannelLogs(channelScheduler.GetChannelLogStore(scheduler.ChannelKindImages)))
}
