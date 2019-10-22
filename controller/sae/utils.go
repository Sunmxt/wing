package sae

import (
	"errors"
	"fmt"
	"strconv"

	"git.stuhome.com/Sunmxt/wing/common"
	ccommon "git.stuhome.com/Sunmxt/wing/controller/common"
	"git.stuhome.com/Sunmxt/wing/controller/sae/operator"
	"git.stuhome.com/Sunmxt/wing/controller/sae/operator/kubernetes"
	"git.stuhome.com/Sunmxt/wing/model/sae"
	"git.stuhome.com/Sunmxt/wing/model/scm"

	"github.com/jinzhu/gorm"
)

// LoadOrchestratorOperator loads cluster operator by orchestrator ID.
func LoadOrchestratorOperator(ctx *ccommon.OperationContext, orchestratorID int) (operator.Operator, error) {
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
func GetClusterOperatorByOrchestratorID(ctx *ccommon.OperationContext, orchestratorID int) (oper operator.Operator, err error) {
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

func GetActiveDeployment(ctx *ccommon.OperationContext, cluster *sae.ApplicationCluster) (*sae.ApplicationDeployment, error) {
	db, err := ctx.Database()
	if err != nil {
		return nil, err
	}
	deployment := &sae.ApplicationDeployment{}
	if err = db.Where("cluster_id = (?) and state in (?)", cluster.Basic.ID, sae.ActiveDeploymentStates).Order("create_time desc").First(deployment).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return deployment, nil
}

func CreateProductSnapshotForApplication(ctx *ccommon.OperationContext, tx *gorm.DB, app *sae.Application) (reqs []sae.ProductRequirement, err error) {
	autoCommit := false
	if app == nil {
		err = errors.New("application missing.")
		return nil, err
	}
	defer func() {
		if autoCommit {
			if err == nil {
				err = tx.Commit().Error
			}
			if err != nil {
				tx.Rollback()
			}
		}
	}()
	if tx == nil {
		if tx, err = ctx.Database(); err != nil {
			return nil, err
		}
		tx = tx.Begin()
		autoCommit = true
	}
	var buildIDs []int
	if err = tx.Where("application_id = (?)", app.ID).
		Model(&sae.BuildDependency{}).
		Pluck("build_id", &buildIDs).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	buildIDs = common.NewIntSet(buildIDs...).List()
	var products []scm.CIRepositoryBuildProduct
	if len(buildIDs) < 1 {
		return nil, nil
	}
	// get the latest product for each build.
	if err = tx.Where("build_id in (?)", buildIDs).
		Order("id desc, create_time desc").
		Limit(len(buildIDs)).
		Find(&products).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	latestProducts := map[int]*scm.CIRepositoryBuildProduct{}
	for _, product := range products {
		latestProducts[product.BuildID] = &product
	}
	missingIDs := make([]int, 0)
	for _, buildID := range buildIDs {
		_, ok := latestProducts[buildID]
		if !ok {
			missingIDs = append(missingIDs, buildID)
		}
	}
	if len(missingIDs) > 0 {
		return nil, fmt.Errorf("products missing for build id %v.", missingIDs)
	}
	reqs = make([]sae.ProductRequirement, len(buildIDs))
	for idx, buildID := range buildIDs {
		ref, extra, product := &reqs[idx], &scm.BuildProductExtra{}, latestProducts[buildID]
		if err := product.DecodeExtra(extra); err != nil {
			return nil, err
		}
		ref.Namespace = extra.Namespace
		ref.Environment = extra.Environment
		ref.Tag = extra.Tag
	}
	return reqs, nil
}

func OverrideProductRequirement(dst, src []sae.ProductRequirement) []sae.ProductRequirement {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = make([]sae.ProductRequirement, len(src), 0)
	}
	makeKey := func(ns, env string) string {
		return "ns(" + ns + ")+env(" + env + ")"
	}
	mapIdentReq := map[string]*sae.ProductRequirement{}
	for _, req := range dst {
		key := makeKey(req.Namespace, req.Environment)
		mapIdentReq[key] = &req
	}
	for _, req := range src {
		key := makeKey(req.Namespace, req.Environment)
		cur, ok := mapIdentReq[key]
		if !ok {
			dst = append(dst, req)
		} else {
			cur.Tag = req.Tag
		}
	}
	return dst
}
