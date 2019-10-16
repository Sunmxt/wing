package sae

import (
	"errors"
	"net/http"

	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	"git.stuhome.com/Sunmxt/wing/model/account"
	mcommon "git.stuhome.com/Sunmxt/wing/model/common"
	"git.stuhome.com/Sunmxt/wing/model/sae"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type CreateOrchestratorRequest struct {
	Type   int      `form:"type" binding:"required"`
	Name   string   `form:"name" binding:"required"`
	Config string   `form:"config"`
	Params []string `form:"params"`
}

func (r *CreateOrchestratorRequest) Clean(ctx *acommon.RequestContext) error {
	switch r.Type {
	case sae.Kubernetes:
	case sae.KubernetesIncluster:
	default:
		return errors.New("SAE.InvalidOrchestratorType")
	}
	return nil
}

type CreateOrchestratorResponse struct {
	ID int `json:"id"`
}

func CreateOrchestrator(ctx *gin.Context) {
	rctx, request, response := acommon.NewRequestContext(ctx), &CreateOrchestratorRequest{}, &CreateOrchestratorResponse{}
	if !rctx.LoginEnsured(true) || !rctx.BindOrFail(request) {
		return
	}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.PermitOrReject("orchestrator", account.VerbCreate) {
		return
	}
	user := rctx.GetAccount()
	orcher := &sae.Orchestrator{
		Name:    request.Name,
		Active:  mcommon.Active,
		Type:    request.Type,
		OwnerID: user.Basic.ID,
	}
	switch request.Type {
	case sae.KubernetesIncluster:
		extra := &sae.KubernetesOrchestrator{
			Namespace: "default",
		}
		if len(ctx.Params) > 0 {
			extra.Namespace = request.Params[0]
		}
		if err := orcher.EncodeExtra(extra); err != nil {
			rctx.AbortWithError(err)
			return
		}
	case sae.Kubernetes:
		extra := &sae.KubernetesOrchestrator{
			Kubeconfig: request.Config,
			Namespace:  "default",
		}
		if len(ctx.Params) > 0 {
			extra.Namespace = request.Params[0]
		}
		if err := orcher.EncodeExtra(extra); err != nil {
			rctx.AbortWithError(err)
			return
		}
	default:
		rctx.FailCodeWithMessage(http.StatusBadRequest, "SCM.InvalidOrchestratorType")
		return
	}
	if err := db.Save(orcher).Error; err != nil {
		rctx.AbortWithError(err)
		return
	}
	response.ID = orcher.Basic.ID
	rctx.Response.Data = response
	rctx.Succeed()
}

type ListOrchestratorRequest struct {
	Limit uint `form:"limit"`
	Page  uint `form:"page"`
}

func (r *ListOrchestratorRequest) Clean(ctx *acommon.RequestContext) error {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.Limit < 1 {
		r.Limit = 10
	}
	return nil
}

type ListOrchestratorEntry struct {
	Type int    `json:"type"`
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type ListOrchestratorResponse struct {
	Entries    []ListOrchestratorEntry
	TotalCount uint `json:"total_count"`
	TotalPage  uint `json:"total_page"`
	Page       uint `json:"page"`
}

func ListOrchestrator(ctx *gin.Context) {
	rctx, request, response := acommon.NewRequestContext(ctx), &ListOrchestratorRequest{}, &ListOrchestratorResponse{}
	if !rctx.LoginEnsured(true) || !rctx.BindOrFail(request) {
		return
	}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.PermitOrReject("orchestrator", account.VerbGet) {
		return
	}
	// require: Admin
	var orchers []sae.Orchestrator
	q := db.Where("active in (?)", []int{mcommon.Active, mcommon.Disabled}).Model(&sae.Orchestrator{})
	if err := q.Count(&response.TotalCount).Error; err != nil {
		rctx.AbortWithError(err)
		return
	}
	response.TotalPage = ((response.TotalCount - 1) / request.Limit) + 1
	q = q.Offset(request.Limit * (request.Page - 1)).Limit(request.Limit)
	if err := q.Find(&orchers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		rctx.AbortWithError(err)
		return
	}
	response.Entries = make([]ListOrchestratorEntry, len(orchers))
	for idx, orcher := range orchers {
		ref := &response.Entries[idx]
		ref.Type = orcher.Type
		ref.ID = orcher.Basic.ID
		ref.Name = orcher.Name
	}
	rctx.Response.Data = response
	rctx.Succeed()
}

type DisableOrchestratorRequest struct {
	ID int `form:"id"`
}

func (r *DisableOrchestratorRequest) Clean(ctx *acommon.RequestContext) error {
	return nil
}

func updateOrcherstratorState(ctx *acommon.RequestContext, ID, active int) *sae.Orchestrator {
	db := ctx.DatabaseOrFail()
	if db == nil {
		return nil
	}
	tx := db.Begin()
	orcher := &sae.Orchestrator{}
	if err := orcher.ByID(tx.Select("id").Where("active in (?)", []int{mcommon.Disabled, mcommon.Active}), ID); err != nil {
		tx.Rollback()
		ctx.AbortWithError(err)
		return nil
	}
	if orcher.Basic.ID < 1 {
		tx.Rollback()
		ctx.FailCodeWithMessage(http.StatusNotFound, "SAE.OrchestratorNotFound")
		return nil
	}
	orcher.Active = active
	if err := tx.Model(orcher).Update("active", active).Where("id", ID).Error; err != nil {
		tx.Rollback()
		ctx.AbortWithError(err)
		return nil
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		ctx.AbortWithError(err)
		return nil
	}
	return orcher
}

func DisableOrchestrator(ctx *gin.Context) {
	rctx, request := acommon.NewRequestContext(ctx), &DisableOrchestratorRequest{}
	if !rctx.LoginEnsured(true) || !rctx.BindOrFail(request) ||
		!rctx.PermitOrReject("orchestrator", account.VerbUpdate) {
		return
	}
	if updateOrcherstratorState(rctx, request.ID, mcommon.Disabled) == nil {
		return
	}
	rctx.OpCtx.Log.Infof("Disable orchestrator %v", request.ID)
	rctx.Succeed()
}

type EnableOrchestratorRequest struct {
	ID int `form:"id"`
}

func (r *EnableOrchestratorRequest) Clean(ctx *acommon.RequestContext) error {
	return nil
}

func EnableOrchestrator(ctx *gin.Context) {
	rctx, request := acommon.NewRequestContext(ctx), &EnableOrchestratorRequest{}
	if !rctx.LoginEnsured(true) || !rctx.BindOrFail(request) ||
		!rctx.PermitOrReject("orchestrator", account.VerbUpdate) {
		return
	}
	if updateOrcherstratorState(rctx, request.ID, mcommon.Active) == nil {
		return
	}
	rctx.Succeed()
}

type DeleteOrchestratorRequest struct {
	ID int `form:"id"`
}

func (r *DeleteOrchestratorRequest) Clean(ctx *acommon.RequestContext) error {
	return nil
}

func DeleteOrchestrator(ctx *gin.Context) {
	rctx, request := acommon.NewRequestContext(ctx), &DeleteOrchestratorRequest{}
	if !rctx.LoginEnsured(true) || !rctx.BindOrFail(request) ||
		!rctx.PermitOrReject("orchestrator", account.VerbUpdate) {
		return
	}
	if updateOrcherstratorState(rctx, request.ID, mcommon.Inactive) == nil {
		return
	}
	rctx.Succeed()
}
