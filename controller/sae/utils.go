package sae

import (
	"errors"
	"strconv"

	"git.stuhome.com/Sunmxt/wing/controller/common"
	"git.stuhome.com/Sunmxt/wing/controller/sae/operator"
	"git.stuhome.com/Sunmxt/wing/controller/sae/operator/kubernetes"
	"git.stuhome.com/Sunmxt/wing/model/sae"

	"github.com/jinzhu/gorm"
)

// LoadOrchestratorOperator loads cluster operator by orchestrator ID.
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
	case sae.Kubernetes, sae.KubernetesIncluster:
		var kubeOper *kubernetes.Operator
		config := &sae.KubernetesOrchestrator{}
		if err = orchor.DecodeExtra(config); err != nil {
			return nil, err
		}
		if orchor.Type == sae.KubernetesIncluster {
			kubeOper, err = kubernetes.NewKubernetesOperator(kubernetes.Incluster)
		} else {
			kubeOper, err = kubernetes.NewKubernetesOperator(orchor.KubeconfigGetter())
		}
		oper = kubeOper.Namespace(config.Namespace)
	default:
		err = errors.New("Unknown orchestrator type: " + strconv.FormatInt(int64(orchor.Type), 10))
	}
	if err != nil {
		ctx.Log.Error(err.Error())
		return nil, err
	}
	return oper, nil
}

// GetClusterOperatorByOrchestratorID gets loaded cluster operator or loads cluster operator by orchestrator ID.
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
