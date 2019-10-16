package scm

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"
	"time"

	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/controller/cicd"
	csae "git.stuhome.com/Sunmxt/wing/controller/sae"
	"git.stuhome.com/Sunmxt/wing/model/sae"
	"git.stuhome.com/Sunmxt/wing/model/scm"
	"git.stuhome.com/Sunmxt/wing/model/scm/gitlab"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"gopkg.in/yaml.v2"
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
	PlatformID   uint `form:"platform_id" binding:"required"`
	RepositoryID uint `form:"repository_id" binding:"required"`
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
	repo := &scm.CIRepository{
		SCMPlatformID: int(request.PlatformID),
		Reference:     strconv.FormatUint(uint64(request.RepositoryID), 10),
		Active:        scm.Active,
	}
	err := db.Where(repo).First(&repo).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.AbortCodeWithError(http.StatusNotFound, common.ErrRepositoryNotFound)
			return
		}
		rctx.AbortWithError(err)
		return
	}
	repo.Active = scm.Inactive
	if err = db.Save(repo).Error; err != nil {
		rctx.AbortWithError(err)
		return
	}
	rctx.Succeed()
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

type ListBuildRequest struct {
	PlatformID   int `form:"platform_id"`
	RepositoryID int `form:"repository_id"`
	Page         int `form:"page"`
	Limit        int `form:"limit"`
}

func (r *ListBuildRequest) Clean(rctx *acommon.RequestContext) error {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.Limit < 1 {
		r.Limit = 10
	}
	return nil
}

type ListBuildResponseEntry struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Command      string `json:"command"`
	Branch       string `json:"branch"`
	BuildID      int    `json:"build_id"`
	PlatformID   int    `json:"platform_id"`
	RepositoryID int    `json:"repository_id"`
}

type ListBuildResponse struct {
	Page       int                      `json:"page"`
	Limit      int                      `json:"limit"`
	TotalPage  int                      `json:"total_pages"`
	TotalCount int                      `json:"total_count"`
	Builds     []ListBuildResponseEntry `json:"builds"`
}

func ListBuilds(ctx *gin.Context) {
	rctx, request, response := acommon.NewRequestContext(ctx), &ListBuildRequest{}, &ListBuildResponse{}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.BindOrFail(request) || !rctx.LoginEnsured(true) {
		return
	}
	response.Page = request.Page
	response.Limit = request.Limit
	builds := []*scm.CIRepositoryBuild{}
	q := db.Where("ci_repository_builds.active in (?)", []int{scm.Active, scm.Disabled}).Order("modify_time desc").Preload("Repository")
	if request.PlatformID < 1 {
		if request.RepositoryID > 0 {
			q = q.Where("repository_id in (?)", request.RepositoryID)
		}
	} else {
		repoIDs := []int(nil)
		q = q.Table("ci_repository_builds")
		qSCM := db.Table("ci_repository").Where("scm_platform_id = (?)", request.PlatformID)
		if request.RepositoryID > 0 {
			qSCM = qSCM.Where("reference = (?)", strconv.FormatInt(int64(request.RepositoryID), 10))
		}
		if err := qSCM.Pluck("distinct id", &repoIDs).Error; err != nil && gorm.IsRecordNotFoundError(err) {
			rctx.AbortWithError(err)
			return
		}
		if len(repoIDs) < 0 { // No matched.
			response.Builds = make([]ListBuildResponseEntry, 0)
			response.TotalPage = 0
			response.TotalCount = 0
			rctx.Succeed()
			return
		}
		q = q.Where("repository_id in (?)", repoIDs)
	}
	totalCount := 0
	if err := q.Count(&totalCount).Error; err != nil {
		rctx.AbortWithError(err)
		return
	}
	response.TotalCount = totalCount
	q = q.Offset((request.Page - 1) * request.Limit).Limit(request.Limit)
	if err := q.Find(&builds).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		rctx.AbortWithError(err)
		return
	}
	response.Builds = make([]ListBuildResponseEntry, len(builds))
	for idx, build := range builds {
		ref := &response.Builds[idx]
		ref.Name = build.Name
		ref.Description = build.Description
		ref.Command = build.BuildCommand
		ref.Branch = build.Branch
		ref.BuildID = build.Basic.ID
		ref.PlatformID = build.Repository.SCMPlatformID
		ref.RepositoryID = build.Repository.Basic.ID
	}
	response.TotalPage = (totalCount / (request.Limit + 1)) + 1
	rctx.Response.Data = response
	rctx.Succeed()
}

