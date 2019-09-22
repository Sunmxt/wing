package cicd

import (
	"git.stuhome.com/Sunmxt/wing/common"
	ccommon "git.stuhome.com/Sunmxt/wing/controller/common"
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/RichardKnop/machinery/v1/backends/result"
	log "github.com/sirupsen/logrus"
)


func RegisterTasks(runtime *common.WingRuntime, server *machinery.Server) (err error) {
	tasks := map[string]interface{}{
		"SubmitCIApproveMergeRequest": SubmitCIApprovalMergeRequest,
	}

	for name, task := range tasks {
		if err := server.RegisterTask(name, task); err != nil {
			log.Error("Register task \"" + name + "\" failure:" + err.Error())
			return err
		}
		log.Info("Register task \"" + name + "\".")
	}
	return nil
}

func AsyncSubmitCIApprovalMergeRequest(ctx *ccommon.OperationContext, platformID, repositoryID, approvalID int) (*result.AsyncResult, error) {
	return ctx.SubmitTask("SubmitCIApprovalMergeRequest", []tasks.Arg{
		{
			Type: "int",
			Value: platformID,
		},
		{
			Type: "int",
			Value: repositoryID,
		},
		{
			Type: "int",
			Value: approvalID,
		},
	})
}

func SubmitCIApprovalMergeRequest(platformID, repositoryID, approvalID int) error {
	return nil
}