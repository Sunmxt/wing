package scm

import (
	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	"git.stuhome.com/Sunmxt/wing/model/scm"
	"github.com/gin-gonic/gin"
	"net/http"
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
	// TODO: Verify permission: Admin

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

type SCMPlatformDetailRequest struct {
}

func SCMPlatformDetail(ctx *gin.Context) {
}
