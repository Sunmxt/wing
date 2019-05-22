package api

import (
	"git.stuhome.com/Sunmxt/wing/common"
	"github.com/gin-gonic/gin"
)

func RegisterAPI(engine *gin.Engine) error {
	engine.POST("/api/login", AuthLoginV1)
	engine.GET("/api/dashboard/tags", ListDashboardTags)

	// Locale API
	engine.GET("/api/locale/login/list", ListLoginLocaleText)
	engine.GET("/api/locale/dashboard/list", ListDashboardLocaleText)
	engine.GET("/api/locale", GetCurrentLocale)
	engine.POST("/api/locale", SetLocale)

	// Application API
	engine.GET("/api/application", ListApplication)
	engine.GET("/api/application/:name/info", GetApplicationInfo)
	engine.POST("/api/application", CreateApplication)
	engine.GET("/api/application/:name/deploy", ListDeployment)
	engine.POST("/api/application/:name/deploy", CreateDeployment)
	engine.GET("/api/application/:name/deploy/:deploy_id/info", GetDeploymentInfo)
	engine.GET("/api/application/:name/deploy/:deploy_id/start", StartDeployment)
	engine.GET("/api/application/:name/deploy/:deploy_id/stop", StopDeployment)
	engine.GET("/api/application/:name/deploy/:deploy_id/rollback", RollbackDeployment)
	engine.GET("/api/application/:name/roll", RollDeployment)

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
