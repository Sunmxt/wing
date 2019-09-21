package common

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"git.stuhome.com/Sunmxt/wing/common"
	ccommon "git.stuhome.com/Sunmxt/wing/controller/common"
	mlog "git.stuhome.com/Sunmxt/wing/log"
	"git.stuhome.com/Sunmxt/wing/model/account"
	ss "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	validator "gopkg.in/go-playground/validator.v8"
)

type APIRequest interface {
	Clean(ctx *RequestContext) error
}

func getLogger(ctx *gin.Context) (logger *log.Entry) {
	raw, ok := ctx.Get("logger")
	if ok {
		logger, ok = raw.(*log.Entry)
		if ok {
			return logger
		}
	}
	return mlog.RequestLogger(ctx)
}

func getRuntime(ctx *gin.Context) (runtime *common.WingRuntime) {
	raw, ok := ctx.Get("runtime")
	if !ok {
		return nil
	}
	rt, _ := raw.(*common.WingRuntime)
	return rt
}

func FormatValidateErrorMessage(ctx *RequestContext, errs validator.ValidationErrors) error {
	invalidFields := []string{}
	for _, err := range errs {
		invalidFields = append(invalidFields, err.Field)
	}
	return errors.New(ctx.TranslateMessage("Partial.InvalidFields") + ":" + strings.Join(invalidFields, ", "))
}

func ParseHeaderForLocale(ctx *gin.Context, acceptedLangs ...string) (result string) {
	raws, exists := ctx.Request.Header["Accept-Language"]
	if !exists || len(raws) < 1 {
		return ""
	}
	tagMap := make(map[string]language.Tag)
	for _, lang := range acceptedLangs {
		tag := message.MatchLanguage(lang)
		if tag != language.Und {
			tagMap[tag.String()] = tag
		}
	}

	raw, weightCur := raws[0], float64(0)
	for _, grp := range strings.Split(raw, ",") {
		langParam := strings.Split(grp, ";")
		if len(langParam) < 1 || len(langParam) > 2 {
			continue
		}
		lang := langParam[0]
		tag := message.MatchLanguage(lang)
		if tag == language.Und {
			continue
		}
		if _, exists = tagMap[tag.String()]; !exists {
			continue
		}
		if len(langParam) == 2 {
			weight, err := strconv.ParseFloat(strings.TrimPrefix(langParam[1], "q="), 64)
			if err != nil {
				continue
			}
			if weight > weightCur {
				weightCur = weight
				result = lang
			}
		}
	}
	return result
}

type RequestContext struct {
	Gin      *gin.Context
	OpCtx    ccommon.OperationContext
	Response common.Response
	Session  ss.Session
	Lang     string
}

func NewRequestContext(ctx *gin.Context) (rctx *RequestContext) {
	rctx = &RequestContext{
		Gin: ctx,
		Response: common.Response{
			Success: true,
		},
		Session: ss.Default(ctx),
		OpCtx: ccommon.OperationContext{
			Log:     getLogger(ctx),
			Runtime: getRuntime(ctx),
		},
	}
	if rctx.OpCtx.Runtime.Config == nil {
		rctx.OpCtx.Log.Panic("[Fatal] Configuration missing.")
	}
	if user, valid := rctx.Session.Get("user").(string); valid {
		rctx.OpCtx.Account.Name = user
	}
	if langSetting, valid := rctx.Session.Get("lang").(string); valid {
		rctx.Lang = langSetting
	} else {
		rctx.Lang = ParseHeaderForLocale(ctx, "en", "zh")
		if rctx.Lang == "" {
			rctx.Lang = rctx.OpCtx.Runtime.Config.DefaultLanguage
		}
		rctx.Session.Set("lang", rctx.Lang)
		rctx.Session.Save()
	}
	return rctx
}

func (ctx *RequestContext) Database() (db *gorm.DB, err error) {
	return ctx.OpCtx.Database()
}

func (ctx *RequestContext) GetAccount() *account.Account {
	return ctx.OpCtx.GetAccount()
}

