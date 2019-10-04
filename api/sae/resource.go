package sae

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	acommon "git.stuhome.com/Sunmxt/wing/api/common"
)

func GetSAEResource(ctx *gin.Context) {
	rctx, resourcePath := acommon.NewRequestContext(ctx), ctx.Param("resource")
	if resourcePath == "" {
		rctx.FailCodeWithMessage(http.StatusNotFound, "not found")
		return
	}
	file, stat, err := acommon.TryOpenStaticFile(rctx, filepath.Join("/sae", resourcePath), "")
	if err != nil {
		rctx.AbortWithDebugMessage(http.StatusNotFound, "cannot serve statics: "+err.Error())
		return
	}
	modTime := time.Time{}
	if stat != nil {
		modTime = stat.ModTime()
	}
	http.ServeContent(ctx.Writer, ctx.Request, resourcePath, modTime, file)
}
