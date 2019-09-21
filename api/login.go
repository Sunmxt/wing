package api

import (
	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/controller"
	"git.stuhome.com/Sunmxt/wing/model/account"
	"github.com/gin-gonic/gin"
)

type UserInfoResponse struct {
	Identify  string `json:"id"`
	Name      string `json:"name"`
	IsLogined bool   `json:"login"`
}

func AuthUserInfoV1(ctx *gin.Context) {
	rctx, resp := acommon.NewRequestContext(ctx), &UserInfoResponse{}
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

type CredentialRequestForm struct {
	User     string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func AuthLoginV1(ctx *gin.Context) {
	rctx, req := acommon.NewRequestContext(ctx), CredentialRequestForm{}
	if rctx.OpCtx.Account.Name != "" {
		rctx.SucceedWithMessage("Succeed")
		return
	}
	if err := ctx.ShouldBind(&req); err != nil {
		rctx.FailWithMessage("invalid parameters: " + err.Error())
		return
	}

	var user *account.Account
	var err, ldapAuthError error
	if rctx.OpCtx.Runtime.Config.Auth.EnableLDAP {
		if user, ldapAuthError = controller.AuthAsLDAPUser(&rctx.OpCtx, req.User, req.Password); err != nil {
			switch ldapAuthError {
			case common.ErrInvalidPassword:
				rctx.FailWithMessage("Login.InvalidAccount")
				return
			case common.ErrInvalidUsername:
			default:
				rctx.AbortWithError(err)
				return
			}
		}
	}
	if user == nil {
		if user, err = controller.AuthAsLegacyUser(&rctx.OpCtx, req.User, req.Password); err != nil {
			switch err {
			case common.ErrInvalidPassword:
				rctx.FailWithMessage("Login.InvalidAccount")
				return
			case common.ErrInvalidUsername:
			default:
				rctx.AbortWithError(err)
				return
			}
		} //else if rctx.OpCtx.Runtime.Config.Auth.SyncLegacyUser {
		//}
	}
	if user == nil {
		rctx.FailWithMessage("Login.InvalidAccount")
		return
	}

	rctx.OpCtx.Account.Name = req.User
	rctx.Session.Set("user", rctx.OpCtx.Account.Name)
	rctx.Session.Save()
	rctx.SucceedWithMessage("Succeed")
}

func RegisterV1(ctx *gin.Context) {
	rctx, req := acommon.NewRequestContext(ctx), CredentialRequestForm{}
	if err := ctx.ShouldBind(&req); err != nil {
		rctx.FailWithMessage("invalid parameters: " + err.Error())
		return
	}
	config := rctx.ConfigOrFail()
	if config == nil {
		return
	}
	if config.Auth.EnableLDAP {
		if !config.Auth.LDAP.AcceptRegistration {
			rctx.AbortWithError(common.ErrRegisterNotAllowed)
			return
		}

		resp, err := controller.LDAPByName(&rctx.OpCtx, req.User)
		if err != nil {
			rctx.AbortWithError(err)
			return
		}
		if len(resp.Entries) > 0 {
			rctx.AbortWithError(common.ErrAccountExists)
			return
		}
		if err = controller.AddLDAPAccount(&rctx.OpCtx, req.User, req.Password, req.User); err != nil {
			rctx.AbortWithError(err)
			return
		}
	} else {
		if err := controller.AddLegacyAccount(&rctx.OpCtx, req.User, req.Password); err != nil {
			rctx.AbortWithError(err)
			return
		}
	}
	rctx.SucceedWithMessage("Succeed")
}
