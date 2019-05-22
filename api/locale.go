package api

import (
	"git.stuhome.com/Sunmxt/wing/common"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
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

func ListLoginLocaleText(ctx *gin.Context) {
	rctx := NewRequestContext(ctx)
	text := &LoginLocaleText{
		Login:                    rctx.TranslateMessage("UI.Login.Login"),
		Register:                 rctx.TranslateMessage("UI.Login.Register"),
		Account:                  rctx.TranslateMessage("UI.Login.Account"),
		Password:                 rctx.TranslateMessage("UI.Login.Password"),
		UsernamePrompt:           rctx.TranslateMessage("UI.Login.Prompt.Username"),
		PasswordPrompt:           rctx.TranslateMessage("UI.Login.Prompt.Password"),
		PasswordConfrim:          rctx.TranslateMessage("UI.Login.PasswordConfrim"),
		PasswordConfrimPrompt:    rctx.TranslateMessage("UI.Login.Prompt.PasswordConfrim"),
		UsernameMissing:          rctx.TranslateMessage("UI.Login.UsernameMissing"),
		PasswordMissing:          rctx.TranslateMessage("UI.Login.PasswordMissing"),
		PasswordConfrimMissing:   rctx.TranslateMessage("UI.Login.PasswordConfrimMissing"),
		PasswordConfrimUnmatched: rctx.TranslateMessage("UI.Login.PasswordConfrimUnmatched"),
	}
	if rctx.GetLocaleLanguage() != language.English {
		text.AccountSub = common.TranslateMessage("en", "UI.Login.Account")
		text.PasswordSub = common.TranslateMessage("en", "UI.Login.Password")
		text.PasswordConfrimSub = common.TranslateMessage("en", "UI.Login.PasswordConfrim")
	}
	rctx.Response.Data = text
	rctx.Succeed()
}

func ListDashboardLocaleText(ctx *gin.Context) {

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
