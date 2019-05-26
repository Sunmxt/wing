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
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

func (c *Wing) Serve() {
	c.LogConfig()
	log.Info("[bootstrap] Wing server bootstraping...")

	if !c.Debug {
		log.Info("[bootstrap] Production mode.")
		gin.SetMode(gin.ReleaseMode)
	}
	c.Runtime.Config.Debug = true
	gin.DefaultWriter = ioutil.Discard

	// Log
	router := gin.Default()
	router.Use(mlog.RequestLogMiddleware)

	// Session
	store := cookie.NewStore([]byte(c.Runtime.Config.SessionToken))
	router.Use(ss.Sessions("wing_session", store))

	// Wing
	router.Use(func(ctx *gin.Context) {
		ctx.Set("runtime", &c.Runtime)
		ctx.Next()
	})

	// Async load kubeconfig
	go func() {
		if c.Runtime.Config.Kube.KubeConfig == "" {
			log.Info("[kubernetes] Kubeconfig not specified. try to load in-cluster config.")
		} else {
			log.Info("[kubernetes] Kubeconfig: " + c.Runtime.Config.Kube.KubeConfig)
		}

		for {
			var (
				kconf *rest.Config
				err   error
			)
			if c.Runtime.Config.Kube.KubeConfig == "" {
				kconf, err = rest.InClusterConfig()
			} else {
				kconf, err = clientcmd.BuildConfigFromFlags("", c.Runtime.Config.Kube.KubeConfig)
			}
			if err == nil {
				c.Runtime.ClusterConfig = kconf
				break
			}
			log.Error("[kubernetes] Failed to load kubernetes config: " + err.Error())
			time.Sleep(10 * time.Second)
		}
		log.Info("[kubernetes] kubernetes config loaded.")
	}()

	log.Info("[bootstrap] Register API.")
	api.RegisterAPI(router)

	log.Infof("[bootstrap] Bind %v", c.Runtime.Config.Bind)
	if err := router.Run(c.Runtime.Config.Bind); err != nil {
		log.Infof("[bootstrap] Server error: %v", err.Error())
	}
}
