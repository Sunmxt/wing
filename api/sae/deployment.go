package sae

import (
	"net/http"

	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	csae "git.stuhome.com/Sunmxt/wing/controller/sae"
	"git.stuhome.com/Sunmxt/wing/model/account"
	"git.stuhome.com/Sunmxt/wing/model/sae"
	"github.com/gin-gonic/gin"
)

type GetDeploymentDetailRequest struct {
	DeploymentID int `form:"deployment_id" binding:"required"`
}

func (r *GetDeploymentDetailRequest) Clean(ctx *acommon.RequestContext) error {
	return nil
}

type GetDeploymentDetailResponse struct {
	State             int                 `json:"state"`
	Stages            []acommon.FlowStage `json:"stages"`
	CurrentStageIndex int                 `json:"current_stage_index"`
}

func GetDeploymentDetail(ctx *gin.Context) {
	rctx, request := acommon.NewRequestContext(ctx), &DeploymentNextRequest{}
	if !rctx.BindOrFail(request) || !rctx.LoginEnsured(true) {
		return
	}
	db := rctx.DatabaseOrFail()
	if db == nil {
		return
	}
}

type DeploymentNextRequest struct {
	DeploymentID int `form:"deployment_id" binding:"required"`
}

func (r *DeploymentNextRequest) Clean(ctx *acommon.RequestContext) error {
	return nil
}

func DeploymentTriggerNext(ctx *gin.Context) {
	rctx, request := acommon.NewRequestContext(ctx), &DeploymentNextRequest{}
	if !rctx.BindOrFail(request) || !rctx.LoginEnsured(true) {
		return
	}
	db := rctx.DatabaseOrFail()
	if db == nil {
		return
	}
	var deployment *sae.ApplicationDeployment
	if err := deployment.ByID(db, request.DeploymentID); err != nil {
		rctx.AbortWithError(err)
		return
	}
	if !rctx.PermitDeploymentOrReject(account.VerbUpdate, deployment) {
		rctx.FailCodeWithMessage(http.StatusForbidden, "Auth.LackOfPermission")
		return
	}
	if deployment.ID < 1 {
		rctx.FailWithMessage("SAE.DeploymentNotFound")
		return
	}
	if err := csae.DeploymentTriggerNext(&rctx.OpCtx, nil, deployment); err != nil {
		rctx.AbortWithError(err)
		return
	}
	rctx.Succeed()
}
