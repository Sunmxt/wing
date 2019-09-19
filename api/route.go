package api

import (
	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	"git.stuhome.com/Sunmxt/wing/api/scm"
	"github.com/gin-gonic/gin"
)

func RegisterAPI(engine *gin.Engine) error {
	engine.POST("/api/login", AuthLoginV1)
	engine.GET("/api/login", AuthUserInfoV1)
	engine.POST("/api/register", RegisterV1)

	engine.GET("/api/settings", WingSettings)

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

	// Source Code Management
	engine.GET("/api/scm/list", scm.ListSCMPlatform)
	engine.GET("/api/scm/repository/list", scm.ListRepository)
	engine.GET("/api/scm/repository/cicd/enable", scm.EnableRepositoryCICD)
	engine.GET("/api/scm/repository/cicd/disable", scm.DisableRepositoryCICD)

	//engine.GET("/api/scm/repository/builds/create", SCMCreateBuild)
	//engine.GET("/api/scm/repository/builds/edit", SCMEditBuild)
	//engine.GET("/api/scm/repository/builds/delete", SCMDeleteBuild)

	engine.NoRoute(ServeDefault)

	return nil
}

type DashboardTags struct {
	TagsCN []string `json:"cn"`
	TagsEN []string `json:"en"`
}

type WingSettingResponse struct {
	AvaliablePanels    []string `json:"avaliable_panels"`
	AcceptRegistration bool     `json:"accept_registration"`
}

func WingSettings(ctx *gin.Context) {
	rctx := acommon.NewRequestContext(ctx)
	config := rctx.ConfigOrFail()
	if config == nil {
		return
	}
	rctx.Response.Data = &WingSettingResponse{
		AvaliablePanels:    []string{"overview", "orchestration"},
		AcceptRegistration: !config.Auth.EnableLDAP || (config.Auth.EnableLDAP && config.Auth.LDAP.AcceptRegistration),
	}
	rctx.Succeed()
}