func (ctx *RequestContext) RBAC() *account.RBACContext {
	return ctx.OpCtx.RBAC()
}

func (ctx *RequestContext) DatabaseOrFail() *gorm.DB {
	db, err := ctx.Database()
	if err != nil {
		ctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
		return nil
	}
	return db
}

func (ctx *RequestContext) ConfigOrFail() *config.WingConfiguration {
	if ctx.OpCtx.Runtime.Config == nil {
		ctx.AbortWithDebugMessage(http.StatusInternalServerError, common.ErrConfigMissing.Error())
		return nil
	}
	return ctx.OpCtx.Runtime.Config
}

func (ctx *RequestContext) FailWithMessage(message string) {
	ctx.Response.Message = ctx.TranslateMessage(message)
	ctx.Response.Success = false
	ctx.Gin.JSON(http.StatusOK, ctx.Response)
}

func (ctx *RequestContext) SucceedWithMessage(message string) {
	ctx.Response.Message = ctx.TranslateMessage(message)
	ctx.Response.Success = true
	ctx.Gin.JSON(http.StatusOK, ctx.Response)
}

func (ctx *RequestContext) Succeed() {
	ctx.SucceedWithMessage("Succeed")
}

func (ctx *RequestContext) TranslateMessage(message string, args ...interface{}) string {
	return common.TranslateMessage(ctx.Lang, message, args...)
}

func (ctx *RequestContext) GetLocaleLanguage() language.Tag {
	return message.MatchLanguage(ctx.Lang)
}

func (ctx *RequestContext) GetDebugMessageForResponse(message string) string {
	debugMode := ctx.OpCtx.Runtime.Config == nil || ctx.OpCtx.Runtime.Config.Debug
	if !debugMode {
		message = fmt.Sprintf("internal server error. [request id: %v]", ctx.OpCtx.Log.Data["request_id"])
	} else {
		message = message + " [request id:" + ctx.OpCtx.Log.Data["request_id"].(string) + "]"
	}
	return message
}

func (ctx *RequestContext) AbortWithDebugMessage(code int, message string) {
	message = ctx.GetDebugMessageForResponse(message)
	ctx.Response.Message = message
	ctx.Response.Success = false
	ctx.Gin.JSON(code, ctx.Response)
}

func (ctx *RequestContext) AbortWithError(err error) {
	message := ""
	if _, isExternal := err.(common.ExternalError); !isExternal {
		message = ctx.GetDebugMessageForResponse(err.Error())
		ctx.FailWithMessage(message)
	} else {
		ctx.FailWithMessage(err.Error())
	}
	ctx.OpCtx.Log.Error("error: " + err.Error())
}

func (ctx *RequestContext) LoginEnsured(fail bool) bool {
	if ctx.OpCtx.Account.Name != "" {
		return true
	}
	if fail {
		ctx.Response.Data = common.RedirectResponse{Next: "/"}
		ctx.FailWithMessage("Auth.Unauthenticated")
	}
	return false
}

func (ctx *RequestContext) RBACOrDeny() *account.RBACContext {
	rbac := ctx.RBAC()
	if rbac != nil {
		ctx.FailWithMessage("Auth.LackOfPermissing")
		return nil
	}
	return rbac
}

func (ctx *RequestContext) PermitOrReject(resource string, verbs int64) bool {
	if !ctx.OpCtx.Permitted(resource, verbs) {
		ctx.FailWithMessage("Auth.LackOfPermission")
		return false
	}
	return true
}

func (ctx *RequestContext) BindOrFail(req APIRequest) bool {
	if err := ctx.Gin.ShouldBind(req); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ctx.FailWithMessage(err.Error())
		} else {
			ctx.FailWithMessage(FormatValidateErrorMessage(ctx, errs).Error())
		}
		return false
	}
	if err := req.Clean(ctx); err != nil {
		ctx.FailWithMessage(err.Error())
		return false
	}
	return true
}
