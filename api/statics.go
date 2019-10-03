package api

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path/filepath"
	"regexp"
	"time"

	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	_ "git.stuhome.com/Sunmxt/wing/statik"
)

var RegexpStaticFilePath *regexp.Regexp

func init() {
	var err error
	if RegexpStaticFilePath, err = regexp.Compile("^\\/*(.*?)$"); err != nil {
		log.Panic("[statics] failed to compile regexp: ", err.Error())
	}
}

func ServeDefault(ctx *gin.Context) {
	rctx, path := acommon.NewRequestContext(ctx), ctx.Request.URL.Path

	rawPaths := RegexpStaticFilePath.FindStringSubmatch(path)
	if rawPaths == nil || len(rawPaths) < 2 {
		rctx.OpCtx.Log.Infof("[statics] invalid request path: %v. serve index.", ctx.Request.URL.Path)
		path = "/index.html"
	} else {
		path = rawPaths[1]
	}
	file, stat, err := acommon.TryOpenStaticFile(rctx, filepath.Join("/dashboard", path), "/dashboard/index.html")
	if err != nil {
		rctx.AbortWithDebugMessage(http.StatusInternalServerError, "cannot serve statics: "+err.Error())
		return
	}
	rctx.OpCtx.Log.Info("[statics] serve static: " + path)
	modTime := time.Time{}
	if stat != nil {
		modTime = stat.ModTime()
	}
	http.ServeContent(ctx.Writer, ctx.Request, path, modTime, file)
}
