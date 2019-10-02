package scm

import (
	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/model/scm"
	"git.stuhome.com/Sunmxt/wing/model/scm/gitlab"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
	"strconv"
)

type RepositoryEntry struct {
	PlatformID   uint   `json:"platform_id"`
	PlatformName string `json:"platform_name"`
	RepositoryID uint   `json:"repo_id"`
	OwnerID      uint   `json:"owner_id"`
	Owner        string `json:"owner"`
	CIEnabled    bool   `json:"ci_enabled"`
	Name         string `json:"name"`
	FullName     string `json:"full_name"`
	Description  string `json:"description"`
	URL          string `json:"url"`

	repoReference string `json:"-"`
}

type ListRepositoryResponse struct {
	Repository []RepositoryEntry `json:"repository"`
	TotalCount uint              `json:"total_count"`
	TotalPage  uint              `json:"total_page"`
	Page       uint              `json:"page"`
}

type ListRepositoryRequest struct {
	PlatformID uint `form:"platform_id" binding:"required"`
	Limit      uint `form:"limit"`
	Page       uint `form:"page"`
}

func (r *ListRepositoryRequest) Clean(ctx *acommon.RequestContext) error {
	if r.Limit < 1 {
		r.Limit = 10
	}
	if r.Page < 1 {
		r.Page = 1
	}
	return nil
}

func pickGitlabSCMRepository(platformID uint, platformName string, projects []gitlab.Project) (entries []RepositoryEntry) {
	entries = make([]RepositoryEntry, len(projects))
	for idx, project := range projects {
		entry := &entries[idx]
		entry.PlatformID = platformID
		entry.PlatformName = platformName
		entry.RepositoryID = project.ID
		entry.Name = project.Name
		entry.FullName = platformName + "/" + project.Namespace.Path + "/" + project.Name
		entry.Description = project.Description
		entry.URL = project.WebURL
		entry.repoReference = strconv.FormatUint(uint64(project.ID), 10)
	}
	return
}

func ListRepository(ctx *gin.Context) {
	rctx, request, response := acommon.NewRequestContext(ctx), &ListRepositoryRequest{}, &ListRepositoryResponse{}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.BindOrFail(request) {
		return
	}

	platform := &scm.SCMPlatform{}
	if err := db.Model(platform).Where("id = ? and active = ?", request.PlatformID, scm.Active).First(platform).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.AbortWithError(common.ErrSCMPlatformNotFound)
			return
		} else {
			rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
			rctx.OpCtx.Log.Error(err.Error())
			return
		}
	}

	switch platform.Type {
	case scm.GitlabSCM:
		client, err := platform.GitlabClient(rctx.OpCtx.Log)
		if err != nil {
			rctx.AbortWithError(err)
			return
		}
		query := client.ProjectQuery()
		query.PerPage(request.Limit)
		query.Page(request.Page)
		if query.Refresh().Error != nil {
			rctx.AbortWithError(err)
			return
		}
		response.Repository = pickGitlabSCMRepository(uint(platform.ID), platform.Name, query.Projects)
		response.TotalCount = query.Cursor.Total
		response.TotalPage = query.Cursor.TotalPage
		response.Page = query.Cursor.Page

	default:
		rctx.AbortWithError(common.ErrSCMPlatformNotSupported)
	}

	var ciRepos []scm.CIRepository
	repoRefs := make([]string, len(response.Repository))
	for idx := range response.Repository {
		repoRefs[idx] = response.Repository[idx].repoReference
	}
	if err := db.Where(&scm.CIRepository{
		Active:        scm.Active,
		SCMPlatformID: platform.Basic.ID,
	}).Where("reference in (?)", repoRefs).Order("modify_time desc").Select("reference, owner_id").Preload("Owner").Find(&ciRepos).Error; err != nil {
		rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
		rctx.OpCtx.Log.Error(err.Error())
		return
	}
	mapActiveCIRepos := map[string]*scm.CIRepository{}
	for idx := range ciRepos {
		mapActiveCIRepos[ciRepos[idx].Reference] = &ciRepos[idx]
	}

	for _, entry := range response.Repository {
		ciRepo := mapActiveCIRepos[entry.repoReference]
		if ciRepo == nil {
			continue
		}

		entry.CIEnabled = true
		if ciRepo.OwnerID > 0 {
			entry.OwnerID = uint(ciRepo.OwnerID)
			entry.Owner = ciRepo.Owner.Name
		}
	}

	rctx.Response.Data = response
	rctx.Succeed()
}
