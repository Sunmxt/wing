package cmd

import (
	"git.stuhome.com/Sunmxt/wing/api"
	mlog "git.stuhome.com/Sunmxt/wing/log"
	ss "github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/ioutil"

	"k8s.io/client-go/rest"
	"time"
)

func (c *Wing) Serve() {
	c.LogConfig()
	log.Info("[bootstrap] Wing server bootstraping...")

	if !c.Debug {
		log.Info("[bootstrap] Production mode.")
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DefaultWriter = ioutil.Discard

	// Log
	router := gin.Default()
	router.Use(mlog.RequestLogMiddleware)

	// Session
	store := cookie.NewStore([]byte(c.Config.SessionToken))
	router.Use(ss.Sessions("wing_session", store))

	// Wing
	router.Use(func(ctx *gin.Context) {
		ctx.Set("config", c.Config)
		ctx.Next()
	})

	go func() {
		for {
			kconf, err := rest.InClusterConfig()
			if err == nil {
				c.Config.Kube.InclusterConfig = kconf
				break
			}
			log.Error("[kubernetes] Failed to load incluster kubernetes config: " + err.Error())
			time.Sleep(5 * time.Second)
		}
		log.Info("[kubernetes] Incluster kubernetes config loaded.")
	}()

	log.Info("[bootstrap] Register API.")
	api.RegisterAPI(router)

	log.Infof("[bootstrap] Bind %v", c.Config.Bind)
	if err := router.Run(c.Config.Bind); err != nil {
		log.Infof("[bootstrap] Server error: %v", err.Error())
	}
}
