package common

import (
	"github.com/RichardKnop/machinery/v1/backends/result"
	"github.com/RichardKnop/machinery/v1/tasks"
)

func (ctx *OperationContext) SubmitTask(name string, args []tasks.Arg, retry uint) (*result.AsyncResult, error) {
	sign := tasks.Signature{
		Name: name,
		Args: args,
	}
	if retry > 0 {
		sign.RetryCount = int(retry)
	}
	return ctx.Runtime.JobServer.SendTask(&sign)
}
