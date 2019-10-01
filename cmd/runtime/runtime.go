package runtime

import (
	"errors"
	"reflect"

	"github.com/RichardKnop/machinery/v1"
	"k8s.io/client-go/rest"

	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/model/scm/gitlab"
)

type WingRuntime struct {
	Config                *config.WingConfiguration
	ClusterConfig         *rest.Config
	MachineID             string
	JobServer             *machinery.Server
	GitlabWebhookEventHub *gitlab.EventHub
}

func (w *WingRuntime) RegisterTask(name string, taskProc interface{}) error {
	if name == "" {
		return errors.New("Task name cannot be blank.")
	}
	if w.JobServer == nil {
		return common.ErrRuntimeNotFullyInited
	}
	ty := reflect.TypeOf(taskProc)
	if ty.Kind() != reflect.Func {
		return errors.New("taskProc is not an function. got " + ty.String() + ".")
	}
	if ty.NumOut() == 1 && ty.NumIn() == 1 &&
		ty.Out(0).Kind() == reflect.Func && ty.In(0) == reflect.TypeOf(w) {
		// TaskFactory
		taskProc = reflect.ValueOf(taskProc).Call([]reflect.Value{
			reflect.ValueOf(w),
		})[0].Interface()
	}
	return w.JobServer.RegisterTask(name, taskProc)
}
