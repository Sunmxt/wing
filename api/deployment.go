package api

import (
	"encoding/json"
	"net/http"

	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/model"
	mcommon "git.stuhome.com/Sunmxt/wing/model/common"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func ListDeployment(ctx *gin.Context) {
	rctx := acommon.NewRequestContext(ctx)
	rctx.FailWithMessage("Not implemented.")
}

type CreateDeploymentResponse struct {
	DeploymentID int `json:"deployment_id"`
}

type CreateDeploymentRequest struct {
	Application string `form:"application" binding:"required"`
}

func (r *CreateDeploymentRequest) Clean(req *acommon.RequestContext) error {
	return nil
}

func CreateDeployment(ctx *gin.Context) {
	rctx, req := acommon.NewRequestContext(ctx), &CreateDeploymentRequest{}
	if !rctx.LoginEnsured(true) || !rctx.BindOrFail(req) {
		return
	}
	db := rctx.DatabaseOrFail()
	if db == nil {
		return
	}
	app := &model.Application{Name: req.Application}
	err := db.Where(app).First(&app).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.FailWithMessage("UI.Operation.ApplicationNotFound")
		} else {
			rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
		}
		return
	}

	var deploy *model.Deployment
	if deploy, _, err = rctx.OpCtx.GetCurrentDeployment(app.Basic.ID); err != nil {
		rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
		return
	}
	if deploy != nil {
		rctx.Response.Data = &CreateDeploymentResponse{
			DeploymentID: deploy.Basic.ID,
		}
		rctx.FailWithMessage("UI.Operation.ExistingDeploymentRunning")
		return
	}

	deploy = &model.Deployment{
		SpecID: app.SpecID,
		AppID:  app.Basic.ID,
		State:  model.Waiting,
	}
	if err := db.Save(deploy).Error; err != nil {
		rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
		return
	}
	rctx.Response.Data = &CreateDeploymentResponse{
		DeploymentID: deploy.Basic.ID,
	}
	rctx.Succeed()
	return
}

type SingleDeploymentRequest struct {
	DeploymentID int `form:"deployment_id" binding:"min=1,required"`
}

func (r *SingleDeploymentRequest) Clean(ctx *acommon.RequestContext) error {
	return nil
}

func StartDeployment(ctx *gin.Context) {
	rctx, req, synced := acommon.NewRequestContext(ctx), &SingleDeploymentRequest{}, false
	if !rctx.LoginEnsured(true) || !rctx.BindOrFail(req) {
		return
	}
	deploy := &model.Deployment{
		Basic: mcommon.Basic{
			ID: req.DeploymentID,
		},
	}
	db := rctx.DatabaseOrFail()
	if db == nil {
		return
	}
	err := db.Where(deploy).Select("ID").First(deploy).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.FailWithMessage("UI.Operation.DeploymentNotFound")
		} else {
			rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
		}
		return
	}
	deploy, synced, err = rctx.OpCtx.SyncDeployment(deploy.Basic.ID, model.Executed)
	if err != nil {
		rctx.OpCtx.Log.Error("Failed to sync deployment: " + err.Error())
		rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
		return
	}
	if !synced {
		switch deploy.State {
		case model.Executed:
			rctx.FailWithMessage("UI.Operation.Deployment.Start.AlreadyStarted")
			return
		case model.Finished:
			rctx.FailWithMessage("UI.Operation.Deployment.Start.AlreadyFinished")
			return
		case model.Terminated:
			rctx.FailWithMessage("UI.Operation.Deployment.Start.AlreadyTerminated")
			return
		case model.Terminating:
			rctx.FailWithMessage("UI.Operation.Deployment.Start.AlreadyTerminating")
			return
		}
	}
	rctx.Succeed()
}

type ApplicationSpecificationInfo struct {
	Image       string                 `json:"image"`
	Environment map[string]interface{} `json:"environment"`
	CPUCore     uint                   `json:"cpu_core"`
	Memory      uint64                 `json:"memory"`
	Command     []string               `json:"command"`
	Args        []string               `json:"args"`
}

func (m *ApplicationSpecificationInfo) FromModel(spec *model.AppSpec) (err error) {
	m.Image = spec.ImageRef
	m.CPUCore = uint(spec.CPUCore)
	m.Memory = spec.Memory
	if err = json.Unmarshal([]byte(spec.Command), &m.Command); err != nil {
		return
	}
	if err = json.Unmarshal([]byte(spec.Args), &m.Args); err != nil {
		return
	}
	if err = json.Unmarshal([]byte(spec.EnvVar), &m.Environment); err != nil {
		return
	}
	return
}

type DeploymentInfoResponse struct {
	Application     string                       `json:"application"`
	Spec            ApplicationSpecificationInfo `json:"spec"`
	TargetReplicas  int                          `json:"target_replicas"`
	Avaliables      uint                         `json:"avaliables"`
	TerminatingPods []common.ApplicationPodInfo  `json:"terminating_pods"`
	Pods            []common.ApplicationPodInfo  `json:"pods"`
	Status          int                          `json:"status"`
	Current         uint                         `json:"current"`
	Total           uint                         `json:"total"`
}

func (m *DeploymentInfoResponse) FromSelf() {
	m.Current = 0
	for _, pod := range m.Pods {
		if pod.State == common.PodReady {
			m.Current += 1
		}
	}
	m.Total = uint(m.TargetReplicas)
}

func GetDeploymentInfo(ctx *gin.Context) {
	rctx, req, resp := acommon.NewRequestContext(ctx), &SingleDeploymentRequest{}, &DeploymentInfoResponse{}
	if !rctx.LoginEnsured(true) || !rctx.BindOrFail(req) {
		return
	}
	deploy := &model.Deployment{
		Basic: mcommon.Basic{
			ID: req.DeploymentID,
		},
	}
	db := rctx.DatabaseOrFail()
	if db == nil {
		return
	}
	if err := db.Where(deploy).Preload("Spec").Preload("App").First(deploy).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			rctx.FailWithMessage("UI.Operation.DeploymentNotFound")
		} else {
			rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
		}
		return
	}
	dp, err := rctx.OpCtx.GetKubeDeploymentManifest(deploy.App.Name, deploy.Basic.ID)
	defer func() {
		if err != nil {
			rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
		}
	}()
	if err != nil {
		return
	}
	if resp.Status, err = rctx.OpCtx.SyncDeploymentState(deploy, dp); err != nil {
		return
	}
	if resp.Pods, err = rctx.OpCtx.ListApplicationPodInfo(deploy.App.Name, deploy.Basic.ID); err != nil {
		return
	}
	if resp.TerminatingPods, err = rctx.OpCtx.ListApplicationPodInfo(deploy.App.Name, deploy.Basic.ID-1); err != nil {
		return
	}
	resp.Application = deploy.App.Name
	resp.Spec.FromModel(deploy.Spec)
	resp.TargetReplicas = deploy.Spec.Replica
	resp.FromSelf()
	rctx.Response.Data = resp
	rctx.Succeed()
}

func StopDeployment(ctx *gin.Context) {
}
