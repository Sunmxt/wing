package cicd

import (
    "git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/cmd/runtime"
	"git.stuhome.com/Sunmxt/wing/model/scm/gitlab"
	"git.stuhome.com/Sunmxt/wing/model/scm"
	ccommon "git.stuhome.com/Sunmxt/wing/controller/common"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/RichardKnop/machinery/v1/backends/result"
	log "github.com/sirupsen/logrus"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"

	"path/filepath"
	"net/url"
	"time"
	"errors"
	"os"
	"fmt"
	"strconv"
)


func RegisterTasks(runtime *runtime.WingRuntime) (err error) {
	tasks := map[string]interface{}{
		"SubmitCIApprovalGitlabMergeRequest": SubmitCIApprovalGitlabMergeRequest,
		"GitlabMergeRequestFinishCIApproval": GitlabMergeRequestFinishCIApproval,
	}

	for name, task := range tasks {
		if err := runtime.RegisterTask(name, task); err != nil {
			log.Error("Register task \"" + name + "\" failure: " + err.Error())
			return err
		}
		log.Info("Register task \"" + name + "\".")
	}
	return nil
}


func PrepareGitlabLocalRepository(octx *ccommon.OperationContext, project *gitlab.Project, accessToken string) (*git.Repository, string, error) {
	if project.Name == "" || project.Namespace == nil || project.Namespace.Name == "" {
		err := fmt.Errorf("missing gitlab project identifier. got project name %v. namespace = %v.", project.Name, project.Namespace)
		octx.Log.Error(err.Error())
		return nil, "", err
	}
	token := common.GenerateRandomToken()
	dirname := "gitlabrepo_approval_" + token + "_" + project.Name + "_" + project.Namespace.Name
	cloneTo, err := filepath.Abs(filepath.Join(octx.Runtime.Config.Session.Job.GitWorkingDir, dirname))
	if err != nil {
		return nil, "", err
	}
	if project.HTTPURLToRepo == "" {
		err = errors.New("missing http url for project.")
		octx.Log.Error(err.Error())
		return nil, "", err
	}
	cloneURL, err := url.Parse(project.HTTPURLToRepo)
	if err != nil {
		return nil, "", err
	}
	if accessToken != "" {
		cloneURL.User = url.UserPassword("oauth2", accessToken)
	}

	var dir *os.File
	if dir, err = os.Open(cloneTo); err != nil && !os.IsNotExist(err) {
		return nil, "", err
	}
	if dir != nil {
		octx.Log.Info(cloneTo + " not empty. try to remove content.")
		dir.Close()
		if err = os.RemoveAll(cloneTo); err != nil {
			octx.Log.Error("cannot remove " + cloneTo + ": " + err.Error())
			return nil, "", err
		}
	}

	// clone
	var repo *git.Repository
	octx.Log.Info("git clone " + project.HTTPURLToRepo + " into " + cloneTo)
	if repo, err = git.PlainClone(cloneTo, false, &git.CloneOptions{
		URL: cloneURL.String(),
	}); err != nil {
		return nil, "", err
	}
	octx.Log.Info("git clone " + project.HTTPURLToRepo + " finish.")

	return repo, cloneTo, err
}


func SyncGitlabCISettingWithBuilds(ctx *ccommon.OperationContext, gitlabCIConfigPath string, approval *scm.CIRepositoryApproval, ciRepo *scm.CIRepository) (bool, error) {
	return false, nil
}

func AsyncGitlabMergeRequestFinishCIApproval(ctx *ccommon.OperationContext, projectID, mergeRequestID uint) (*result.AsyncResult, error) {
	return ctx.SubmitTask("GitlabMergeRequestFinishCIApproval", []tasks.Arg{
		{Type: "uint", Value: projectID},
		{Type: "uint", Value: mergeRequestID},
	}, 5)
}

