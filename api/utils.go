package api

import (
	"errors"
	"fmt"
	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"git.stuhome.com/Sunmxt/wing/common"
	mlog "git.stuhome.com/Sunmxt/wing/log"
	"git.stuhome.com/Sunmxt/wing/uac"
	ss "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"net/http"
	"strconv"
	"strings"
)

var ErrConfigMissing error = errors.New("Configuration is missing in context.")

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

func getConfig(ctx *gin.Context) (conf *config.WingConfiguration) {
	raw, ok := ctx.Get("config")
	if !ok {
		return nil
	}
	conf, _ = raw.(*config.WingConfiguration)
	return conf
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
	Log      *log.Entry
	Response common.Response
	Session  ss.Session
	Config   *config.WingConfiguration
	DB       *gorm.DB
	Lang     string

	RBACContext *uac.RBACContext
	User        string
}

func NewRequestContext(ctx *gin.Context) (rctx *RequestContext) {
	rctx = &RequestContext{
		Gin: ctx,
		Log: getLogger(ctx),
		Response: common.Response{
			Success: true,
		},
		Session: ss.Default(ctx),
		Config:  getConfig(ctx),
	}
	if rctx.Config == nil {
		rctx.Log.Panic("[Fatal] Configuration missing.")
	}
	if user, valid := rctx.Session.Get("user").(string); valid {
		rctx.User = user
	}
	if langSetting, valid := rctx.Session.Get("lang").(string); valid {
		rctx.Lang = langSetting
	} else {
		rctx.Lang = ParseHeaderForLocale(ctx, "en", "zh")
		if rctx.Lang == "" {
			rctx.Lang = rctx.Config.DefaultLanguage
		}
		rctx.Session.Set("lang", rctx.Lang)
		rctx.Session.Save()
	}
	return rctx
}

func (ctx *RequestContext) RBAC() *uac.RBACContext {
	if ctx.RBACContext != nil {
		return ctx.RBACContext
	}
	if ctx.User == "" {
		ctx.Log.Info("[RBAC] Anonymous request.")
		return nil
	}
	ctx.RBACContext = uac.NewRBACContext(ctx.User)
	db, err := ctx.Database()
	if err != nil {
		ctx.Log.Warnf("[RBAC] Cannot load RBAC rule for user \"%v\"", ctx.User)
		return nil
	}
	if err = ctx.RBACContext.Load(db); err != nil {
		ctx.Log.Warn("[RBAC] RBAC rules not loaded: " + err.Error())
		return nil
	}
	return ctx.RBACContext
}

func (ctx *RequestContext) Database() (db *gorm.DB, err error) {
	if ctx.DB != nil {
		return ctx.DB, nil
	}
	if ctx.Config == nil {
		return nil, ErrConfigMissing
	}
	if ctx.DB, err = gorm.Open(ctx.Config.DB.SQLEngine, ctx.Config.DB.SQLDsn); err != nil {
		return nil, errors.New("Failed too open database: " + err.Error())
	}
	return ctx.DB, nil
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
	if ctx.Config == nil {
		ctx.AbortWithDebugMessage(http.StatusInternalServerError, ErrConfigMissing.Error())
		return nil
	}
	return ctx.Config
}

func (ctx *RequestContext) Permitted(resource string, verbs int64) bool {
	return true
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

func (ctx *RequestContext) AbortWithDebugMessage(code int, message string) {
	debugMode := ctx.Config == nil || ctx.Config.Debug
	if !debugMode {
		ctx.Log.Error("internal error: " + message)
		message = fmt.Sprintf("internal server error. [request id: %v]", ctx.Log.Data["request_id"])
	} else {
		message = message + " [request id:" + ctx.Log.Data["request_id"].(string) + "]"
	}
	ctx.Response.Message = message
	ctx.Response.Success = false
	ctx.Gin.JSON(code, message)
}

func (ctx *RequestContext) LoginEnsured(fail bool) bool {
	if ctx.User != "" {
		return true
	}
	if fail {
		ctx.Response.Data = common.RedirectResponse{Next: "/"}
		ctx.FailWithMessage("Auth.Unauthenticated")
	}
	return false
}
