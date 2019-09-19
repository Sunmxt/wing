package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rakyll/statik/fs"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"regexp"
	"time"

	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	_ "git.stuhome.com/Sunmxt/wing/statik"
)

var StaticsFS http.FileSystem
var RegexpStaticFilePath *regexp.Regexp

func init() {
	s, err := fs.New()
	if err != nil {
		log.Panic("[statics] failed to load static http filesystem: ", err.Error())
	}
	StaticsFS = s

	if RegexpStaticFilePath, err = regexp.Compile("^\\/*(.*?)$"); err != nil {
		log.Panic("[statics] failed to compile regexp: ", err.Error())
	}
}

func tryOpenStaticFile(rctx *acommon.RequestContext, path string, defaultPath string) (file http.File, stat os.FileInfo, err error) {
	serveDefault := false
	setDefault := func() {
		rctx.OpCtx.Log.Infof("[statics] serve index \"%v\" for path \"%v\".", defaultPath, path)
		serveDefault = true
		path = defaultPath
		file = nil
		stat = nil
	}

	for {
		file, err = StaticsFS.Open(path)
		if err != nil {
			if err == os.ErrNotExist && !serveDefault {
				setDefault()
				continue
			}
			return nil, nil, err
		}
		if serveDefault {
			return file, stat, err
		}
		if file == nil {
			setDefault()
			continue
		}
		stat, err = file.Stat()
		if stat.IsDir() || err != nil {
			setDefault()
			continue
		}
		break
	}
	return file, stat, nil
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
	file, stat, err := tryOpenStaticFile(rctx, "/"+path, "/index.html")
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