type CreateBuildRequest struct {
	Branch       string `form:"branch" binding:"required"`
	PlatformID   uint   `form:"platform_id" binding:"required"`
	RepositoryID uint   `form:"repository_id" binding:"required"`
	Command      string `form:"command" binding:"required"`
	ProductPath  string `form:"product_path" binding:"required"`
	Name         string `form:"name" binding:"required"`
	Description  string `form:"description"`
}

func (r *CreateBuildRequest) Clean(rctx *acommon.RequestContext) error {
	if r.Name == "" {
		return errors.New("name should not be empty.")
	}
	if r.Command == "" {
		return errors.New("command should not be empty")
	}
	if r.ProductPath == "" {
		return errors.New("product path should not be empty")
	}
	if r.Branch == "" {
		r.Branch = "master"
	}
	return nil
}

type CreateBuildResponse struct {
	BuildID int `json:"build_id"`
}

func CreateBuild(ctx *gin.Context) {
	rctx, request, response := acommon.NewRequestContext(ctx), &CreateBuildRequest{}, CreateBuildResponse{}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.BindOrFail(request) {
		return
	}
	tx := db.Begin()
	repo := &scm.CIRepository{
		SCMPlatformID: int(request.PlatformID),
		Reference:     strconv.FormatUint(uint64(request.RepositoryID), 10),
		Active:        scm.Active,
	}
	err := tx.Where(repo).First(repo).Select("id").Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.AbortWithError(common.ErrRepositoryNotFound)
			return
		}
		rctx.AbortWithError(err)
		return
	}
	build := &scm.CIRepositoryBuild{
		Name:         request.Name,
		Description:  request.Description,
		ExecType:     scm.GitlabCIBuild, // gitlab ci only now.
		Active:       scm.Active,
		BuildCommand: request.Command,
		ProductPath:  request.ProductPath,
		Branch:       request.Branch,
		RepositoryID: repo.Basic.ID,
	}
	if err = tx.Save(build).Error; err != nil {
		rctx.AbortWithError(err)
		return
	}
	if err = tx.Commit().Error; err != nil {
		rctx.AbortWithError(err)
		return
	}
	response.BuildID = build.Basic.ID
	rctx.Response.Data = response
	rctx.Succeed()
}

type EditBuildRequest struct {
	BuildID     int    `form:"build_id" binding:"required"`
	Command     string `form:"command"`
	ProductPath string `form:"product_path"`
	Name        string `form:"name"`
	Description string `form:"description"`
	Branch      string `form:"branch"`
}

func (r *EditBuildRequest) Clean(rctx *acommon.RequestContext) error {
	return nil
}

func authAndFetchBuild(db *gorm.DB, buildID, userID int) (*scm.CIRepositoryBuild, error) {
	build := &scm.CIRepositoryBuild{}
	if err := build.ByID(db.Preload("Repository").Where("active in (?)", []int{scm.Active, scm.Disabled}), buildID); err != nil {
		return nil, err
	}
	if build.Basic.ID < 1 {
		return nil, common.ErrBuildNotFound
	}
	if userID != build.Repository.OwnerID {
		return nil, common.ErrUnauthenticated
	}
	return build, nil
}

func EditBuild(ctx *gin.Context) {
	rctx, request := acommon.NewRequestContext(ctx), &EditBuildRequest{}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.BindOrFail(request) || !rctx.LoginEnsured(true) {
		return
	}
	tx, user := db.Begin(), rctx.GetAccount()
	build, err := authAndFetchBuild(tx, request.BuildID, user.Basic.ID)
	if err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	build.BuildCommand = request.Command
	build.ProductPath = request.ProductPath
	build.Branch = request.Branch
	build.Name = request.Name
	build.Description = request.Description
	if err = tx.Save(build).Error; err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	rctx.Succeed()
}

