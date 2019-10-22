package sae

import (
	"errors"

	ccommon "git.stuhome.com/Sunmxt/wing/controller/common"
	"git.stuhome.com/Sunmxt/wing/model/sae"
	"github.com/jinzhu/gorm"
)

func CreateNewDeploymentForCluster(ctx *ccommon.OperationContext, tx *gorm.DB, cluster *sae.ApplicationCluster, spec *sae.ClusterSpecificationDetail, requireOverrided []sae.ProductRequirement) (deployment *sae.ApplicationDeployment, err error) {
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
	}
	if err = tx.Save(deployment).Error; err != nil {
		return nil, err
	}
	return
}
