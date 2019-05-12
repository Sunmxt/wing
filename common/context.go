package common

import (
    "github.com/gin-gonic/gin"
    "git.stuhome.com/Sunmxt/wing/log"
	"github.com/sirupsen/logrus"
)

type RequestContext struct {
    Gin  *gin.Context
    Log  *logrus.Entry
    Response Response
}

func NewRequestContext(ctx *gin.Context) *RequestContext {
    return &RequestContext{
        Gin: ctx,
        Log: log.RequestLogger(ctx),
        Response: Response{
            Success: true,
        },
    }
}