func GitlabMergeRequestFinishCIApproval(ctx *runtime.WingRuntime) interface{} {
	return func (projectID, mergeRequestID uint) error {
		octx := ccommon.NewOperationContext(ctx)
		octx.Log.Data["task"] = "GitlabMergeRequestFinishCIApproval"
		db, err := octx.Database()
		defer func() {
			if err != nil {
				octx.Log.Error(err.Error())
			}
		}()
		if err != nil {
			return err
		}
		octx.Log.Infof("Start finish approval related to gitlab merge request %v. related project %v", mergeRequestID, projectID)
		reference := strconv.FormatUint(uint64(projectID), 10)
		approval := &scm.CIRepositoryApproval{}
		tx := db.Begin()
		defer func() {
			if err != nil {
				tx.Rollback()
			}
		}()
		if err = approval.ByReference(tx.Preload("SCM").Where("stage in (?)", []int{scm.ApprovalCreated, scm.ApprovalWaitForAccepted}), reference); err != nil {
			return err
		}
		if approval.Basic.ID < 1 {
			octx.Log.Errorf("related approval not found for gitlab merge request %v", mergeRequestID)
			return nil
		}
		oldApprovalStage := approval.Stage
		approval.Stage = scm.ApprovalAccepted
		if err = tx.Save(approval).Error; err != nil {
			octx.Log.Errorf("save error: " + err.Error())
			return err
		}
		if _, _, err = scm.LogApprovalStageChanged(tx, approval.SCM.Basic.ID, int(projectID), approval.Basic.ID, oldApprovalStage, approval.Stage); err != nil {
			err = errors.New("save repo ci log error" + err.Error())
			return err
		}
		ciRepo := &scm.CIRepository{
			SCMPlatformID: approval.SCM.Basic.ID,
			Reference: approval.Reference,
			Active: scm.Active,
			AccessToken: approval.AccessToken,
			OwnerID: approval.OwnerID,
		}
		if err = tx.Save(ciRepo).Error; err != nil {
			octx.Log.Errorf("save error: " + err.Error())
			return err
		}
		err = tx.Commit().Error
		return err
	}
}

func AsyncSubmitCIApprovalGitlabMergeRequest(ctx *ccommon.OperationContext, platformID int, repositoryID uint, approvalID int, retry uint) (*result.AsyncResult, error) {
	return ctx.SubmitTask("SubmitCIApprovalGitlabMergeRequest", []tasks.Arg{
		{ Type: "int", Value: platformID },
		{ Type: "uint", Value: repositoryID },
		{ Type: "int", Value: approvalID },
	}, retry)
}

