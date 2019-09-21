package cicd

import (
	ccommon "git.stuhome.com/Sunmxt/wing/controller/common"
	"git.stuhome.com/Sunmxt/wing/common"
	"strconv"
	"git.stuhome.com/Sunmxt/wing/model/scm/gitlab"
	"git.stuhome.com/Sunmxt/wing/model/scm"
)

func SubmitGitlabRepositoryCIApproval(ctx *ccommon.OperationContext, platform *scm.SCMPlatform, ownerID uint, repositoryID string) (*scm.CIRepositoryApproval ,error) {
	repoID, err := strconv.ParseUint(repositoryID, 10, 64)
	if err != nil {
		return nil, common.ErrInvalidRepositoryID
	}
	var client *gitlab.GitlabClient
	if client, err = platform.GitlabClient(ctx.Log); err != nil {
		return nil, err
	}
	query := client.ProjectQuery()
	project := query.Single(uint(repoID))
	if query.Error != nil {
		return nil, err
	}
	if project == nil {
		return nil, common.ErrInvalidRepositoryID
	}

	// Submit
	db, err := ctx.Database()
	if err != nil {
		return nil, err
	}
	approval := &scm.CIRepositoryApproval{
		Type: scm.GitlabMergeRequestApproval,
		SCMPlatformID: platform.Basic.ID,
		Reference: scm.GetGitlabProjectReference(project),
		Stage: scm.ApprovalCreated,
		OwnerID: int(ownerID),
	}
	if err = db.Save(approval).Error; err != nil {
		return nil, err
	}

	// submit job to create mr.
	return approval, nil
}