package api

import (
	"git.stuhome.com/Sunmxt/wing/common"
	"github.com/gin-gonic/gin"
)

func RegisterAPI(engine *gin.Engine) error {
	engine.POST("/api/login", AuthLoginV1)
	engine.GET("/api/login", AuthUserInfoV1)
	engine.GET("/api/dashboard/tags", ListDashboardTags)

	// Locale API
	engine.GET("/api/locale", GetCurrentLocale)
	engine.POST("/api/locale", SetLocale)

	// Application API
	engine.POST("/api/application/create", CreateApplication)
	engine.GET("/api/application/list", ListApplication)
	engine.GET("/api/application/info", GetApplicationInfo)

	engine.POST("/api/application/deploy", CreateDeployment)
	engine.POST("/api/application/deploy/start", StartDeployment)
	engine.GET("/api/application/deploy/list", ListDeployment)
	engine.GET("/api/application/deploy/info", GetDeploymentInfo)
	engine.POST("/api/application/deploy/stop", StopDeployment)

	engine.NoRoute(ServeDefault)

	return nil
}

type DashboardTags struct {
	TagsCN []string `json:"cn"`
	TagsEN []string `json:"en"`
}

func ListDashboardTags(ctx *gin.Context) {
	rctx := NewRequestContext(ctx)
	if !rctx.LoginEnsured(true) {
		return
	}
	rctx.Response.Data = DashboardTags{
		TagsCN: []string{
			common.TranslateMessage("zh", "UI.Tag.Overview"),
			common.TranslateMessage("zh", "UI.Tag.Orchestration"),
			common.TranslateMessage("zh", "UI.Tag.LoadBalance"),
			common.TranslateMessage("zh", "UI.Tag.Management"),
		},
		TagsEN: []string{
			common.TranslateMessage("en", "UI.Tag.Overview"),
			common.TranslateMessage("en", "UI.Tag.Orchestration"),
			common.TranslateMessage("en", "UI.Tag.LoadBalance"),
			common.TranslateMessage("en", "UI.Tag.Management"),
		},
	}
	rctx.Succeed()
}
