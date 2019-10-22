package sae

import (
	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	csae "git.stuhome.com/Sunmxt/wing/controller/sae"
	"git.stuhome.com/Sunmxt/wing/model/account"
	mcommon "git.stuhome.com/Sunmxt/wing/model/common"
	"git.stuhome.com/Sunmxt/wing/model/sae"
	"github.com/gin-gonic/gin"
)

type CreateApplicationClusterRequest struct {
	OrcherstratorID int `form:"orchestrator_id" binding:"required"`
	ApplicationID   int `form:"application_id" binding:"required"`

	Command              string            `form:"command"`
	Replicas             int               `form:"replicas" binding:"required"`
	TestingReplicas      int               `form:"testing_replicas" binding:"required"`
	EnvironmentVariables map[string]string `form:"env_vars"`
	Core                 float32           `form:"core" binding:"required"`
	Memory               uint64            `form:"memory" binding:"required"`
	BaseImage            string            `form:"base_image" binding:"required"`
}

func (c *CreateApplicationClusterRequest) Clean(ctx *acommon.RequestContext) error {
	return nil
}

type CreateApplicationClusterResponse struct {
	ClusterID    int `json:"cluster_id"`
	DeploymentID int `json:"deployment_id"`
}

func CreateApplicationCluster(ctx *gin.Context) {
	rctx, request, response := acommon.NewRequestContext(ctx), &CreateApplicationClusterRequest{}, &CreateApplicationClusterResponse{}
	if !rctx.LoginEnsured(true) || !rctx.BindOrFail(request) {
		return
	}
	db := rctx.DatabaseOrFail()
	if db == nil || !rctx.PermitOrReject("application", account.VerbCreate) {
		return
	}
	tx := db.Begin()
	orcher, app := &sae.Orchestrator{}, &sae.Application{}
	if err := app.ByID(tx, request.ApplicationID); err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	if app.Basic.ID < 1 {
		tx.Rollback()
		rctx.FailWithMessage("SAE.ApplicationNotFound")
		return
	}
	if err := orcher.ByID(tx, request.OrcherstratorID); err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	if orcher.Basic.ID < 1 {
		tx.Rollback()
		rctx.FailWithMessage("SAE.OrchestratorNotFound")
		return
	}
	cluster := &sae.ApplicationCluster{
		ApplicationID:  app.ID,
		OrchestratorID: orcher.Basic.ID,
		Active:         mcommon.Active,
	}
	if err := tx.Save(cluster).Error; err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	cluster.Application = app
	products, err := csae.CreateProductSnapshotForApplication(&rctx.OpCtx, tx, cluster.Application)
	if err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
	}
	spec := &sae.ClusterSpecificationDetail{
		Command:              request.Command,
		ReplicaCount:         request.Replicas,
		EnvironmentVariables: request.EnvironmentVariables,
		TestingReplicaCount:  request.TestingReplicas,
		Resource: &sae.ResourceRequirement{
			Core:   request.Core,
			Memory: request.Memory,
		},
		Product:   products,
		BaseImage: request.BaseImage,
	}
	deployment, err := csae.CreateNewDeploymentForCluster(&rctx.OpCtx, tx, cluster, spec, nil)
	if err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	response.ClusterID = cluster.Basic.ID
	response.DeploymentID = deployment.Basic.ID
	rctx.Response.Data = response
	rctx.Succeed()
}
