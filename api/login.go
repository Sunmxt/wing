package api

import (
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/model"
	"github.com/gin-gonic/gin"
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

type CredentialRequestForm struct {
	User     string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func AuthLoginV1(ctx *gin.Context) {
	rctx, req := NewRequestContext(ctx), CredentialRequestForm{}
	if rctx.OpCtx.Account.Name != "" {
		rctx.SucceedWithMessage("Succeed")
		return
	}
	if err := ctx.ShouldBind(&req); err != nil {
		rctx.FailWithMessage("invalid parameters: " + err.Error())
		return
	}

	var account *model.Account
	var err, ldapAuthError error
	if rctx.OpCtx.Runtime.Config.Auth.EnableLDAP {
		if account, ldapAuthError = rctx.OpCtx.AuthAsLDAPUser(req.User, req.Password); err != nil {
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
	if account == nil {
		if account, err = rctx.OpCtx.AuthAsLegacyUser(req.User, req.Password); err != nil {
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
	if account == nil {
		rctx.FailWithMessage("Login.InvalidAccount")
		return
	}

	rctx.OpCtx.Account.Name = req.User
	rctx.Session.Set("user", rctx.OpCtx.Account.Name)
	rctx.Session.Save()
	rctx.SucceedWithMessage("Succeed")
}

func RegisterV1(ctx *gin.Context) {
	rctx, req := NewRequestContext(ctx), CredentialRequestForm{}
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

		resp, err := rctx.OpCtx.LDAPByName(req.User)
		if err != nil {
			rctx.AbortWithError(err)
			return
		}
		if len(resp.Entries) > 0 {
			rctx.AbortWithError(common.ErrAccountExists)
			return
		}
		if err = rctx.OpCtx.AddLDAPAccount(req.User, req.Password, req.User); err != nil {
			rctx.AbortWithError(err)
			return
		}
	} else {
		if err := rctx.OpCtx.AddLegacyAccount(req.User, req.Password); err != nil {
			rctx.AbortWithError(err)
			return
		}
	}
	rctx.SucceedWithMessage("Succeed")
}
