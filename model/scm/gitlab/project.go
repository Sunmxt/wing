package gitlab

import (
	"git.stuhome.com/Sunmxt/wing/common"
	"net/http"
	"strconv"
)

type Project struct {
	ID                uint             `json:"id"`
	Description       string           `json:"description"`
	Name              string           `json:"name"`
	NameWithNamespace string           `json:"name_with_namespace"`
	Path              string           `json:"path"`
	PathWithNamespace string           `json:"path_with_namespace"`
	CreateAt          string           `json:"create_at"`
	DefaultBranch     string           `json:"default_branch"`
	TagList           []string         `json:"tag_list"`
	SSHURLToRepo      string           `json:"ssh_url_to_repo"`
	HTTPURLToRepo     string           `json:"http_url_to_repo"`
	WebURL            string           `json:"web_url"`
	ReadmeURL         string           `json:"readme_url"`
	AvatarURL         string           `json:"avatar_url"`
	StarCount         uint             `json:"star_count"`
	ForksCount        uint             `json:"forks_count"`
	LastActivityAt    string           `json:"last_activity_at"`
	Namespace         *GitlabNamespace `json:"namespace"`

	Detail *ProjectDetail `json:"-"`
}

//func (q *Project) MergeRequest() *GitlabMergeRequestContext {
//	return NewMergeRequestContext(q)
//}

type GitlabOwner struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ProjectPermissionLevel struct {
	AccessLevel       uint `json:"access_level"`
	NotificationLevel uint `json:"notification_level"`
}

type ProjectPermission struct {
	ProjectAccess ProjectPermissionLevel `json:"project_access"`
	GroupAccess   ProjectPermissionLevel `json:"group_access"`
}

type GitlabLicense struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	NickName  string `json:"nickname"`
	HTMLURL   string `json:"html_url"`
	SourceURL string `json:"source_url"`
}

type ProjectSharedWithGroup struct {
	ID          uint   `json:"group_id"`
	Name        string `json:"group_name"`
	FullPath    string `json:"full_path"`
	AccessLevel uint   `json:"access_level"`
}

type ProjectStatistics struct {
	CommitCount      uint `json:"commit_count"`
	StorageSize      uint `json:"storage_size"`
	RepositorySize   uint `json:"repository_size"`
	WikiSize         uint `json:"wiki_size"`
	LFSObjectsSIze   uint `json:"lfs_objects_size"`
	JobArtifactsSize uint `json:"job_artifacts_size"`
	PackagesSize     uint `json:"packages_size"`
}

type ProjectLinks struct {
	Self          string `json:"self"`
	Issues        string `json:"issues"`
	MergeRequests string `json:"merge_requests"`
	RepoBranches  string `json:"repo_branches"`
	Labels        string `json:"lables"`
	Events        string `json:"events"`
	Members       string `json:"members"`
}

type ProjectDetail struct {
	Visibility                             string                   `json:"visibility"`
	Owner                                  *GitlabOwner             `json:"owner"`
	IssuesEnabled                          bool                     `json:"issues_enabled"`
	OpenIssuesCount                        bool                     `json:"open_issues_count"`
	JobsEnabled                            bool                     `json:"jobs_enabled"`
	WikiEnabled                            bool                     `json:"wiki_enabled"`
	SnippetsEnabled                        bool                     `json:"SnippetsEnabled"`
	ResolveOutdatedDiffDiscussions         bool                     `json:"resolve_outdated_diff_discussions"`
	ContainerRegistryEnabled               bool                     `json:"container_registry_enabled"`
	CreatorID                              uint                     `json:"creator_id"`
	ImportStatus                           string                   `json:"import_status"`
	ImportError                            string                   `json:"import_error"`
	Permissions                            *ProjectPermission       `json:"permissions"`
	Archived                               bool                     `json:"Archived"`
	LicenseURL                             string                   `json:"license_url"`
	License                                *GitlabLicense           `json:"license"`
	SharedRunnerEnabled                    bool                     `json:"shared_runners_enabled"`
	RunnersToken                           string                   `json:"runners_token"`
	CIDefaultGitDepth                      uint                     `json:"ci_default_git_depth"`
	PublicJobs                             bool                     `json:"public_jobs"`
	SharedWithGroups                       []ProjectSharedWithGroup `json:"shared_with_groups"`
	RepositoryStorage                      string                   `json:"repository_storage"`
	OnlyAllowMergeIfAllDiscussionsResolved bool                     `json:"only_allow_merge_if_all_discussions_are_resolved"`
	OnlyAllowMergeIfAllPipelineSucceeds    bool                     `json:"only_allow_merge_if_pipeline_succeeds"`
	RequestAccessEnabled                   bool                     `json:"request_access_enabled"`
	MergeMethod                            string                   `json:"merge_method"`
	Statistics                             *ProjectStatistics       `json:"statistics"`
	Links                                  *ProjectLinks            `json:"_links"`
}

type ProjectQuery struct {
	Cursor   GitlabPagination
	Projects []Project
	Client   *GitlabClient
	Error    error
}

func NewProjectQuery(client *GitlabClient) (q *ProjectQuery) {
	q = &ProjectQuery{
		Client: client,
	}
	q.Cursor.Reset()
	return q
}

func (q *ProjectQuery) Single(ID uint) *Project {
	q.Error = nil
	if q.Client == nil || q.Client.Endpoint == nil {
		q.Error = common.ErrEndpointMissing
		return nil
	}
	req, err := q.Client.NewRequest("GET", "api/v4/projects/"+strconv.FormatUint(uint64(ID), 10), nil)
	if err != nil {
		q.Error = err
		return nil
	}
	project := &Project{}
	if _, err = q.Client.Do(req, project, &project.Detail); err != nil {
		q.Error = err
		return nil
	}
	return project
}

func (q *ProjectQuery) ResetCursor() error {
	q.Cursor.Reset()
	return nil
}

func (g *ProjectQuery) PerPage(perPage uint) {
	g.Cursor.PerPage = perPage
	g.Cursor.Reset()
}

func (g *ProjectQuery) Page(page uint) {
	g.Cursor.SetPage(page)
}

func (q *ProjectQuery) Refresh() *ProjectQuery {
	q.Error = nil
	if q.Client == nil || q.Client.Endpoint == nil {
		q.Error = common.ErrEndpointMissing
		return q
	}
	req, err := q.Client.NewRequest("GET", "api/v4/projects?"+q.Cursor.GetPagnationURLQueries(), nil)
	if err != nil {
		q.Error = err
		return q
	}
	if q.Projects != nil {
		q.Projects = q.Projects[0:0]
	}
	var resp *http.Response
	if resp, err = q.Client.Do(req, &q.Projects); err != nil {
		q.Client.Err("[Gitlab Client] list gitlab projects failure: " + err.Error())
		q.Error = err
		return q
	}
	q.Cursor.updateCursorFromResponse(q.Client, resp)
	return q
}

func (q *ProjectQuery) Next() bool {
	if !q.Cursor.Next() {
		return false
	}
	q.Refresh()
	return true
}
