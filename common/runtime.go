package common

import (
	"errors"
	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"github.com/RichardKnop/machinery/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"reflect"
)

type WingRuntime struct {
	Config        *config.WingConfiguration
	ClusterConfig *rest.Config
	MachineID     string
	JobServer     *machinery.Server
}

func (w *WingRuntime) RegisterTask(name string, taskProc interface{}) error {
	if name == "" {
		return errors.New("Task name cannot be blank.")
	}
	if w.JobServer == nil {
		return ErrRuntimeNotFullyInited
	}
	ty := reflect.TypeOf(taskProc)
	if ty.Kind() != reflect.Func {
		return errors.New("taskProc is not an function. got " + ty.String() + ".")
	}
	log.Info(reflect.TypeOf(taskProc))
	if ty.NumOut() == 1 && ty.NumIn() == 1 &&
		ty.Out(0).Kind() == reflect.Func && ty.In(0) == reflect.TypeOf(w) {
		// TaskFactory
		taskProc = reflect.ValueOf(taskProc).Call([]reflect.Value{
			reflect.ValueOf(w),
		})[0].Interface()
	}
	log.Info(reflect.TypeOf(taskProc))
	return w.JobServer.RegisterTask(name, taskProc)
}
