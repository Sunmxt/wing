package sae

import (
	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	csae "git.stuhome.com/Sunmxt/wing/controller/sae"
	"git.stuhome.com/Sunmxt/wing/model/account"
	mcommon "git.stuhome.com/Sunmxt/wing/model/common"
	"git.stuhome.com/Sunmxt/wing/model/sae"
	"github.com/gin-gonic/gin"
)

// CreateApplicationClusterRequest : request from to create application cluster.
type CreateApplicationClusterRequest struct {
	OrcherstratorID int `form:"orchestrator_id" binding:"required"`
	ApplicationID   int `form:"application_id" binding:"required"`

	// Command tell how to launch this application
	Command string `form:"command"`
	// Replicas specifies the number of instances to deploy.
	Replicas int `form:"replicas" binding:"required"`

	// Environment variables.
	EnvironmentVariables map[string]string `form:"env_vars"`

	// Number of required CPU Cores.
	Core float32 `form:"core" binding:"required"`

	// Number (in bytes) of required memory.
	Memory uint64 `form:"memory" binding:"required"`

	// reference to base image.
	BaseImage string `form:"base_image" binding:"required"`
}

// Clean validate the form.
func (c *CreateApplicationClusterRequest) Clean(ctx *acommon.RequestContext) error {
	return nil
}

// CreateApplicationClusterResponse contains JSON response for new created applcation cluster.
type CreateApplicationClusterResponse struct {
	ClusterID    int `json:"cluster_id"`
	DeploymentID int `json:"deployment_id"`
}

// CreateApplicationCluster : create new cluster for application.
func CreateApplicationCluster(ctx *gin.Context) {
	rctx, request, response := acommon.NewRequestContext(ctx), &CreateApplicationClusterRequest{}, &CreateApplicationClusterResponse{}
	if !rctx.LoginEnsured(true) || !rctx.BindOrFail(request) {
		return
	}
	db := rctx.DatabaseOrFail()

	// Should have create permission for application.
	if db == nil || !rctx.PermitOrReject("application", account.VerbCreate) {
		return
	}
	account := rctx.GetAccount()
	tx := db.Begin()
	orcher, app := &sae.Orchestrator{}, &sae.Application{}

	// Application should exists.
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

	// Orchestrator should exists.
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

	// save cluster.
	cluster := &sae.ApplicationCluster{
		ApplicationID:  app.ID,
		OrchestratorID: orcher.Basic.ID,
		Active:         mcommon.Active,
		OwnerID:        account.ID,
	}
	if err := tx.Save(cluster).Error; err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	cluster.Application = app
	// get latest products according to dependencies.
	products, err := csae.CreateProductSnapshotForApplication(&rctx.OpCtx, tx, cluster.Application)
	if err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
	}
	spec := &sae.ClusterSpecificationDetail{
		Command:              request.Command,
		ReplicaCount:         request.Replicas,
		EnvironmentVariables: request.EnvironmentVariables,
		Resource: &sae.ResourceRequirement{
			Core:   request.Core,
			Memory: request.Memory,
		},
		Product:   products,
		BaseImage: request.BaseImage,
	}
	deployment, err := csae.CreateNewDeploymentForCluster(&rctx.OpCtx, tx, cluster, spec, nil, account.ID)
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
