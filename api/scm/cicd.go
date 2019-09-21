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

//type DisableRepositoryCICDRequest struct {
//	PlatformID   uint   `form:"platform_id" binding:"required"`
//	RepositoryID string `form:"repository_id" binding:"required"`
//}
//
//func DisableRepositoryCICD(ctx *gin.Context) {
//	rctx, request := acommon.NewRequestContext(ctx), &DisableRepositoryCICDRequest{}
//	db := rctx.DatabaseOrFail()
//	if db == nil || !rctx.BindOrFail(request) {
//		return
//	}
//}
