package controller

import (
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/controller/cicd"
)

func RegisterTasks(runtime *common.WingRuntime) (err error) {
	if err = cicd.RegisterTasks(runtime); err != nil {
		return err
	}
	return
}