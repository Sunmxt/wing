package cmd

import (
	mlog "git.stuhome.com/Sunmxt/wing/log"
	"git.stuhome.com/Sunmxt/wing/uac"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

func (c *Wing) Serve() {
	c.LogConfig()
	log.Info("[bootstrap] Wing server bootstraping...")

	if !c.Debug {
		log.Info("[bootstrap] Production mode.")
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DefaultWriter = ioutil.Discard

	router := gin.Default()
	router.Use(mlog.RequestLogMiddleware)

	log.Info("[bootstrap] Register UAC API.")
	uac.RegisterAPI(router)

	log.Infof("[bootstrap] Bind %v", c.Config.Bind)
	if err := router.Run(c.Config.Bind); err != nil {
		log.Infof("[bootstrap] Server error: %v", err.Error())
	}
}
