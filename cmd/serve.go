package cmd

import (
	"io/ioutil"
	"time"

	"git.stuhome.com/Sunmxt/wing/api"
	"git.stuhome.com/Sunmxt/wing/api/scm"
	mlog "git.stuhome.com/Sunmxt/wing/log"

	ss "github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func (c *Wing) Serve() {
	c.LogConfig()
	log.Info("[bootstrap] Wing server bootstrapping...")

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
	store := cookie.NewStore([]byte(c.Runtime.Config.Session.Token))
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
	if err := api.RegisterAPI(router); err != nil {
		log.Error("[bootstrap] Register API Failure: " + err.Error())
		return
	}
	if err := scm.RegisterGitlabWebhookWatcher(&c.Runtime); err != nil {
		log.Error("[bootstrap] Register Gitlab webhook watcher Failure: " + err.Error())
		return
	}

	log.Infof("[bootstrap] Bind %v", c.Runtime.Config.Bind)
	if err := router.Run(c.Runtime.Config.Bind); err != nil {
		log.Infof("[bootstrap] Server error: %v", err.Error())
	}
}