func SubmitCIApprovalGitlabMergeRequest(ctx *runtime.WingRuntime) interface{} {
	return func (platformID int, repositoryID uint, approvalID int) error {
		octx, platform := ccommon.NewOperationContext(ctx), &scm.SCMPlatform{}
		db, err := octx.Database()
		defer func () {
			if err != nil {
				octx.Log.Error("[SubmitCIApprovalGitlabMergeRequest] " + err.Error())
			}
		}()
		if err != nil {
			return err
		}
		if err = platform.ByID(db, platformID); err != nil {
			return err
		}
		if platform.Basic.ID < 1 {
			octx.Log.Error("[SubmitCIApprovalGitlabMergeRequest] scm platform not found.")
			return nil
		}
	    if platform.Type != scm.GitlabSCM {
			octx.Log.Error("[SubmitCIApprovalGitlabMergeRequest] not a gitlab scm. id = " + strconv.FormatInt(int64(platformID), 10))
			return nil
		}
		var client *gitlab.GitlabClient
		if client, err = platform.GitlabClient(octx.Log); err != nil {
			return err
		}
		project := client.ProjectQuery().Single(repositoryID)
		// Clone
		var repo *git.Repository
		var repoPath string
		if repo, repoPath, err = PrepareGitlabLocalRepository(octx, project, client.AccessToken); err != nil {
			return err
		}
		defer func () {
			// Cleaning.
			if err := os.RemoveAll(repoPath); err != nil {
				octx.Log.Warn("Cannot remove " + repoPath + ": " + err.Error())
			}
		}()
		// start submit mr.
		tx, approval := db.Begin(), &scm.CIRepositoryApproval{}
		defer func () {
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()
		if err = approval.ByID(tx, approvalID); err != nil {
			return err
		}
		if approval.Basic.ID < 1 {
			octx.Log.Error("[SubmitCIApprovalGitlabMergeRequest] approval not found. approval ID =" + strconv.FormatInt(int64(approvalID), 10))
			return nil
		}
		if approval.Stage != scm.ApprovalCreated {
			octx.Log.Info("[SubmitCIApprovalGitlabMergeRequest] merge request already created. approval ID = " + strconv.FormatInt(int64(approvalID), 10))
			return nil
		}
		if project == nil {
			approval.Stage = scm.ApprovalRejected
			tx.Save(approval)
			octx.Log.Error("[SubmitCIApprovalGitlabMergeRequest]" + common.ErrInvalidRepositoryID.Error())
			return nil
		}
		// branching
		branchName := "wing_scm_approval_" + strconv.FormatInt(int64(approval.Basic.ID), 10)
		branchRef := plumbing.ReferenceName("refs/heads/" + branchName)
		var ref *plumbing.Reference
		if ref, err = repo.Head(); err != nil {
			err = errors.New("cannot get head ref for repo: " + err.Error())
			return err
		}
		ref = plumbing.NewHashReference(branchRef, ref.Hash())
		if err = repo.Storer.SetReference(ref); err != nil {
			err = errors.New("branching \"" + string(branchRef) + "\" failure: " + err.Error())
			return err
		}
		// checkout branch
		var tree *git.Worktree
		if tree, err = repo.Worktree(); err != nil {
			return err
		}
		if err = tree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/heads/" + branchName),
		}); err != nil {
			return err
		}
		// update gitlab-ci.yml
		gitlabCIYAML, synced := filepath.Join(repoPath, ".gitlab-ci.yml"), false
		if synced, err = SyncGitlabCISettingWithBuilds(octx, gitlabCIYAML, approval, nil); err != nil {
			return err
		}
		if synced {
			if _, err := tree.Add(gitlabCIYAML); err != nil {
				return err
			}
		}
		// Commit
		userQuery := client.User()
		user := userQuery.Current()		
		if userQuery.Error != nil {
			err = userQuery.Error
			return err
		}
		if user == nil {
			err = errors.New("[SubmitCIApprovalGitlabMergeRequest] unable to get current gitlab user.")
			return err
		}
		var hash plumbing.Hash
		if hash, err = tree.Commit("[bot] wing: approval for enabling wing SCM of this project.", &git.CommitOptions{
			Author: &object.Signature{
				When: time.Now(),
				Name: user.Name,
				Email: user.Email,
			},
			Parents: []plumbing.Hash{
				ref.Hash(),
			},
		}); err != nil {
			return err
		}
		octx.Log.Info("[SubmitCIApprovalGitlabMergeRequest] commit hash: " + hash.String())
		// Remove remote branch
		octx.Log.Info("[SubmitCIApprovalGitlabMergeRequest] push branch " + branchName)
		if err = repo.Push(&git.PushOptions{
			RemoteName: "origin",
			RefSpecs: []config.RefSpec{
				config.RefSpec(":refs/heads/"+branchName),
			},
			Auth: &githttp.BasicAuth{
				Username: "oauth2", 
				Password: client.AccessToken,
			},
		}); err != nil {
			if err != git.NoErrAlreadyUpToDate{
				return err
			}
			err = nil
		}
		// Push
		if err = repo.Push(&git.PushOptions{
			RemoteName: "origin",
			RefSpecs: []config.RefSpec{
				config.RefSpec("refs/heads/"+branchName+":refs/heads/"+branchName),
			},
			Auth: &githttp.BasicAuth{
				Username: "oauth2", 
				Password: client.AccessToken,
			},
		}); err != nil {
			return err
		}
		// submit gitlab mr for approval.
		octx.Log.Info("[SubmitCIApprovalGitlabMergeRequest] submit merge request.")
		mr := &gitlab.MergeRequest{
			SourceBranch: branchName,
			TargetBranch: "master",
			Title: "[Wing] Enable SCM Build.",
		}
		if err = client.MergeRequest().WithProject(project).Create(mr); err != nil || mr.ID < 1 {
			err = fmt.Errorf("[SubmitCIApprovalGitlabMergeRequest] merge request not created. reason: %v", err)
			return err
		}
		// Configure access token for gitlab repo.
		octx.Log.Info("[SubmitCIApprovalGitlabMergeRequest] sync ci token for gitlab project " + project.NameWithNamespace)
		if err = client.Variable().WithProject(project).Save(&gitlab.Variable{
			Key: "WING_CI_TOKEN",
			Value: approval.AccessToken,
			Protected: true,
			Masked: true,
		}); err != nil {
			err = errors.New("[SubmitCIApprovalGitlabMergeRequest] update repo variable failure: " + err.Error())
			return err
		}
		var approvalExtra *scm.GitlabApprovalExtra
		approvalExtra = approval.GitlabExtra()
		if err != nil {
			octx.Log.Warn("[SubmitCIApprovalGitlabMergeRequest] invalid gitlab approval extra: " + err.Error() + ". extra will be refresh.")
		}
		if approvalExtra == nil {
			approvalExtra = &scm.GitlabApprovalExtra{}
		}
		approvalExtra.MergeRequestID = mr.ID
		approvalExtra.WebURL = mr.WebURL
		approval.SetGitlabExtra(approvalExtra)
		oldApprovalStage := approval.Stage
		approval.Stage = scm.ApprovalWaitForAccepted
		tx.Save(approval)
		if _, _, err = scm.LogApprovalStageChanged(tx, platformID, int(repositoryID), approvalID, oldApprovalStage, approval.Stage); err != nil {
			err = errors.New("[SubmitCIApprovalGitlabMergeRequest] cannot save repo ci log: " + err.Error())
			return err
		}
		return nil
	}
}


// SyncCIApproval submit related tasks to recover from inconsistent states caused by system failures. 
func SyncCIApproval(ctx *runtime.WingRuntime) (func() error) {
	return func () error {
		return nil
	}
}