type DisableBuildRequest struct {
	BuildID int `form:"build_id" binding:"required"`
}

func (r *DisableBuildRequest) Clean(rctx *acommon.RequestContext) error {
	return nil
}

func DisableBuild(ctx *gin.Context) {
	rctx, request := acommon.NewRequestContext(ctx), &DisableBuildRequest{}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.BindOrFail(request) {
		return
	}
	tx, user := db.Begin(), rctx.GetAccount()
	build, err := authAndFetchBuild(tx, request.BuildID, user.Basic.ID)
	if err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	build.Active = scm.Disabled
	if err = tx.Save(build).Error; err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	rctx.Succeed()
}

type EnableBuildRequest struct {
	BuildID int `form:"build_id" binding:"required"`
}

func (r *EnableBuildRequest) Clean(rctx *acommon.RequestContext) error {
	return nil
}

func EnableBuild(ctx *gin.Context) {
	rctx, request := acommon.NewRequestContext(ctx), &DisableBuildRequest{}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.BindOrFail(request) {
		return
	}
	tx, user := db.Begin(), rctx.GetAccount()
	build, err := authAndFetchBuild(tx, request.BuildID, user.Basic.ID)
	if err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	build.Active = scm.Active
	if err = tx.Save(build).Error; err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	rctx.Succeed()
}

type DeleteBuildRequest struct {
	BuildID int `form:"build_id" binding:"required"`
}

func (r *DeleteBuildRequest) Clean(rctx *acommon.RequestContext) error {
	return nil
}

func DeleteBuild(ctx *gin.Context) {
	rctx, request := acommon.NewRequestContext(ctx), &DisableBuildRequest{}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.BindOrFail(request) {
		return
	}
	tx, user := db.Begin(), rctx.GetAccount()
	build, err := authAndFetchBuild(tx, request.BuildID, user.Basic.ID)
	if err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	build.Active = scm.Inactive
	if err = tx.Save(build).Error; err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	rctx.Succeed()
}

func GetGitlabCIIncludingJobs(ctx *gin.Context) {
	rctx := acommon.NewRequestContext(ctx)
	rawRepositoryID := ctx.Param("id")
	db := rctx.DatabaseOrFail()
	if db == nil {
		return
	}
	repositoryID, err := strconv.ParseUint(rawRepositoryID, 10, 64)
	if err != nil {
		rctx.AbortCodeWithError(http.StatusBadRequest, err)
		return
	}
	repo := &scm.CIRepository{}
	if err = db.Where("id = (?) and active = (?)", repositoryID, scm.Active).First(repo).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.AbortCodeWithError(http.StatusNotFound, common.ErrRepositoryNotFound)
			return
		}
		rctx.AbortWithError(err)
		return
	}
	var jobs map[string]*gitlab.CIJob
	if jobs, err = cicd.GenerateGitlabCIJobsForRepository(&rctx.OpCtx, uint(repositoryID)); err != nil {
		rctx.AbortWithError(err)
		return
	}
	buf := &bytes.Buffer{}
	if err = yaml.NewEncoder(buf).Encode(jobs); err != nil {
		rctx.AbortWithError(err)
		return
	}
	ctx.Writer.Write(buf.Bytes())
	ctx.Writer.WriteHeader(http.StatusOK)
}

