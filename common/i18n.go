package common

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func registerI18NMessage() {
	// common
	message.SetString(language.Chinese, "SCM.PlatformNotFound", "代码管理平台不存在")
	message.SetString(language.English, "SCM.PlatformNotFound", "SCM Platform not found.")
	message.SetString(language.Chinese, "SCM.PlatformNotSupported", "Wing 暂不支持这个代码管理平台哦")
	message.SetString(language.English, "SCM.PlatformNotSupported", "This SCM Platform is not supported by Wing.")
	message.SetString(language.Chinese, "SCM.RepositoryNotFound", "没有找到指定的代码仓库")
	message.SetString(language.English, "SCM.RepositoryNotFound", "Repository not found.")
	message.SetString(language.Chinese, "SCM.InvalidRepositoryID", "代码仓库ID无效")
	message.SetString(language.English, "SCM.InvalidRepositoryID", "Invalid repository ID.")
	message.SetString(language.Chinese, "SCM.InvalidSCMPlatformID", "代码管理平台ID无效")
	message.SetString(language.English, "SCM.InvalidSCMPlatformID", "Invalid SCM platform ID.")
	message.SetString(language.Chinese, "SCM.InvalidApprovalID", "Approval ID 无效")
	message.SetString(language.English, "SCM.InvalidSCMPlatformID", "Invalid approval ID.")
	message.SetString(language.Chinese, "SCM.CIApprovalCreationFailure", "创建 CI 审核失败.")
	message.SetString(language.English, "SCM.CIApprovalCreationFailure", "Create CI approval failure.")
	message.SetString(language.Chinese, "SCM.RepositoryCIAlreadyEnabled", "代码仓库已开启自动构建")
	message.SetString(language.English, "SCM.RepositoryCIAlreadyEnabled", "CI has already enabled for this repository.")
	message.SetString(language.Chinese, "SCM.BuildNotFound", "构建不存在")
	message.SetString(language.English, "SCM.BuildNotFound", "build not found.")

	message.SetString(language.Chinese, "SAE.SerciceNameAlreadyExists", "服务名已存在")
	message.SetString(language.English, "SAE.SerciceNameAlreadyExists", "Service name already exists.")
	message.SetString(language.Chinese, "SAE.InvalidOrchestratorType", "无效的编排器类型")
	message.SetString(language.English, "SAE.InvalidOrchestratorType", "Invalid orchestrator type.")
	message.SetString(language.Chinese, "SAE.OrchestratorNotFound", "没有找到相应的编排器.")
	message.SetString(language.English, "SAE.OrchestratorNotFound", "Orchestrator not found.")

	message.SetString(language.Chinese, "Partial.InvalidFields", "无效的字段")
	message.SetString(language.English, "Partial.InvalidFields", "Invalid fields")
	message.SetString(language.Chinese, "Login.InvalidAccount", "用户名或密码无效")
	message.SetString(language.English, "Login.InvalidAccount", "invalid username or password")
	message.SetString(language.Chinese, "Auth.Unauthenticated", "身份未认证")
	message.SetString(language.English, "Auth.Unauthenticated", "unauthenticated")
	message.SetString(language.Chinese, "Auth.LackOfPermission", "权限不足")
	message.SetString(language.English, "Auth.LackOfPermission", "Lack of permission.")
	message.SetString(language.Chinese, "Account.Exists", "用户已存在")
	message.SetString(language.English, "Account.Exists", "Account already exists.")
	message.SetString(language.Chinese, "Account.NotAMail", "用户名不是邮箱")
	message.SetString(language.English, "Account.NotAMail", "Username is not a e-mail.")
	message.SetString(language.Chinese, "Account.WeakPassword", "密码强度过低，换一个密码吧。")
	message.SetString(language.English, "Account.WeakPassword", "Weak password detected. Maybe try another one.")
	message.SetString(language.Chinese, "Register.NotAllowed", "目前不允许注册")
	message.SetString(language.English, "Register.NotAllowed", "目前不允许注册")

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

	// StarStudio Application Engine.
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

	// SCM
	message.SetString(language.Chinese, "UI.Flow.Stage.SubmitRepositoryBuildEnableApproval", "提交")
	message.SetString(language.English, "UI.Flow.Stage.SubmitRepositoryBuildEnableApproval", "Submit")
	message.SetString(language.Chinese, "UI.Flow.Stage.SubmitGitlabMergeRequestApproval", "Gitlab 审核")
	message.SetString(language.English, "UI.Flow.Stage.SubmitGitlabMergeRequestApproval", "Gitlab Approval")
	message.SetString(language.English, "UI.Flow.Stage.RepositoryBuildEnabled", "构建开启")
	message.SetString(language.English, "UI.Flow.Stage.RepositoryBuildEnabled", "Build Enabled")
	message.SetString(language.Chinese, "UI.Flow.Stage.Prompt.SubmitRepositoryBuildEnableApproval", "为此代码仓库发起 SCM 平台构建审核流程")
	message.SetString(language.English, "UI.Flow.Stage.Prompt.SubmitRepositoryBuildEnableApproval", "Submit approval workflow for enabling SCM build of the repository.")
	message.SetString(language.Chinese, "UI.Flow.Stage.Prompt.SubmitGitlabMergeRequestApproval", "发起 Gitlab 合并请求以确认代码仓库权限")
	message.SetString(language.English, "UI.Flow.Stage.Prompt.SubmitGitlabMergeRequestApproval", "Submit gitlab merge request to verify repository permission.")
	message.SetString(language.English, "UI.Flow.Stage.Prompt.RepositoryBuildEnabled", "SCM 构建成功开启")
	message.SetString(language.English, "UI.Flow.Stage.Prompt.RepositoryBuildEnabled", "SCM build is enabled for the repository.")
}

func TranslateMessage(lang, key string, args ...interface{}) string {
	tag := message.MatchLanguage(lang)
	if tag == language.Und {
		return key
	}
	p := message.NewPrinter(tag)
	return p.Sprintf(key, args...)
}
