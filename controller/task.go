package controller

import (
	"git.stuhome.com/Sunmxt/wing/cmd/runtime"
	"git.stuhome.com/Sunmxt/wing/controller/cicd"
)

func RegisterTasks(runtime *runtime.WingRuntime) (err error) {
	if err = cicd.RegisterTasks(runtime); err != nil {
		return err
	}
	return
}
