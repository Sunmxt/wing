package common

import (
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/RichardKnop/machinery/v1/backends/result"
)

func (ctx *OperationContext) SubmitTask(name string, args []tasks.Arg) (*result.AsyncResult, error) {
	return ctx.Runtime.JobServer.SendTask(&tasks.Signature{
		Name: name,
		Args: args,
	})
}