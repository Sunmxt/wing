package controller

import (
	"git.stuhome.com/Sunmxt/wing/common"
	"github.com/RichardKnop/machinery/v1"
	"git.stuhome.com/Sunmxt/wing/controller/cicd"
)

func RegisterTasks(runtime *common.WingRuntime, server *machinery.Server) (err error) {
	if err = cicd.RegisterTasks(runtime, server); err != nil {
		return err
	}
	return
}