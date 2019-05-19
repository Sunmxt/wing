package common

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func init() {
	message.SetString(language.Chinese, "Login.InvalidAccount", "用户名或密码无效")
	message.SetString(language.English, "Login.InvalidAccount", "invalid username or password")

	message.SetString(language.Chinese, "Auth.Unauthenticated", "身份未认证")
	message.SetString(language.English, "Auth.Unauthenticated", "unauthenticated")

	message.SetString(language.Chinese, "Succeed", "成功")
	message.SetString(language.English, "Succeed", "succeed.")

	// Dashboard tags
	message.SetString(language.Chinese, "UI.Tag.Overview", "概览")
	message.SetString(language.Chinese, "UI.Tag.Orchestration", "应用编排")
	message.SetString(language.Chinese, "UI.Tag.Management", "管理")
	message.SetString(language.Chinese, "UI.Tag.LoadBalance", "负载均衡")
	message.SetString(language.English, "UI.Tag.Overview", "Overview")
	message.SetString(language.English, "UI.Tag.Orchestration", "Orchestration")
	message.SetString(language.English, "UI.Tag.Management", "Management")
	message.SetString(language.English, "UI.Tag.LoadBalance", "LoadBalance")

	// Login page
	message.SetString(language.Chinese, "UI.Login.Login", "登录")
	message.SetString(language.Chinese, "UI.Login.Register", "注册")
	message.SetString(language.Chinese, "UI.Login.Account", "账户")
	message.SetString(language.Chinese, "UI.Login.Password", "密码")
	message.SetString(language.Chinese, "UI.Login.Prompt.Username", "用户名")
	message.SetString(language.Chinese, "UI.Login.Prompt.Password", "密码")
	message.SetString(language.Chinese, "UI.Login.Prompt.PasswordConfrim", "重复输入以确认密码")
	message.SetString(language.Chinese, "UI.Login.PasswordConfrim", "确认密码")

	message.SetString(language.English, "UI.Login.Login", "Login")
	message.SetString(language.English, "UI.Login.Register", "Register")
	message.SetString(language.English, "UI.Login.Account", "Account")
	message.SetString(language.English, "UI.Login.Password", "Password")
	message.SetString(language.English, "UI.Login.Prompt.Username", "Username")
	message.SetString(language.English, "UI.Login.Prompt.Password", "Password")
	message.SetString(language.English, "UI.Login.Prompt.PasswordConfrim", "Confrim password")
	message.SetString(language.English, "UI.Login.PasswordConfrim", "Password confrim")
}

func TranslateMessage(lang, key string, args ...interface{}) string {
	tag := message.MatchLanguage(lang)
	if tag == language.Und {
		return key
	}
	p := message.NewPrinter(tag)
	return p.Sprintf(key, args...)
}
