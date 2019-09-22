package cicd

import (
	"git.stuhome.com/Sunmxt/wing/common"
	ccommon "git.stuhome.com/Sunmxt/wing/controller/common"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/RichardKnop/machinery/v1/backends/result"
	log "github.com/sirupsen/logrus"
)


func RegisterTasks(runtime *common.WingRuntime) (err error) {
	tasks := map[string]interface{}{
		"SubmitCIApprovalMergeRequest": SubmitCIApprovalMergeRequest,
	}

	for name, task := range tasks {
		if err := runtime.RegisterTask(name, task); err != nil {
			log.Error("Register task \"" + name + "\" failure:" + err.Error())
			return err
		}
		log.Info("Register task \"" + name + "\".")
	}
	return nil
}

func AsyncSubmitCIApprovalMergeRequest(ctx *ccommon.OperationContext, platformID int, repositoryID uint, approvalID int) (*result.AsyncResult, error) {
	return ctx.SubmitTask("SubmitCIApprovalMergeRequest", []tasks.Arg{
		{
			Type: "int",
			Value: platformID,
		},
		{
			Type: "uint",
			Value: repositoryID,
		},
		{
			Type: "int",
			Value: approvalID,
		},
	})
}


func SubmitCIApprovalMergeRequest(ctx *common.WingRuntime) (func (int, uint, int) error) {
	return func (platformID int, repositoryID uint, approvalID int) error {
		return nil
	}
}