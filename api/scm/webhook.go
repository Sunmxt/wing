package scm

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"

	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/model/scm"
	"net/http"
)

func GitlabWebhookCallbackWithToken(ctx *gin.Context) {
	GitlabWebhookCallback(ctx)
	ctx.Request.URL.Path = ctx.Request.URL.Path[:strings.LastIndex(ctx.Request.URL.Path, "/")] + "/<masked token>"
}

func GitlabWebhookCallback(ctx *gin.Context) {
	rctx, code := acommon.NewRequestContext(ctx), uint(0)
	rawPlatformID, token := ctx.Param("platform_id"), ctx.Param("token")
	db := rctx.DatabaseOrFail()
	if db == nil {
		return
	}
	platformID, err := strconv.ParseUint(rawPlatformID, 10, 64)
	if err != nil {
		rctx.AbortCodeWithError(http.StatusBadRequest, err)
		return
	}
	platform := scm.SCMPlatform{}
	if err = platform.ByID(db, int(platformID)); err != nil {
		rctx.AbortWithError(err)
		return
	}
	if platform.Basic.ID < 1 {
		rctx.AbortCodeWithError(http.StatusNotFound, common.ErrSCMPlatformNotFound)
		return
	}
	// verify token
	if platform.Token != "" {
		if token == "" {
			token = ctx.Request.Header.Get("X-Gitlab-Token")
		}
		if token != platform.Token {
			rctx.FailCodeWithMessage(http.StatusForbidden, "Auth.Unauthenticated")
			return
		}
	}
	if code, err = rctx.OpCtx.Runtime.GitlabWebhookEventHub.ProcessWebhook(ctx.Request); err != nil {
		rctx.AbortCodeWithError(code, err)
		return
	}
	rctx.Succeed()
}
