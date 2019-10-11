package cicd

import (
	"git.stuhome.com/Sunmxt/wing/common"
	ccommon "git.stuhome.com/Sunmxt/wing/controller/common"
	"git.stuhome.com/Sunmxt/wing/model/scm"
	"git.stuhome.com/Sunmxt/wing/model/scm/gitlab"
	"github.com/jinzhu/gorm"
	"strconv"
)

func SubmitGitlabRepositoryCIApproval(ctx *ccommon.OperationContext, platform *scm.SCMPlatform, ownerID uint, repositoryID string) (*scm.CIRepositoryApproval, error) {
	repoID, err := strconv.ParseUint(repositoryID, 10, 64)
	defer func() {
		if err != nil {
			ctx.Log.Error(err.Error())
		}
	}()
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
		return nil, query.Error
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
		Type:          scm.GitlabMergeRequestApproval,
		SCMPlatformID: platform.Basic.ID,
		Reference:     scm.GetGitlabProjectReference(project),
		OwnerID:       int(ownerID),
	}
	tx := db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	// Only one approval can be in progress.
	if err = tx.Where(approval).Where("stage in (?)", []int{scm.ApprovalCreated, scm.ApprovalWaitForAccepted}).First(&approval).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	if approval.Basic.ID > 0 {
		return approval, nil
	}
	// Should not submit approval for ci enabled project.
	repo := &scm.CIRepository{
		SCMPlatformID: platform.Basic.ID,
		Reference:     approval.Reference,
		Active:        scm.Active,
	}
	if err = tx.Select("id").Where(repo).First(repo).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	if repo.Basic.ID > 0 {
		return nil, common.ErrRepositoryCIAlreadyEnabled
	}
	approval.Stage = scm.ApprovalCreated
	approval.AccessToken = common.GenerateRandomToken()
	approval.SetGitlabExtra(&scm.GitlabApprovalExtra{
		RepositoryID: uint(repoID),
	})
	if err = tx.Save(approval).Error; err != nil {
		return nil, err
	}
	// Save Log
	if _, _, err = scm.LogApprovalStageChanged(tx, platform.Basic.ID, int(repoID), approval.Basic.ID, -1, approval.Stage); err != nil {
		return nil, err
	}
	tx.Commit()

	// submit job to create mr.
	if _, err = AsyncSubmitCIApprovalGitlabMergeRequest(ctx, platform.Basic.ID, uint(repoID), approval.Basic.ID, 10); err != nil {
		return approval, err
	}

	return approval, nil
}
