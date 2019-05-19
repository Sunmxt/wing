package api

import (
	"net/http"

	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/uac"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func RegisterAPI(engine *gin.Engine) error {
	engine.GET("/api/login", AuthLoginV1)
	engine.GET("/api/dashboard/tags", ListDashboardTags)
	engine.NoRoute(ServeDefault)

	return nil
}

type LoginRequestForm struct {
	User     string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func AuthLoginV1(ctx *gin.Context) {
	rctx, req := NewRequestContext(ctx), LoginRequestForm{}
	if rctx.User != "" {
		rctx.SucceedWithMessage("Succeed")
		return
	}
	if err := ctx.ShouldBind(&req); err != nil {
		rctx.FailWithMessage("invalid parameters: " + err.Error())
		return
	}
	db, config := rctx.DatabaseOrFail(), rctx.ConfigOrFail()
	if db == nil || config == nil {
		return
	}
	hasher, err := uac.NewSecretHasher(config.SessionToken)
	if err != nil {
		rctx.AbortWithDebugMessage(http.StatusInternalServerError, "invalid session token.")
		return
	}
	account := &uac.Account{}
	if err = db.Where(&uac.Account{Name: req.User}).First(account).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.FailWithMessage("Login.InvalidAccount")
			return
		}
	}

	var toVerify string
	toVerify, err = hasher.HashString(req.Password)
	if toVerify != account.Credentials {
		rctx.FailWithMessage("Login.InvalidAccount")
		return
	}
	rctx.User = req.User
	rctx.Session.Set("user", rctx.User)
	rctx.Session.Save()
	rctx.SucceedWithMessage("Succeed")
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
