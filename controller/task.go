package controller

import (
	"git.stuhome.com/Sunmxt/wing/cmd/runtime"
	"git.stuhome.com/Sunmxt/wing/controller/cicd"
	"git.stuhome.com/Sunmxt/wing/controller/sae"
)

func RegisterTasks(runtime *runtime.WingRuntime) (err error) {
	if err = cicd.RegisterTasks(runtime); err != nil {
		return err
	}
	if err = sae.RegisterTasks(runtime); err != nil {
		return err
	}
	return
}
