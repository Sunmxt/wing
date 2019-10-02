package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	"git.stuhome.com/Sunmxt/wing/common"

	"git.stuhome.com/Sunmxt/wing/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func GetApplicationInfo(ctx *gin.Context) {
}

type ApplicationResponse struct {
	Name         string `json:"name"`
	Owner        string `json:"owner"`
	State        int    `json:"state"`
	DeploymentID int    `json:"deployment_id"`
}

type ApplicationListResponse struct {
	Applications ApplicationResponse `json:"application"`
}

func ListApplication(ctx *gin.Context) {
}

type ApplicationCreateRequest struct {
	Name       string                 `form:"name" binding:"required"`
	Image      string                 `form:"image" binding:"required"`
	EnvVar     map[string]interface{} `form:"env_vars"`
	Replicas   int                    `form:"replicas" binding:"min=0"`
	CPUCore    float32                `form:"cpu_cores" binding:"min=0"`
	ArgsRaw    string                 `form:"args" binding:"required"`
	CommandRaw string                 `form:"command" binding:"required"`
	Memory     uint64                 `form:"memory" binding:"required"`

	Args    []string `form:"-"`
	Command []string `form:"-"`
}

func (r *ApplicationCreateRequest) Clean(ctx *acommon.RequestContext) (err error) {
	invalidFields := []string{}
	if r.Command == nil {
		r.Command = make([]string, 0)
		if err = json.Unmarshal([]byte(r.CommandRaw), &r.Command); err != nil {
			invalidFields = append(invalidFields, "Command")
		}
	}
	if r.Args == nil {
		r.Args = make([]string, 0)
		if err = json.Unmarshal([]byte(r.ArgsRaw), &r.Args); err != nil {
			invalidFields = append(invalidFields, "Args")
		}
	}
	if r.EnvVar == nil {
		invalidFields = append(invalidFields, "EnvVar")
	}
	if !common.ValidApplicationName(r.Name) {
		invalidFields = append(invalidFields, "Name")
	}
	if len(invalidFields) > 0 {
		return errors.New(ctx.TranslateMessage("Partial.InvalidFields") + ":" + strings.Join(invalidFields, ", "))
	}
	return nil
}

func (r *ApplicationCreateRequest) MarshalEnvVarJSON() (string, error) {
	bin, err := json.Marshal(r.EnvVar)
	if err != nil {
		return "", err
	}
	return string(bin), nil
}

func (r *ApplicationCreateRequest) MarshalCommandJSON() (string, error) {
	bin, err := json.Marshal(r.Command)
	if err != nil {
		return "", err
	}
	return string(bin), nil
}

func (r *ApplicationCreateRequest) MarshalArgsJSON() (string, error) {
	bin, err := json.Marshal(r.Args)
	if err != nil {
		return "", err
	}
	return string(bin), nil

}

func CreateApplication(ctx *gin.Context) {
	rctx, req, err := acommon.NewRequestContext(ctx), &ApplicationCreateRequest{}, error(nil)
	defer func() {
		if err != nil {
			rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
		}
	}()
	if !rctx.LoginEnsured(true) || !rctx.BindOrFail(req) {
		return
	}
	//if !rctx.PermitOrReject("wing/application", model.VerbCreate) {
	//	return // reject
	//}
	account := rctx.GetAccount()
	if account == nil {
		rctx.FailWithMessage("Auth.LackOfPermission")
		return
	}
	db := rctx.DatabaseOrFail()
	if db == nil {
		return
	}
	spec, app := &model.AppSpec{
		ImageRef: req.Image,
		Replica:  req.Replicas,
		Memory:   req.Memory,
		CPUCore:  req.CPUCore,
	}, &model.Application{
		Name: req.Name,
	}
	if err = db.Where(&model.Application{Name: req.Name}).First(app).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			rctx.OpCtx.Log.Error("[create application] cannot check whether application exists: " + err.Error())
			rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		rctx.FailWithMessage("UI.Operation.ApplicationFound")
		return
	}
	if spec.EnvVar, err = req.MarshalEnvVarJSON(); err != nil {
		return
	}
	if spec.Command, err = req.MarshalCommandJSON(); err != nil {
		return
	}
	if spec.Args, err = req.MarshalArgsJSON(); err != nil {
		return
	}
	if err = db.Save(spec).Error; err != nil {
		rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
		return
	}
	app.OwnerID = account.Basic.ID
	app.SpecID = spec.Basic.ID
	if err = db.Save(app).Error; err != nil {
		rctx.AbortWithDebugMessage(http.StatusInternalServerError, err.Error())
		return
	}
	rctx.Succeed()
}