func GetCIJob(ctx *gin.Context) {
	rctx := acommon.NewRequestContext(ctx)
	rawBuildID := ctx.Param("id")
	db := rctx.DatabaseOrFail()
	if db == nil {
		return
	}
	buildID, err := strconv.ParseUint(rawBuildID, 10, 64)
	if err != nil {
		rctx.AbortCodeWithError(http.StatusBadRequest, err)
		return
	}
	// verify permission.
	token := ctx.Request.Header.Get("Wing-Auth-Token")
	if token == "" {
		rctx.FailCodeWithMessage(http.StatusForbidden, common.ErrUnauthenticated.Error())
		return
	}
	build := &scm.CIRepositoryBuild{}
	if err = db.Where("id = (?)", buildID).Preload("Repository").First(build).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.FailCodeWithMessage(http.StatusForbidden, common.ErrUnauthenticated.Error())
			return
		}
		rctx.AbortWithError(err)
		return
	}
	if build.Repository.AccessToken != token {
		rctx.FailCodeWithMessage(http.StatusForbidden, common.ErrUnauthenticated.Error())
		return
	}
	buf := &bytes.Buffer{}
	if err = cicd.GenerateScriptForBuild(&rctx.OpCtx, buf, build); err != nil {
		rctx.AbortWithError(err)
		return
	}
	ctx.Writer.Write(buf.Bytes())
	ctx.Writer.WriteHeader(http.StatusOK)
}

type ReportBuildResultRequest struct {
	Type         uint   `form:"type" binding:"required"`
	Reason       string `form:"reason"`
	Succeed      bool   `form:"succeed"`
	Namespace    string `form:"namespace" binding:"required"`
	Environment  string `form:"environment" binding:"required"`
	Tag          string `form:"tag" binding:"required"`
	CommitHash   string `form:"commit_hash" binding:"required"`
	ProductToken string `form:"product_token" binding:"required"`
}

func (r *ReportBuildResultRequest) Clean(rctx *acommon.RequestContext) error {
	return nil
}

const (
	StartPackageReport  = 1
	FinishPackageReport = 2
)

func ReportBuildResult(ctx *gin.Context) {
	rctx, request := acommon.NewRequestContext(ctx), &ReportBuildResultRequest{}
	rawBuildID := ctx.Param("id")
	token := ctx.Request.Header.Get("Wing-Auth-Token")
	if token == "" {
		rctx.FailCodeWithMessage(http.StatusForbidden, common.ErrUnauthenticated.Error())
		return
	}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.BindOrFail(request) {
		return
	}
	buildID, err := strconv.ParseUint(rawBuildID, 10, 64)
	if err != nil {
		rctx.AbortCodeWithError(http.StatusBadRequest, err)
		return
	}
	build := &scm.CIRepositoryBuild{}
	if err = db.Where("id = (?)", buildID).Preload("Repository").First(build).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.FailCodeWithMessage(http.StatusForbidden, common.ErrUnauthenticated.Error())
			return
		}
		rctx.AbortWithError(err)
		return
	}
	if build.Repository.AccessToken != token {
		rctx.FailCodeWithMessage(http.StatusForbidden, common.ErrUnauthenticated.Error())
		return
	}
	var packageLog *scm.CIRepositoryLog
	product := &scm.CIRepositoryBuildProduct{
		CommitHash: request.CommitHash,
		Active:     scm.Active,
		BuildID:    build.Basic.ID,
	}
	tx := db.Begin()
	defer func() {
		if rec := recover(); rec != nil || err != nil {
			tx.Rollback()
		}
	}()
	if err = tx.Where("product_token = (?) and active in (?)", request.ProductToken, []int{scm.Active, scm.Disabled}).
		Order("create_time desc").Find(product).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		rctx.AbortWithError(err)
		return
	}
	if product.Basic.ID < 1 {
		product.ProductToken = request.ProductToken
	}
	switch request.Type {
	case StartPackageReport:
		if packageLog, _, err = scm.LogBuildPackage(tx, scm.CILogPackageStart, build.Basic.ID, request.Reason, request.Namespace,
			request.Environment, request.Tag, request.CommitHash); err != nil {
			rctx.AbortWithError(err)
			return
		}
		product.Stage = scm.ProductBuilding

	case FinishPackageReport:
		logType := scm.CILogPackageSucceed
		product.Stage = scm.ProductBuildSucceed
		if !request.Succeed {
			logType = scm.CILogPackageFailure
			product.Stage = scm.ProductBuildFailure
		}
		if packageLog, _, err = scm.LogBuildPackage(tx, logType, build.Basic.ID, request.Reason, request.Namespace,
			request.Environment, request.Tag, request.CommitHash); err != nil {
			rctx.AbortWithError(err)
			return
		}

	default:
		rctx.FailCodeWithMessage(http.StatusBadRequest, "invalid report type.")
		tx.Rollback()
		return
	}
	productExtra := scm.BuildProductExtra{
		Namespace:   request.Namespace,
		Environment: request.Environment,
		Tag:         request.Tag,
		LogID:       packageLog.Basic.ID,
	}
	if err = product.EncodeExtra(productExtra); err != nil {
		rctx.AbortWithError(err)
		return
	}
	if err = tx.Save(product).Error; err != nil {
		rctx.AbortWithError(err)
		return
	}
	if err = tx.Commit().Error; err != nil {
		rctx.AbortWithError(err)
		return
	}
	rctx.Succeed()
}

