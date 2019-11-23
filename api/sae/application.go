package sae

import (
	"errors"

	acommon "git.stuhome.com/Sunmxt/wing/api/common"
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/model/sae"
	"git.stuhome.com/Sunmxt/wing/model/scm"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type ApplicationCreateRequest struct {
	Name        string `form:"name" binding:"required"`
	ServiceName string `form:"service_name" binding:"required"`
	Description string `form:"description"`
	BuildIDs    []int  `form:"build_ids"`
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

// CreateApplication : create new application.
func CreateApplication(ctx *gin.Context) {
	rctx, request, response := acommon.NewRequestContext(ctx), &ApplicationCreateRequest{}, &ApplicationCreateResponse{}
	if !rctx.LoginEnsured(true) || !rctx.BindOrFail(request) {
		return
	}
	db := rctx.DatabaseOrFail()
	if db == nil {
		return
	}
	tx := db.Begin()
	user := rctx.GetAccount()
	app := &sae.Application{}

	// No duplicated service name of application.
	if err := tx.Where("service_name = (?)", request.ServiceName).
		Select("id").Find(app).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}
	if app.Basic.ID > 0 {
		tx.Rollback()
		rctx.FailWithMessage("SAE.ServiceNameAlreadyExists")
		return
	}

	// All application dependencies should exist.
	depCount := 0
	if len(request.BuildIDs) > 0 {
		if err := tx.Where("id in (?)", request.BuildIDs).Model(&scm.CIRepositoryBuild{}).Count(&depCount).Error; err != nil {
			tx.Rollback()
			rctx.AbortWithError(err)
			return
		}
		if depCount != len(request.BuildIDs) {
			tx.Rollback()
			rctx.OpCtx.Log.Infof("dependency count not matched: %v != %v", depCount, len(request.BuildIDs))
			rctx.FailWithMessage("SCM.BuildNotFound")
			return
		}
	}
	app.Name = request.Name
	app.ServiceName = request.ServiceName
	app.Description = request.Description
	app.OwnerID = user.Basic.ID

	// Create new application.
	if err := tx.Save(app).Error; err != nil {
		tx.Rollback()
		rctx.AbortWithError(err)
		return
	}

	// Append build dependencies.
	if depCount > 0 {
		for _, depID := range request.BuildIDs {
			ref := &sae.BuildDependency{
				BuildID:       depID,
				ApplicationID: app.ID,
			}
			if err := tx.Save(ref).Error; err != nil {
				rctx.AbortWithError(err)
				tx.Rollback()
				return
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		rctx.AbortWithError(err)
		tx.Rollback()
		return
	}
	response.ApplicationID = app.Basic.ID
	rctx.Response.Data = response
	rctx.Succeed()
}
