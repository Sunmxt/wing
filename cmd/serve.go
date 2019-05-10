package cmd

import (
	"git.stuhome.com/Sunmxt/wing/uac"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (c *Wing) Serve() {
	c.LogConfig()
	log.Info("[bootstrap] Wing server bootstraping...")

	if !c.Debug {
		log.Info("[bootstrap] production mode.")
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	log.Info("[bootstrap] Register UAC API.")
	uac.RegisterAPI(router)

	router.Run("0.0.0.0:10089")
}
