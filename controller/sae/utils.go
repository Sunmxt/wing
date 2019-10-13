package sae

import (
	"git.stuhome.com/Sunmxt/wing/controller/common"
	"git.stuhome.com/Sunmxt/wing/controller/sae/operator"
	"git.stuhome.com/Sunmxt/wing/controller/sae/operator/kubernetes"
	"git.stuhome.com/Sunmxt/wing/model/sae"

	"github.com/jinzhu/gorm"
)

func LoadOrchestratorOperator(ctx *common.OperationContext, orchestratorID int) (operator.Operator, error) {
	db, err := ctx.Database()
	if err != nil {
		ctx.Log.Info(err.Error())
		return nil, err
	}
	orchor := sae.Orchestrator{}
	if err = orchor.ByID(db, orchestratorID); err != nil {
		ctx.Log.Error(err.Error())
		return nil, err
	}
	if orchor.Basic.ID < 1 {
		ctx.Log.Error("orchestrator not found: %v", orchestratorID)
		return nil, gorm.ErrRecordNotFound
	}

	var oper operator.Operator
	switch orchor.Type {
	case sae.Kubernetes:
		oper, err = kubernetes.NewKubernetesOperator(orchor.KubeconfigGetter())
	case sae.KubernetesIncluster:
		oper, err = kubernetes.NewKubernetesOperator(kubernetes.Incluster)
	}
	if err != nil {
		ctx.Log.Error(err.Error())
		return nil, err
	}
	return oper, nil
}

func GetClusterOperatorByOrchestratorID(ctx *common.OperationContext, orchestratorID int) (oper operator.Operator, err error) {
	raw, exists := ctx.Runtime.ClusterOperator.Load(orchestratorID)
	if exists {
		oper, exists = raw.(operator.Operator)
	}
	if oper == nil || !exists {
		oper, err = LoadOrchestratorOperator(ctx, orchestratorID)
		if err != nil {
			return nil, err
		}
		ctx.Runtime.ClusterOperator.Store(orchestratorID, oper)
	}
	return oper, nil
}
