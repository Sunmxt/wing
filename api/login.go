package api

import (
	"net/http"

	"git.stuhome.com/Sunmxt/wing/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type UserInfoResponse struct {
	Identify  string `json:"id"`
	Name      string `json:"name"`
	IsLogined bool   `json:"login"`
}

func AuthUserInfoV1(ctx *gin.Context) {
	rctx, resp := NewRequestContext(ctx), &UserInfoResponse{}
	rctx.Response.Data = resp
	defer rctx.Succeed()
	if !rctx.LoginEnsured(false) {
		resp.IsLogined = false
		return
	} else {
		resp.IsLogined = true
	}
	resp.Name = rctx.OpCtx.Account.Name
	resp.Identify = rctx.OpCtx.Account.Name
}

type LoginRequestForm struct {
	User     string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func AuthLoginV1(ctx *gin.Context) {
	rctx, req := NewRequestContext(ctx), LoginRequestForm{}
	if rctx.OpCtx.Account.Name != "" {
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
	hasher, err := model.NewSecretHasher(config.SessionToken)
	if err != nil {
		rctx.AbortWithDebugMessage(http.StatusInternalServerError, "invalid session token.")
		return
	}
	account := &model.Account{}
	if err = db.Where(&model.Account{Name: req.User}).First(account).Error; err != nil {
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
	rctx.OpCtx.Account.Name = req.User
	rctx.Session.Set("user", rctx.OpCtx.Account.Name)
	rctx.Session.Save()
	rctx.SucceedWithMessage("Succeed")
}