type ListProductRequest struct {
	BuildID int  `form:"build_id"`
	Page    uint `form:"page"`
	Limit   uint `form:limit`
}

func (r *ListProductRequest) Clean(rctx *acommon.RequestContext) error {
	if r.Limit < 1 {
		r.Limit = 1
	}
	if r.Page < 1 {
		r.Page = 1
	}
	return nil
}

type ListProductResponse struct {
	Page       uint           `json:"page"`
	Limit      uint           `json:"limit"`
	TotalPage  uint           `json:"total_pages"`
	TotalCount uint           `json:"total_count"`
	Products   []ProductEntry `json:"products"`
}

type ProductEntry struct {
	Namespace   string `json:"namespace"`
	Environment string `json:"environment"`
	Tag         string `json:"tag"`
	UpdateTime  string `json:"update_time"`
	CommitHash  string `json:"commit_hash"`
	State       int    `json:"state"`
	BuildID     int    `json:"build_id"`
}

const (
	ProductInProgress = 1
	ProductInFailure  = 2
	ProductInSuccess  = 3
)

func ListProduct(ctx *gin.Context) {
	rctx, request, response := acommon.NewRequestContext(ctx), &ListProductRequest{}, &ListProductResponse{}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.BindOrFail(request) || !rctx.LoginEnsured(true) {
		return
	}
	var products []scm.CIRepositoryBuildProduct
	q := db.Where("active in (?)", []int{scm.Disabled, scm.Active}).Order("create_time desc").
		Table("ci_repository_build_products")
	if request.BuildID > 0 {
		q = q.Where("build_id = (?)", request.BuildID)
	}
	totalCount := uint(0)
	if err := q.Count(&totalCount).Error; err != nil {
		rctx.AbortWithError(err)
		return
	}
	q = q.Offset((request.Page - 1) * request.Limit).Limit(request.Limit)
	if err := q.Find(&products).Error; err != nil && gorm.IsRecordNotFoundError(err) {
		rctx.AbortWithError(err)
		return
	}
	entries := make([]ProductEntry, len(products))
	for idx, product := range products {
		entry := &entries[idx]
		extra := &scm.BuildProductExtra{}
		if err := product.DecodeExtra(extra); err != nil {
			rctx.AbortWithError(err)
			return
		}
		entry.Namespace = extra.Namespace
		entry.Environment = extra.Environment
		entry.Tag = extra.Tag
		entry.CommitHash = product.CommitHash
		entry.BuildID = product.BuildID
		entry.UpdateTime = product.Basic.ModifyTime.Format(time.RFC3339)
		switch product.Stage {
		case scm.ProductBuilding:
			entry.State = ProductInProgress
		case scm.ProductBuildSucceed:
			entry.State = ProductInSuccess
		case scm.ProductBuildFailure:
			entry.State = ProductInFailure
		default:
			rctx.OpCtx.Log.Warnf("invalid product state: %v for build %v commit %v", product.Stage, product.BuildID, product.CommitHash)
		}
	}
	response.Page = request.Page
	response.Limit = request.Limit
	response.TotalCount = totalCount
	response.TotalPage = (totalCount / (request.Limit + 1)) + 1
	response.Products = entries
	rctx.Response.Data = response
	rctx.Succeed()
}

