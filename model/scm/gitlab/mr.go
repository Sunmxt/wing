package gitlab

import (
	"errors"
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
	ID              uint     `json:"id" form:"-"`
	InternalID      uint     `json:"iid" form:"-"`
	ProjectID       uint     `json:"ProjectID" form:"id"`
	Project         *Project `json:"-" form:"-"`
	Title           string   `json:"title" form:"title"`
	Description     string   `json:"description" form:"description,omitempty"`
	State           string   `json:"state" form:"-"`
	MergeBy         *User    `json:"merge_by" form:"-"`
	MergeAt         string   `json:"merge_at" form:"-"`
	CloseBy         *User    `json:"close_by" form:"-"`
	CloseAt         string   `json:"close_at" form:"-"`
	UpdateAt        string   `json:"update_at" form:"-"`
	TargetBranch    string   `json:"target_branch" form:"target_branch"`
	SourceBranch    string   `json:"source_branch" form:"source_branch"`
	Upvotes         uint     `json:"upvotes" form:"-"`
	Downvotes       uint     `json:"downvotes" form:"-"`
	Author          *Author  `json:"author" form:"-"`
	Assignee        *Author  `json:"assignee" form:"-"`
	AssigneeID      uint     `json:"-" form:"assignee_id,omitempty"`
	Assignees       []Author `json:"assignees" form:"-"`
	SourceProjectID uint     `json:"source_project_id" form:"-"`
	TargetProjectID uint     `json:"target_project_id" form:"target_project_id,omitempty"`
	//Labels          []string `json:"labels" form:"labels,omitempty"`
	//WorkInProgresss           uint       `json:"work_in_progress" form:"-"`
	Milestone                 *Milestone `json:"milestone" form:"-"`
	MilestoneID               uint       `json:"-" form:"milestone_id,omitempty"`
	MergeWhenPipelineSucceeds bool       `json:"merge_when_pipeline_succeeds" form:"-"`
	MergeStatus               string     `json:"merge_when_pipiline_succeeds" form:"-"`
	SHA                       string     `json:"sha" form:"-"`
	MergeCommitSHA            string     `json:"merge_commit_sha" form:"-"`
	UserNotesCount            uint       `json:"user_notes_count" form:"-"`
	DiscussionLocked          bool       `json:"discussion_locked" form:"-"`
	ShouldRemoveSourceBranch  bool       `json:"should_remove_source_branch" form:"-"`
	ForceRemoveSourceBranch   bool       `json:"force_remove_source_branch" form:"remove_source_branch,omitempty"`
	AllowCollaboration        bool       `json:"allow_collaboration" form:"allow_collaboration,omitempty"`
	AllowMaintainerToPush     bool       `json:"allow_maintainer_to_push" form:"allow_maintainer_to_push,omitempty"`
	WebURL                    string     `json:"web_url" form:"-"`
	TimeStats                 struct {
		TimeEstimate      uint `json:"time_estimate" form:"-"`
		TotalTimeSpent    uint `json:"total_time_spent" form:"-"`
		HumanTimeEstimate uint `json:"human_time_estimate" form:"-"`
	} `json:"time_stats" form:"-"`
	Squash               bool `json:"squash" form:"squash,omitempty"`
	TaskCompletionStatus struct {
		Count          uint `json:"count" form:"-"`
		CompletedCount uint `json:"completed_count" form:"-"`
	} `json:"task_completion_status" form:"-"`
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
		return c.Error
	}
	projectID := mr.ProjectID
	if projectID < 1 && c.Project != nil {
		projectID = c.Project.ID
	}
	if projectID < 1 {
		return errors.New("Project ID not given.")
	}
	if mr.SourceBranch == "" {
		return errors.New("source branch not given.")
	}
	if mr.TargetBranch == "" {
		return errors.New("target branch not given.")
	}
	if mr.Title == "" {
		return errors.New("title not given.")
	}
	mr.ProjectID = projectID
	projectIDString := strconv.FormatUint(uint64(projectID), 10)
	req, err := c.Client.NewRequest("POST", "api/v4/projects/"+projectIDString+"/merge_requests", mr)
	if err != nil {
		c.Error = err
		return err
	}
	if mr.Assignee != nil {
		mr.AssigneeID = mr.Assignee.ID
	} else {
		mr.AssigneeID = 0
	}
	if mr.Milestone != nil {
		mr.MilestoneID = mr.Milestone.ID
	} else {
		mr.MilestoneID = 0
	}
	if _, err = c.Client.Do(req, mr); err != nil {
		c.Error = err
		return err
	}
	return nil
}
