package common

import (
	"time"

	"github.com/RichardKnop/machinery/v1/backends/result"
	"github.com/RichardKnop/machinery/v1/tasks"
)

func (ctx *OperationContext) SubmitTask(name string, args []tasks.Arg, retry uint) (*result.AsyncResult, error) {
	return ctx.SubmitDelayTask(name, args, 0, retry)
}

func (ctx *OperationContext) SubmitDelayTask(name string, args []tasks.Arg, delay uint, retry uint) (*result.AsyncResult, error) {
	sign := tasks.Signature{
		Name: name,
		Args: args,
	}
	if retry > 0 {
		sign.RetryCount = int(retry)
	}
	if delay > 0 {
		eta := time.Now().UTC().Add(time.Second * 5)
		sign.ETA = &eta
	}
	return ctx.Runtime.JobServer.SendTask(&sign)
}
