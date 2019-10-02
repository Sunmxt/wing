package scm

import (
	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/controller/cicd"
	"git.stuhome.com/Sunmxt/wing/model/scm"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
)

type EnableRepositoryCICDRequest struct {
	PlatformID   uint   `form:"platform_id" binding:"required"`
	RepositoryID string `form:"repository_id" binding:"required"`
}

func (r *EnableRepositoryCICDRequest) Clean(rctx *acommon.RequestContext) error {
	if r.PlatformID < 1 {
		return common.ErrInvalidSCMPlatformID
	}
	if r.RepositoryID == "" {
		return common.ErrInvalidRepositoryID
	}
	return nil
}

type EnableRepositoryCICDResponse struct {
	ApproveID int `form:"approval_id"`
}

func EnableRepositoryCICD(ctx *gin.Context) {
	rctx, request, response := acommon.NewRequestContext(ctx), &EnableRepositoryCICDRequest{}, EnableRepositoryCICDResponse{}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.BindOrFail(request) || !rctx.LoginEnsured(true) {
		return
	}
	account := rctx.GetAccount()
	if account == nil {
		rctx.FailWithMessage("Auth.LackOfPermission")
		return
	}

	platform := &scm.SCMPlatform{}
	err := db.Model(platform).Where("id = ? and active = ?", request.PlatformID, scm.Active).First(platform).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.AbortWithError(common.ErrSCMPlatformNotFound)
			return
		} else {
			rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
			rctx.OpCtx.Log.Error(err.Error())
			return
		}
	}

	var approval *scm.CIRepositoryApproval
	switch platform.Type {
	case scm.GitlabSCM:
		if approval, err = cicd.SubmitGitlabRepositoryCIApproval(&rctx.OpCtx, platform, uint(account.Basic.ID), request.RepositoryID); err != nil {
			rctx.AbortWithError(err)
			return
		}

	default:
		rctx.AbortWithError(common.ErrSCMPlatformNotSupported)
	}
	if approval == nil {
		rctx.FailWithMessage("SCM.CIApprovalCreationFailure")
		return
	}
	response.ApproveID = approval.Basic.ID
	rctx.Response.Data = response
	rctx.Succeed()

}

type DisableRepositoryCICDRequest struct {
	PlatformID   uint   `form:"platform_id" binding:"required"`
	RepositoryID string `form:"repository_id" binding:"required"`
}

func (r *DisableRepositoryCICDRequest) Clean(rctx *acommon.RequestContext) error {
	return nil
}

func DisableRepositoryCICD(ctx *gin.Context) {
	rctx, request := acommon.NewRequestContext(ctx), &DisableRepositoryCICDRequest{}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.BindOrFail(request) {
		return
	}
}

type GetCICDApprovalDetailRequest struct {
	ApprovalID int `form:"approval_id" binding:"required"`
}

type FlowStage struct {
	Name   string      `json:"name"`
	Prompt string      `json:"prompt"`
	State  uint        `json:"status"`
	Extra  interface{} `json:"extra"`
}

const (
	FlowStageWait      = 0
	FlowStageInProcess = 1
	FlowStagePassed    = 2
	FlowStageRejected  = 3
	FlowStageError     = 4
	FlowStageSkip      = 5
)

type GetCICDApprovalDetailResponse struct {
	ID           int         `json:"approval_id"`
	CurrentStage int         `json:"current_stage_index"`
	Stages       []FlowStage `json:"stages"`
}

func (r *GetCICDApprovalDetailRequest) Clean(rctx *acommon.RequestContext) error {
	return nil
}

