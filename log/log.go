package log

import (
	"github.com/gin-gonic/gin"
	guuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"time"
)

func RequestLogger(ctx *gin.Context) *log.Entry {
	entry := log.NewEntry(log.StandardLogger())
	request_id, ok := ctx.Get("RequestID")
	if !ok {
		request_id = guuid.NewV4().String()
	}
	entry.Data["request_id"] = request_id
	//entry.Data["path"] = ctx.Request.URL.RequestURI()
	entry.Data["module"] = "request"
	return entry
}

func RequestLogMiddleware(ctx *gin.Context) {
	ctx.Set("RequestID", guuid.NewV4().String())
	start := time.Now()

	ctx.Next()

	latency := time.Since(start)
	logger := RequestLogger(ctx)
	ctx.Set("logger", logger)

	logger.Data["cost"] = latency
	logger.Data["status"] = ctx.Writer.Status()
	//logger.Data["method"] = ctx.Request.Method
	logger.Data["remote"] = ctx.Request.RemoteAddr
	logger.Data["module"] = "api"
	logger.Infof("[request] %v %v %v", ctx.Request.Method, ctx.Request.URL.RequestURI(), ctx.Writer.Status())
}

type InfoLogger interface {
	Info(...interface{})
}

type ErrorLogger interface {
	Error(...interface{})
}

type NormalLogger interface {
	InfoLogger
	ErrorLogger
}
