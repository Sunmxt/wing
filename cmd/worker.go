package cmd

import (
	log "github.com/sirupsen/logrus"
)

func (c *Wing) Worker() {
	c.LogConfig()
	log.Info("[bootstrap] Wing worker bootstrapping.")

	worker := c.Runtime.JobServer.NewWorker("wing_worker_"+c.Runtime.MachineID, c.Runtime.Config.Session.Job.Concurrency)
	if err := worker.Launch(); err != nil {
		log.Error("[Worker] worker launches with failure: " + err.Error())
		return
	}
}
