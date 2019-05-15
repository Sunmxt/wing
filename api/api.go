package api

import (
	"git.stuhome.com/Sunmxt/wing/uac"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
)

func RegisterAPI(engine *gin.Engine) error {
	engine.GET("/api/login", AuthLoginV1)

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
	if err := ctx.ShouldBind(req); err != nil {
		rctx.FailWithMessage("invalid parameters.")
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
	toVerify, err = hasher.HashString(req.User)
	if toVerify != account.Credentials {
		rctx.FailWithMessage("Login.InvalidAccount")
		return
	}
	rctx.Session.Set("User", rctx.User)
	rctx.Session.Save()
	rctx.SucceedWithMessage("Succeed")
}
