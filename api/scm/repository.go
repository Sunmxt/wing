package scm

import (
	"github.com/gin-gonic/gin"
	"git.stuhome.com/Sunmxt/wing/model/scm"
	"github.com/jinzhu/gorm"
)

type RepositoryInformationEntry struct {
}

type ListRepositoryResponse struct {
}

func ListRepository(ctx *gin.Context) {
	rctx := NewRequestContext(ctx)
}

type SCMPlatformEntry struct {
	ID int `json:"id"`
	TypeID int `json:"type_id"`
	Type int `json:"type"`
	Name string `json:"name"`
	RepoURL string `json:"repo_url"`
}

type ListSCMPlatformResponse struct {
	Repositories []SCMPlatformEntry `json:"repos"`
	TotalCount `json:"total_count"`
	TotalPage  `json:"total_page"`
}

type ListSCMPlatformRequest struct {
	Limit uint `form:"limit" binding:"min=1"`
	
}

func ListSCMPlatform(ctx *gin.Context) {
	rctx, request, response := NewRequestContext(ctx), ctx.Request, &ListRepositoryResponse{}
	db := rctx.DatabaseOrFail()
	if db == nil {
		return
	}

	if err := db.Model(&model.SCMPlatform{}).OrderBy("modify_time desc").Scan(&response.Repositories).Error; err != nil {
		rctx.OpCtx.Log.Error("list SCM platform failure: " + err.Error())
	}
}
