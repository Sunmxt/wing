package cicd

import (
	"git.stuhome.com/Sunmxt/wing/common"
	ccommon "git.stuhome.com/Sunmxt/wing/controller/common"
	"git.stuhome.com/Sunmxt/wing/model/scm"
	"git.stuhome.com/Sunmxt/wing/model/scm/gitlab"

	"github.com/jinzhu/gorm"

	"io"
	"net/url"
	"strconv"
)

func GenerateGitlabCIJobsForRepository(ctx *ccommon.OperationContext, repositoryID uint) (map[string]*gitlab.CIJob, error) {
	db, err := ctx.Database()
	if err != nil {
		return nil, err
	}
	var externalURL *url.URL
	if externalURL, err = url.Parse(ctx.Runtime.Config.ExternalURL); err != nil {
		ctx.Log.Error("invalid external url: " + ctx.Runtime.Config.ExternalURL)
		return nil, err
	}
	var builds []scm.CIRepositoryBuild
	if err = db.Where("repository_id = (?) and exec_type = (?) and active = (?)", repositoryID, scm.GitlabCIBuild, scm.Active).Find(&builds).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	jobs := map[string]*gitlab.CIJob{}
	for _, build := range builds {
		name := "wing-dynamic-build-" + common.GenerateRandomToken()
		externalURL.Path = "api/scm/builds/" + strconv.FormatInt(int64(build.Basic.ID), 10) + "/job"
		jobURL := externalURL.String()
		externalURL.Path = "api/scm/builds/" + strconv.FormatInt(int64(build.Basic.ID), 10) + "/result/report"
		reportURL := externalURL.String()
		job := &gitlab.CIJob{
			Stage: "build",
			Script: []string{
				"ci_build wing-gitlab '" + jobURL + "' '" + reportURL + "' '" + build.ProductPath + "'",
			},
		}
		jobs[name] = job
	}
	return jobs, nil
}

func GenerateScriptForBuild(ctx *ccommon.OperationContext, w io.Writer, build *scm.CIRepositoryBuild) error {
	io.WriteString(w, "#! /usr/bin/env bash \n")
	io.WriteString(w, "# Generated by Wing. \n\n")
	io.WriteString(w, build.BuildCommand)
	return nil
}

//type JobScriptDockerRuntimeImageBuild struct {
//	ID RuntimeImageIdentifier
//}
//
//func (c *JobScriptDockerRuntimeImageBuild) ScriptWrite(w *io.Writer) {
//}
//
//type ProductBuild struct {
//	ID ProductIdentifier
//	ProductPath string
//	Registry string
//}
//
//func (c *JobScriptPackageImageBuild) ScriptWrite(w *io.Writer) {
//}
//
//
//type JobScriptLoadRuntime struct {
//	RuntimeURL string
//}
//
//func (c *JobScriptLoadRuntime) ScriptWrite(w *io.Writer)
//
//func JobScriptHeader(w *io.Writer) (int, error) {
//	return w.Write([]byte("#! /usr/bin/env bash"))
//}
//
