package gitlab

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"git.stuhome.com/Sunmxt/wing/common"
)

type Author struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	UserName  string `json:"username"`
	State     string `json:"state"`
	AvatarURL string `json:"avatar_url"`
	WebURL    string `json:"web_url"`
}

type Milestone struct {
	ID          uint   `json:"id"`
	InternalID  uint   `json:"iid"`
	ProjectID   uint   `json:"project_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	CreateAt    string `json:"create_at"`
	UpdateAt    string `json:"update_at"`
	DueDate     string `json:"due_date"`
	StartDate   string `json:"start_date"`
	WebURL      string `json:"web_url"`
}

type MergeRequest struct {
	ID                        uint `json:"id"`
	InternalID                uint `json:"iid"`
	ProjectID                 uint `json:"ProjectID"`
	Project                   *Project
	Title                     string     `json:"title"`
	Description               string     `json:"description"`
	State                     string     `json:"state"`
	MergeBy                   *User      `json:"merge_by"`
	MergeAt                   string     `json:"merge_at"`
	CloseBy                   *User      `json:"close_by"`
	CloseAt                   string     `json:"close_at"`
	UpdateAt                  string     `json:"update_at"`
	TargetBranch              string     `json:"target_branch"`
	SourceBranch              string     `json:"source_branch"`
	Upvotes                   uint       `json:"upvotes"`
	Downvotes                 uint       `json:"downvotes"`
	Author                    *Author    `json:"author"`
	Assignee                  *Author    `json:"assignee"`
	Assignees                 []Author   `json:"assignees"`
	SourceProjectID           uint       `json:"source_project_id"`
	TargetProjectID           uint       `json:"target_project_id"`
	Labels                    []string   `json:"labels"`
	WorkInProgresss           uint       `json:"work_in_progress"`
	Milestone                 *Milestone `json:"milestone"`
	MergeWhenPipelineSucceeds bool       `json:"merge_when_pipeline_succeeds"`
	MergeStatus               string     `json:"merge_when_pipiline_succeeds"`
	SHA                       string     `json:"sha"`
	MergeCommitSHA            string     `json:"merge_commit_sha"`
	UserNotesCount            uint       `json:"user_notes_count"`
	DiscussionLocked          bool       `json:"discussion_locked"`
	ShouldRemoveSourceBranch  bool       `json:"should_remove_source_branch"`
	ForceRemoveSourceBranch   bool       `json:"force_remove_source_branch"`
	AllowCollaboration        bool       `json:"allow_collaboration"`
	AllowMaintainerToPush     bool       `json:"allow_maintainer_to_push"`
	WebURL                    string     `json:"web_url"`
	TimeStats                 *struct {
		TimeEstimate      uint `json:"time_estimate"`
		TotalTimeSpent    uint `json:"total_time_spent"`
		HumanTimeEstimate uint `json:"human_time_estimate"`
	} `json:"time_stats"`
	Squash               bool `json:"squash"`
	TaskCompletionStatus *struct {
		Count          uint `json:"count"`
		CompletedCount uint `json:"completed_count"`
	} `json:"task_completion_status"`
}

type MergeRequestContext struct {
	Cursor  GitlabPagination
	Project *Project
	Client  *GitlabClient
	Error   error
}

func NewMergeRequestContext(client *GitlabClient) *MergeRequestContext {
	return &MergeRequestContext{
		Client: client,
	}
}

func (c *MergeRequestContext) Clone() *MergeRequestContext {
	return &MergeRequestContext{
		Client:  c.Client,
		Project: c.Project,
	}
}

func (c *MergeRequestContext) WithProject(project *Project) *MergeRequestContext {
	newContext := c.Clone()
	newContext.Project = project
	return newContext
}

func (c *MergeRequestContext) Create(mr *MergeRequest) error {
	c.Error = nil
	if c.Client == nil || c.Client.Endpoint == nil {
		c.Error = common.ErrEndpointMissing
		return nil
	}
	projectID := mr.ProjectID
	if projectID < 1 && c.Project != nil {
		projectID = c.Project.ID
	}
	if projectID < 1 {
		return errors.New("Project ID not given.")
	}
	qURL := &url.URL{}
	*qURL = *c.Client.Endpoint
	projectIDString := strconv.FormatUint(uint64(projectID), 10)
	qURL.Path = "api/v4/" + projectIDString + "merge_requests"
	qURL.RawPath = ""
	qURL.ForceQuery = false
	qURL.RawQuery = ""
	qURL.Fragment = ""
	req, err := c.Client.NewRequest("POST", qURL.String(), nil)
	if err != nil {
		c.Error = err
		return nil
	}
	if mr.SourceBranch == "" {
		return errors.New("source branch not given.")
	}
	req.Form.Add("source_branch", mr.SourceBranch)
	if mr.TargetBranch == "" {
		return errors.New("target branch not given.")
	}
	req.Form.Add("target_branch", mr.TargetBranch)
	if mr.Title == "" {
		return errors.New("title not given.")
	}
	req.Form.Add("title", mr.Title)
	req.Form.Add("id", projectIDString)
	if mr.Assignee != nil && mr.Assignee.ID > 0 {
		req.Form.Add("assignee_id", strconv.FormatUint(uint64(mr.Assignee.ID), 10))
	}
	if mr.Description != "" {
		req.Form.Add("description", mr.Description)
	}
	if mr.Milestone != nil && mr.Milestone.ID > 0 {
		req.Form.Add("milestone_id", strconv.FormatUint(uint64(mr.Milestone.ID), 10))
	}
	if mr.ShouldRemoveSourceBranch {
		req.Form.Add("remove_source_branch", strconv.FormatBool(mr.ShouldRemoveSourceBranch))
	}
	if mr.AllowCollaboration {
		req.Form.Add("allow_collaboration", strconv.FormatBool(mr.AllowCollaboration))
	}
	if mr.AllowMaintainerToPush {
		req.Form.Add("allow_maintainer_to_push", strconv.FormatBool(mr.AllowMaintainerToPush))
	}
	if mr.Squash {
		req.Form.Add("squash", strconv.FormatBool(mr.Squash))
	}
	client := http.Client{}

	var resp *http.Response
	c.Client.Info("[Gitlab Client] create merge request for project " + projectIDString)
	if resp, err = client.Do(req); err != nil {
		c.Client.Err("[Gitlab Client] create merge request failure:" + err.Error())
		c.Error = err
		return nil
	}
	mr = c.parseDetail(resp, mr)
	return c.Error
}

func (c *MergeRequestContext) parseDetail(resp *http.Response, mr *MergeRequest) *MergeRequest {
	body := make([]byte, resp.ContentLength)
	if _, err := io.ReadFull(resp.Body, body); err != nil {
		c.Client.Err("[Gitlab Client] read merge request detail response body failure: " + err.Error())
		c.Error = err
		return nil
	}
	if mr == nil {
		mr = &MergeRequest{}
	}
	if err := json.Unmarshal(body, mr); err != nil {
		c.Client.Err("[Gitlab Client] unmarsh merge request detail failure:" + err.Error())
		c.Error = err
		return nil
	}
	return mr
}
