package scm

import (
	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/model/scm"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
	"strconv"
)

type SCMPlatformEntry struct {
	ID          int    `json:"id"`
	TypeID      uint   `json:"type_id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PublicURL   string `json:"public_url"`
}

type ListSCMPlatformResponse struct {
	Platform   []SCMPlatformEntry `json:"platform"`
	TotalCount uint               `json:"total_count"`
	TotalPage  uint               `json:"total_page"`
	Page       uint               `json:"page"`
}

type ListSCMPlatformRequest struct {
	Limit uint `form:"limit"`
	Page  uint `form:"page"`
}

func (r *ListSCMPlatformRequest) Clean(ctx *acommon.RequestContext) error {
	if r.Limit < 1 {
		r.Limit = 10
	}
	if r.Page < 1 {
		r.Page = 1
	}
	return nil
}

func ListSCMPlatform(ctx *gin.Context) {
	rctx, request, response := acommon.NewRequestContext(ctx), &ListSCMPlatformRequest{}, &ListSCMPlatformResponse{}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.BindOrFail(request) {
		return
	}

	scmQuery, totalCount := db.Model(&scm.SCMPlatform{}).Order("modify_time desc"), uint(0)
	if err := scmQuery.Count(&totalCount).Error; err != nil {
		rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
		rctx.OpCtx.Log.Error(err.Error())
		return
	}
	response.TotalCount = totalCount
	response.TotalPage = (totalCount / request.Limit) + 1
	response.Page = request.Page

	scmRawEntries := make([]scm.SCMPlatform, 0, request.Limit)

	scmQuery = scmQuery.Offset(request.Limit * (request.Page - 1)).Limit(request.Limit)
	if err := scmQuery.Scan(&scmRawEntries).Error; err != nil {
		rctx.OpCtx.Log.Error("list SCM platform failure: " + err.Error())
		rctx.AbortWithDebugMessage(http.StatusInternalServerError, "list SCM platform failure:"+err.Error())
		return
	}

	// Pick
	response.Platform = make([]SCMPlatformEntry, len(scmRawEntries))
	for idx, scmRawEntry := range scmRawEntries {
		var exists bool

		entry := &response.Platform[idx]
		entry.ID = scmRawEntry.Basic.ID
		entry.TypeID = uint(scmRawEntry.Type)
		entry.Type, exists = scm.SCMPlatformTypeString[entry.TypeID]
		if !exists {
			entry.Type = "Unknown"
		}
		entry.Name = scmRawEntry.Name
		entry.Description = scmRawEntry.Description
		entry.PublicURL = scmRawEntry.PublicURL()
	}
	rctx.Response.Data = response
	rctx.Succeed()
}

// List repository
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

func pickGitlabSCMRepository(platformID uint, platformName string, projects []scm.GitlabProject) (entries []RepositoryEntry) {
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
		query, err := platform.GitlabProjectQuery()
		if err != nil {
			rctx.AbortWithError(err)
			return
		}
		query.PerPage(request.Limit)
		query.Page(request.Page)
		query.Logger = rctx.OpCtx.Log
		err = query.Refresh().Error
		if err != nil {
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
