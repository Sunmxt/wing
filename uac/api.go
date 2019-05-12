package uac

import (
	"git.stuhome.com/Sunmxt/wing/common"
	"github.com/gin-gonic/gin"
)

func RegisterAPI(engine *gin.Engine) error {
	engine.GET("/api/login", AuthLoginV1)
	return nil
}

func AuthLoginV1(ctx *gin.Context) {
	rctx := common.NewRequestContext(ctx)
	ctx.JSON(200, rctx.Response)
}
