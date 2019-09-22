package cicd

import (
	ccommon "git.stuhome.com/Sunmxt/wing/controller/common"
	"git.stuhome.com/Sunmxt/wing/common"
	"strconv"
	"git.stuhome.com/Sunmxt/wing/model/scm/gitlab"
	"git.stuhome.com/Sunmxt/wing/model/scm"
	"github.com/jinzhu/gorm"
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
		OwnerID: int(ownerID),
	}
	tx := db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	// Only one approval can be in progress.
	if err = db.Where(approval).Where("stage >= (?)", scm.ApprovalCreated).First(&approval).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	if approval.Basic.ID > 0 {
		tx.Rollback()
		return approval, nil
	}
	approval.Stage = scm.ApprovalCreated
	if err = tx.Save(approval).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	// submit job to create mr.
	AsyncSubmitCIApprovalMergeRequest(ctx, platform.Basic.ID, uint(repoID), approval.Basic.ID)

	return approval, nil
}