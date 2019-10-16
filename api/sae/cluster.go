package sae

import (
	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	"github.com/gin-gonic/gin"
)

type CreateApplicationCluster struct {
	OrcherstratorID int `form:"orchestrator_id"`
}

func (c *CreateApplicationCluster) Clean(ctx *acommon.RequestContext) error {
	return nil
}

func CreateApplicationCluster(ctx *gin.Context) {
}
