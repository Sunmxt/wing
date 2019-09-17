package api

import (
	"github.com/gin-gonic/gin"
)

type LoginLocaleText struct {
	Login                    string `json:"login"`
	Register                 string `json:"register"`
	Account                  string `json:"account"`
	Password                 string `json:"password"`
	AccountSub               string `json:"account_subname"`
	PasswordSub              string `json:"password_subname"`
	UsernamePrompt           string `json:"username_prompt"`
	PasswordPrompt           string `json:"password_prompt"`
	PasswordConfrimPrompt    string `json:"password_confrim_prompt"`
	PasswordConfrim          string `json:"password_confrim"`
	PasswordConfrimSub       string `json:"password_confrim_subname"`
	UsernameMissing          string `json:"username_missing"`
	PasswordMissing          string `json:"password_missing"`
	PasswordConfrimMissing   string `json:"password_confrim_missing"`
	PasswordConfrimUnmatched string `json:"password_confrim_unmatched"`
}

type CurrentLocale struct {
	Language string `json:"lang"`
}

func GetCurrentLocale(ctx *gin.Context) {
	rctx := NewRequestContext(ctx)
	rctx.Response.Data = CurrentLocale{
		Language: rctx.Lang,
	}
	rctx.Succeed()
}

func SetLocale(ctx *gin.Context) {
}
