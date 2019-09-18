package scm

import (
	"encoding/json"
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/log"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type GitlabNamespace struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	FullPath string `json:"full_path"`
	ParentID int    `json:"parent_id"`
}

type GitlabPageInfo struct {
	TotalPage uint
	PerPage   uint
	Page      uint
	NextPage  uint
	PrevPage  uint
}

func (i *GitlabPageInfo) Next() bool {
	if i.Page != 0 && i.Page >= i.TotalPage {
		return false
	}
	i.Page = i.NextPage
	i.NextPage++
	i.PrevPage++
	return true
}

func (i *GitlabPageInfo) GetPerPage() uint {
	if i.PerPage > 0 {
		return i.PerPage
	}
	return 10
}

func (i *GitlabPageInfo) GetPagnationURLQueries() string {
	return "page=" + strconv.FormatUint(uint64(i.Page), 10) + "&" + "per_page=" + strconv.FormatUint(uint64(i.GetPerPage()), 10)
}

func (i *GitlabPageInfo) updateCursorFromResponse(resp *http.Response) {
}

func (i *GitlabPageInfo) Reset() {
	i.Page = 1
	i.NextPage = 2
	i.PrevPage = 0
}

type GitlabProject struct {
	ID                int              `json:"id"`
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
	StarCount         string           `json:"star_count"`
	ForksCount        string           `json:"forks_count"`
	LastActivityAt    string           `json:"last_activity_at"`
	Namespace         *GitlabNamespace `json:"namespace"`
}

type GitlabProjectQuery struct {
	Cursor      GitlabPageInfo
	Total       uint
	AccessToken string
	Error       error
	Endpoint    *url.URL
	Projects    []GitlabProject
	Logger      log.NormalLogger
}

func NewGitlabProjectQuery(endpoint string) (*GitlabProjectQuery, error) {
	if endpoint == "" {
		return nil, common.ErrEndpointMissing
	}
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	q := &GitlabProjectQuery{
		Endpoint: endpointURL,
	}
	q.Cursor.Reset()

	return q, nil
}

func (q *GitlabProjectQuery) ResetCursor() error {
	q.Cursor.Reset()
	return nil
}

func (g *GitlabProjectQuery) PerPage(perPage uint) {
	g.Cursor.PerPage = perPage
	g.Cursor.Reset()
}

func (q *GitlabProjectQuery) info(args ...interface{}) {
	if q.Logger == nil {
		return
	}
	q.Logger.Info(args...)
}

func (q *GitlabProjectQuery) error(args ...interface{}) {
	if q.Logger == nil {
		return
	}
	q.Logger.Error(args...)
}

func (q *GitlabProjectQuery) Refresh() *GitlabProjectQuery {
	if q.Endpoint == nil {
		q.Error = common.ErrEndpointMissing
		return nil
	}
	qURL := &url.URL{}
	*qURL = *q.Endpoint

	qURL.Path = "api/v4/projects"
	qURL.RawPath = ""
	qURL.ForceQuery = false
	qURL.RawQuery = q.Cursor.GetPagnationURLQueries()
	qURL.Fragment = ""

	req, err := http.NewRequest("GET", qURL.String(), nil)
	if err != nil {
		q.Error = err
		return nil
	}
	client := http.Client{}

	var resp *http.Response
	q.info("[Gitlab Client] list gitlab projects. " + qURL.String())
	if resp, err = client.Do(req); err != nil {
		q.error("[Gitlab Client] list gitlab projects failure: " + err.Error())
		q.Error = err
		return nil
	}

	return q.parseResult(resp)
}

func (q *GitlabProjectQuery) parseResult(resp *http.Response) *GitlabProjectQuery {
	q.Cursor.updateCursorFromResponse(resp)
	body := make([]byte, resp.ContentLength, resp.ContentLength)
	if _, err := io.ReadFull(resp.Body, body); err != nil {
		q.error("[Gitlab Client] read projects response body failure: " + err.Error())
		q.Error = err
		return nil
	}
	if q.Projects == nil {
		q.Projects = make([]GitlabProject, q.Cursor.GetPerPage())
	}
	if err := json.Unmarshal(body, &q.Projects); err != nil {
		q.error("[Gitlab Client] unmarshal projects response failure: " + err.Error())
		q.Error = err
		return nil
	}
	return q
}

func (q *GitlabProjectQuery) Next() bool {
	if !q.Cursor.Next() {
		return false
	}
	q.Refresh()
	return true
}
