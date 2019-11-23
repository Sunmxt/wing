package sae

import (
	"errors"

	ccommon "git.stuhome.com/Sunmxt/wing/controller/common"
	"git.stuhome.com/Sunmxt/wing/model/sae"
	"github.com/jinzhu/gorm"
)

func DeploymentTriggerImageBuild(ctx *ccommon.OperationContext, tx *gorm.DB, deployment *sae.ApplicationDeployment) error {

}

func AbortDeployment(ctx *ccommon.OperationContext, tx *gorm.DB, deployment *sae.ApplicationDeployment) error {

}

func DeploymentTriggerNext(ctx *ccommon.OperationContext, tx *gorm.DB, deployment *sae.ApplicationDeployment) error {
	orcher, err := LoadOrchestratorOperator(ctx, deployment.Cluster.OrchestratorID)
	if err != nil {
		return err
	}
	shouldCommit := false
	defer func() {
		if shouldCommit {
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
			return err
		}
		tx = tx.Begin()
		shouldCommit = true
	}
	var synced bool
	var nextState int

	switch deployment.State {
	case sae.DeploymentCreated:
		if err = tx.Model(deployment).Update("state", sae.DeploymentImageBuildInProgress).Error; err != nil {
			return err
		}
		if err = DeploymentTriggerImageBuild(ctx, tx, deployment); err != nil {
			return err
		}
		return nil

	case sae.DeploymentImageBuildInProgress:
		return nil
	case sae.DeploymentImageBuildFinished:
		nextState = sae.DeploymentTestingReplicaInProgress
	case sae.DeploymentTestingReplicaInProgress:
		nextState = sae.DeploymentTestingReplicaFinished
	case sae.DeploymentTestingReplicaFinished:
		nextState = sae.DeploymentInProgress
	case sae.DeploymentInProgress:
		nextState = sae.DeploymentFinished
	}
	if synced, err = orcher.Synchronize(deployment, sae.DeploymentImageBuildInProgress); err != nil {
		return err
	}
}

func CreateNewDeploymentForCluster(ctx *ccommon.OperationContext, tx *gorm.DB, cluster *sae.ApplicationCluster, spec *sae.ClusterSpecificationDetail, requireOverrided []sae.ProductRequirement, OwnerID int) (deployment *sae.ApplicationDeployment, err error) {
	shouldCommit := false
	defer func() {
		if shouldCommit {
			if err == nil {
				err = tx.Commit().Error
			}
			if err != nil {
				tx.Rollback()
			}
		}
	}()
	if deployment, err = GetActiveDeployment(ctx, cluster); err != nil {
		return nil, err
	}
	if deployment != nil && deployment.ID > 0 {
		return deployment, nil
	}
	if spec == nil {
		err = errors.New("application specification missing.")
		return nil, err
	}
	if cluster == nil {
		err = errors.New("cluster missing.")
		return nil, err
	}
	if tx == nil {
		if tx, err = ctx.Database(); err != nil {
			return nil, err
		}
		tx = tx.Begin()
		shouldCommit = true
	}
	if spec.Product == nil {
		if spec.Product, err = CreateProductSnapshotForApplication(ctx, tx, cluster.Application); err != nil {
			return nil, err
		}
	}
	spec.Product = OverrideProductRequirement(spec.Product, requireOverrided)
	rawSpec := &sae.ClusterSpecification{}
	if err = rawSpec.UpdateSpecification(spec); err != nil {
		return nil, err
	}
	if err = tx.Save(rawSpec).Error; err != nil {
		return nil, err
	}
	deployment = &sae.ApplicationDeployment{
		OldSpecificationID: cluster.SpecificationID,
		NewSpecificationID: rawSpec.Basic.ID,
		ClusterID:          cluster.ID,
		State:              sae.DeploymentCreated,
		OwnerID:            OwnerID,
	}
	if err = tx.Save(deployment).Error; err != nil {
		return nil, err
	}
	return
}
