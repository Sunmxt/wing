package common

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func init() {
	message.SetString(language.Chinese, "Login.InvalidAccount", "用户名或密码无效")
	message.SetString(language.English, "Login.InvalidAccount", "invalid username or password")

	message.SetString(language.Chinese, "Succeed", "成功")
	message.SetString(language.English, "Succeed", "succeed.")
}

func TranslateMessage(lang, key string, args ...interface{}) string {
	tag := message.MatchLanguage(lang)
	if tag == language.Und {
		return key
	}
	p := message.NewPrinter(tag)
	return p.Sprintf(key, args...)
}
