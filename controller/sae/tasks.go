package sae

import (
	"git.stuhome.com/Sunmxt/wing/cmd/runtime"
	ccommon "git.stuhome.com/Sunmxt/wing/controller/common"
	"git.stuhome.com/Sunmxt/wing/controller/sae/operator"
	"git.stuhome.com/Sunmxt/wing/model/sae"
	"github.com/RichardKnop/machinery/v1/backends/result"
	"github.com/RichardKnop/machinery/v1/tasks"

	log "github.com/sirupsen/logrus"
)

func RegisterTasks(runtime *runtime.WingRuntime) (err error) {
	tasks := map[string]interface{}{
		"SynchronizeKubernetesDeployment": SynchronizeKubernetesDeployment,
	}
	for name, task := range tasks {
		if err := runtime.RegisterTask(name, task); err != nil {
			log.Errorf("Register task \"%v\" failure: " + err.Error())
			return err
		}
		log.Infof("Register task \"%v\".", name)
	}
	return nil
}

func AsyncSynchronizeKubernetesDeployment(ctx *ccommon.OperationContext, deploymentID, targetState int, retry, delay uint) (*result.AsyncResult, error) {
	return ctx.SubmitDelayTask("SynchronizeKubernetesDeployment", []tasks.Arg{
		{Type: "int", Value: deploymentID},
		{Type: "int", Value: targetState},
	}, delay, retry)
}

func SynchronizeKubernetesDeployment(runtime *runtime.WingRuntime) interface{} {
	return func(deploymentID, targetState int) error {
		ctx := ccommon.NewOperationContext(runtime)
		ctx.Log.Data["task"] = "SynchronizeKubernetesDeployment"
		db, err := ctx.Database()
		if err != nil {
			ctx.Log.Info(err.Error())
			return err
		}
		deploy, tx := &sae.ApplicationDeployment{}, db.Begin()
		defer func() {
			if err != nil {
				tx.Rollback()
			}
		}()
		if err = deploy.ByID(db.Preload("OldSpecification").
			Preload("NewSpecification").Preload("Cluster").Preload("Orchestrator"), deploymentID); err != nil {
			return err
		}
		if deploy.Basic.ID < 1 {
			ctx.Log.Errorf("deployment not found: %v", deploymentID)
			return nil
		}
		var oper operator.Operator
		if oper, err = GetClusterOperatorByOrchestratorID(ctx, deploy.Cluster.Orchestrator.Basic.ID); err != nil {
			ctx.Log.Error(err.Error())
			return err
		}
		var updated bool
		if updated, err = oper.Synchronize(deploy, targetState); err != nil {
			ctx.Log.Error("synchronize error: ", err.Error())
			return err
		}
		if updated {
			if err = tx.Save(deploy).Error; err != nil {
				ctx.Log.Error("save error:", err.Error())
				return err
			}
		}
		if err = tx.Commit().Error; err != nil {
			ctx.Log.Error("commit error:" + err.Error())
			return err
		}
		if deploy.State != targetState {
			if _, err := AsyncSynchronizeKubernetesDeployment(ctx, deploymentID, targetState, 10, 10); err != nil {
				ctx.Log.Errorf("publish next task failure: %v", deploymentID)
				return err
			}
		}
		return err
	}
}