func GetCIRuntimeBuildJob(ctx *gin.Context) {
	rctx, rawClusterID := acommon.NewRequestContext(ctx), ctx.Param("id")
	token := ctx.Request.Header.Get("Wing-Auth-Token")
	if token == "" {
		rctx.FailCodeWithMessage(http.StatusForbidden, common.ErrUnauthenticated.Error())
		return
	}
	db := rctx.DatabaseOrFail()
	if db == nil {
		return
	}
	clusterID, err := strconv.ParseInt(rawClusterID, 10, 64)
	if err != nil {
		rctx.AbortCodeWithError(http.StatusBadRequest, err)
		return
	}
	cluster := &sae.ApplicationCluster{}
	if err = cluster.ByID(db, int(clusterID)); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.FailCodeWithMessage(http.StatusForbidden, common.ErrUnauthenticated.Error())
			return
		}
		rctx.AbortWithError(err)
		return
	}
	// Auth
	var IDs []int
	if err = db.Model(sae.BuildDependency{}).
		Where("application_id = (?)", cluster.ApplicationID).
		Pluck("build_id", &IDs).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.FailCodeWithMessage(http.StatusForbidden, common.ErrUnauthenticated.Error())
			return
		}
		rctx.AbortCodeWithError(http.StatusForbidden, err)
		return
	}
	q := db.Model(scm.CIRepositoryBuild{}).Where("id in (?) and active in (?)", IDs, []int{scm.Active, scm.Disabled})
	IDs = IDs[0:0]
	if err = q.Pluck("repository_id", &IDs).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.FailCodeWithMessage(http.StatusForbidden, common.ErrUnauthenticated.Error())
			return
		}
		rctx.AbortCodeWithError(http.StatusForbidden, err)
		return
	}
	var validTokens []string
	if err = db.Model(scm.CIRepository{}).Where("id in (?)").Pluck("token", validTokens).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.FailCodeWithMessage(http.StatusForbidden, common.ErrUnauthenticated.Error())
			return
		}
		rctx.AbortCodeWithError(http.StatusForbidden, err)
		return
	}
	valid := true
	for _, validToken := range validTokens {
		if token == validToken {
			valid = true
			break
		}
	}
	if !valid {
		rctx.FailCodeWithMessage(http.StatusForbidden, common.ErrUnauthenticated.Error())
		return
	}
	// generate script according to active deployment.
	var deployment *sae.ApplicationDeployment
	if deployment, err = csae.GetActiveDeployment(&rctx.OpCtx, cluster); err != nil {
		rctx.AbortWithError(err)
		return
	}
	if deployment == nil || (deployment.State != sae.DeploymentImageBuildInProgress &&
		deployment.State != sae.DeploymentImageBuildFinished &&
		deployment.State != sae.DeploymentCreated) { // no active deployment.
		ctx.Writer.WriteHeader(http.StatusOK)
		return
	}
	buf := &bytes.Buffer{}
	if err = cicd.GenerateScriptForGitlabCIRuntimeBuild(&rctx.OpCtx, buf, nil); err != nil {
		rctx.AbortWithError(err)
		return
	}
	ctx.Writer.Write(buf.Bytes())
	ctx.Writer.WriteHeader(http.StatusOK)
}

func ReportRuntimeBuildResult(ctx *gin.Context) {
	//rctx, rawClusterID := acommon.NewRequestContext(ctx), ctx.Param("id")
	//token := ctx.Request.Header.Get("Wing-Auth-Token")
	//if token == "" {
	//	rctx.FailCodeWithMessage(http.StatusForbidden, common.ErrUnauthenticated.Error())
	//	return
	//}
	//db := rctx.DatabaseOrFail()
	//if db == nil || !rctx.BindOrFail(request) {
	//	return
	//}
}
