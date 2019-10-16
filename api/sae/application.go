package sae

import (
	"errors"

	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/model/sae"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type ApplicationCreateRequest struct {
	Name        string `form:"name" binding:"required"`
	ServiceName string `form:"service_name" binding:"required"`
	Description string `form:"description"`
	Type        string `form:"type"`
}

func (r *ApplicationCreateRequest) Clean(ctx *acommon.RequestContext) (err error) {
	if !common.ValidServiceName(r.ServiceName) {
		return errors.New("Invalid service name: " + r.ServiceName)
	}
	return nil
}

type ApplicationCreateResponse struct {
	ApplicationID int `form:"application_id"`
}

func CreateApplication(ctx *gin.Context) {
	rctx, request, response := acommon.NewRequestContext(ctx), &ApplicationCreateRequest{}, &ApplicationCreateResponse{}
	if !rctx.LoginEnsured(true) || !rctx.BindOrFail(request) {
		return
	}
	db := rctx.DatabaseOrFail()
	if db == nil {
		return
	}
	user := rctx.GetAccount()
	app := &sae.Application{}
	if err := db.Where("service_name = (?)", request.ServiceName).
		Select("id").Find(app).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		rctx.AbortWithError(err)
		return
	}
	if app.Basic.ID > 0 {
		rctx.FailWithMessage("SAE.ServiceNameAlreadyExists")
		return
	}
	app.Name = request.Name
	app.ServiceName = request.ServiceName
	app.Description = request.Description
	app.OwnerID = user.Basic.ID
	if err := db.Save(app).Error; err != nil {
		rctx.AbortWithError(err)
		return
	}
	response.ApplicationID = app.Basic.ID
	rctx.Response.Data = response
	rctx.Succeed()
}
