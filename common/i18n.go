package common

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func init() {
	// common
	message.SetString(language.Chinese, "Partial.InvalidFields", "无效的字段")
	message.SetString(language.English, "Partial.InvalidFields", "Invalid fields")
	message.SetString(language.Chinese, "Login.InvalidAccount", "用户名或密码无效")
	message.SetString(language.English, "Login.InvalidAccount", "invalid username or password")
	message.SetString(language.Chinese, "Auth.Unauthenticated", "身份未认证")
	message.SetString(language.English, "Auth.Unauthenticated", "unauthenticated")
	message.SetString(language.Chinese, "Auth.LackOfPermission", "权限不足")
	message.SetString(language.English, "Auth.LackOfPermission", "Lack of permission.")

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

	// Orchestration
	message.SetString(language.Chinese, "UI.Operation.ApplicationNotFound", "应用不存在")
	message.SetString(language.English, "UI.Operation.ApplicationNotFound", "Application doesn't exists.")
	message.SetString(language.Chinese, "UI.Operation.ApplicationFound", "应用已存在")
	message.SetString(language.English, "UI.Operation.ApplicationFound", "Application already exists.")
	message.SetString(language.Chinese, "UI.Operation.ExistingDeploymentRunning", "部署正在进行中，不能创建新的部署任务。")
	message.SetString(language.English, "UI.Operation.ExistingDeploymentRunning", "Existing deployment is in process.")
	message.SetString(language.Chinese, "UI.Operation.DeploymentNotFound", "部署任务不存在")
	message.SetString(language.English, "UI.Operation.DeploymentNotFound", "Deployment doesn't exists.")
	message.SetString(language.Chinese, "UI.Operation.Deployment.Start.AlreadyStarted", "部署任务已开始")
	message.SetString(language.English, "UI.Operation.Deployment.Start.AlreadyStarted", "Deployment has already started.")
	message.SetString(language.Chinese, "UI.Operation.Deployment.Start.AlreadyFinished", "部署早已完成")
	message.SetString(language.English, "UI.Operation.Deployment.Start.AlreadyFinished", "Deployment has already finished.")
	message.SetString(language.Chinese, "UI.Operation.Deployment.Start.AlreadyTerminated", "部署已终止，不能重新开始")
	message.SetString(language.English, "UI.Operation.Deployment.Start.AlreadyTerminated", "Deployment has already terminated.")
	message.SetString(language.Chinese, "UI.Operation.Deployment.Start.AlreadyTerminating", "部署正在终止，不能重新开始")
	message.SetString(language.English, "UI.Operation.Deployment.Start.AlreadyTerminating", "Deployment is terminating.")
}

func TranslateMessage(lang, key string, args ...interface{}) string {
	tag := message.MatchLanguage(lang)
	if tag == language.Und {
		return key
	}
	p := message.NewPrinter(tag)
	return p.Sprintf(key, args...)
}