func GetCICDApprovalDetail(ctx *gin.Context) {
	rctx, request, response := acommon.NewRequestContext(ctx), &GetCICDApprovalDetailRequest{}, GetCICDApprovalDetailResponse{}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.BindOrFail(request) {
		return
	}
	approval := scm.CIRepositoryApproval{}
	err := approval.ByID(db.Preload("SCM").Order("modify_time desc"), request.ApprovalID)
	if err != nil {
		rctx.AbortWithError(err)
		return
	}
	if approval.Basic.ID < 1 {
		rctx.AbortWithError(common.ErrInvalidApprovalID)
		return
	}
	if approval.SCM == nil {
		rctx.AbortWithError(common.ErrSCMPlatformNotFound)
		return
	}
	var repoID int
	switch approval.SCM.Type {
	case scm.GitlabSCM:
		extra := approval.GitlabExtra()
		if extra == nil {
			rctx.AbortWithDebugMessage(http.StatusInternalServerError, "cannot get gitlab scm extra.")
			return
		}
		repoID = int(extra.RepositoryID)

	default:
		rctx.AbortWithError(common.ErrSCMPlatformNotSupported)
		return
	}

	logs, err := scm.GetApprovalStageChangedLogs(db.Order("modify_time desc"), approval.SCMPlatformID, repoID, approval.Basic.ID)
	if err != nil {
		rctx.AbortWithError(err)
		return
	}

	// pick
	response.Stages = make([]FlowStage, 3)
	response.Stages[0].Name = rctx.TranslateMessage("UI.Flow.Stage.SubmitRepositoryBuildEnableApproval")
	response.Stages[1].Name = rctx.TranslateMessage("UI.Flow.Stage.SubmitGitlabMergeRequestApproval")
	response.Stages[2].Name = rctx.TranslateMessage("UI.Flow.Stage.RepositoryBuildEnabled")
	response.Stages[0].Prompt = rctx.TranslateMessage("UI.Flow.Stage.Prompt.SubmitRepositoryBuildEnableApproval")
	response.Stages[1].Prompt = rctx.TranslateMessage("UI.Flow.Stage.Prompt.SubmitGitlabMergeRequestApproval")
	response.Stages[2].Prompt = rctx.TranslateMessage("UI.Flow.Stage.Prompt.RepositoryBuildEnabled")
	response.Stages[0].State = FlowStageWait
	response.Stages[1].State = FlowStageWait
	response.Stages[2].State = FlowStageWait

	// pick workflow status according to latest ci log.
	stageAccepted, extra := true, &scm.CIRepositoryLogApprovalStageChangedExtra{}
	if len(logs) > 0 {
		if err := logs[0].DecodeExtra(extra); err != nil {
			rctx.AbortWithError(err)
			return
		}
		if extra.OldStage < 0 {
			switch extra.NewStage {
			case scm.ApprovalAccepted:
				response.Stages[0].State = FlowStagePassed
				response.Stages[1].State = FlowStageSkip
				response.Stages[2].State = FlowStagePassed
				response.CurrentStage = 3

			case scm.ApprovalRejected:
				response.Stages[0].State = FlowStageRejected
				response.CurrentStage = 1

			case scm.ApprovalCreated:
				response.Stages[0].State = FlowStageInProcess
				response.CurrentStage = 1

			case scm.ApprovalWaitForAccepted:
				response.Stages[0].State = FlowStagePassed
				response.Stages[1].State = FlowStageInProcess
				response.CurrentStage = 2

			default:
				stageAccepted = false
			}
		} else {
			switch extra.OldStage {
			case scm.ApprovalCreated:
				switch extra.NewStage {
				case scm.ApprovalWaitForAccepted:
					response.Stages[0].State = FlowStagePassed
					response.Stages[1].State = FlowStageInProcess
					response.CurrentStage = 2

				case scm.ApprovalAccepted:
					response.Stages[0].State = FlowStagePassed
					response.Stages[1].State = FlowStageSkip
					response.Stages[2].State = FlowStagePassed
					response.CurrentStage = 3

				case scm.ApprovalRejected:
					response.Stages[0].State = FlowStagePassed
					response.Stages[1].State = FlowStageRejected
					response.CurrentStage = 2
				default:
					stageAccepted = false
				}

			case scm.ApprovalWaitForAccepted:
				switch extra.NewStage {
				case scm.ApprovalAccepted:
					response.Stages[0].State = FlowStagePassed
					response.Stages[1].State = FlowStagePassed
					response.Stages[2].State = FlowStagePassed
					response.CurrentStage = 3

				case scm.ApprovalRejected:
					response.Stages[0].State = FlowStagePassed
					response.Stages[1].State = FlowStageRejected
					response.CurrentStage = 2
				default:
					stageAccepted = false
				}
			default:
				stageAccepted = false
			}
		}
	}
	if !stageAccepted {
		rctx.OpCtx.Log.Warnf("Unrecognized state changing: %v -- > %v", extra.OldStage, extra.NewStage)
	}
	response.ID = request.ApprovalID
	rctx.Response.Data = response
	rctx.Succeed()
}

func GetJobScript(ctx *gin.Context) {
	return
}

func GetGitlabCIIncludingJobs(ctx *gin.Context) {
	return
}
