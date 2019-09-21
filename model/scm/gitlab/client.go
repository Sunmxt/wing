package gitlab

import (
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/log"
	"net/http"
	"net/url"
	"strconv"
)

type GitlabPagnation struct {
	Total     uint
	TotalPage uint
	PerPage   uint
	Page      uint
	NextPage  uint
	PrevPage  uint
}

type GitlabNamespace struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	FullPath string `json:"full_path"`
	ParentID int    `json:"parent_id"`
}

func (i *GitlabPagnation) Next() bool {
	if i.Page != 0 && i.Page >= i.TotalPage {
		return false
	}
	i.Page = i.NextPage
	i.NextPage++
	i.PrevPage++
	return true
}

func (i *GitlabPagnation) GetPerPage() uint {
	if i.PerPage > 0 {
		return i.PerPage
	}
	return 10
}

func (i *GitlabPagnation) SetPage(page uint) {
	if page < 1 {
		page = 1
	}
	i.PrevPage = page - 1
	i.NextPage = page + 1
	i.Page = page
}

func (i *GitlabPagnation) GetPagnationURLQueries() string {
	return "page=" + strconv.FormatUint(uint64(i.Page), 10) + "&" + "per_page=" + strconv.FormatUint(uint64(i.GetPerPage()), 10)
}

func (i *GitlabPagnation) updateCursorFromResponse(resp *http.Response, q *ProjectQuery) {
	tryParse := func(headerName string, defaultValue uint) uint {
		pages, exists := resp.Header[headerName]
		if exists && len(pages) > 0 && len(pages[0]) > 0 {
			page, err := strconv.ParseUint(pages[0], 10, 64)
			if err != nil {
				q.error("[Gitlab Client] Parse \"" + headerName + "\" with value \"" + pages[0] + "\" failure: " + err.Error())
			} else {
				defaultValue = uint(page)
			}
		}
		return defaultValue
	}

	i.Total = tryParse("X-Total", 0)
	i.Page = tryParse("X-Page", i.Page+1)
	i.NextPage = tryParse("X-Next-Page", i.Page+1)
	i.PrevPage = tryParse("X-Prev-Page", i.Page-1)
	i.TotalPage = tryParse("X-Total-Pages", i.Total/i.PerPage+1)
}

func (i *GitlabPagnation) Reset() {
	i.Page = 1
	i.NextPage = 2
	i.PrevPage = 0
}

type GitlabClient struct {
	Error    error
	Endpoint *url.URL
	Logger   log.NormalLogger
}

func NewGitlabClient(endpoint string, logger log.NormalLogger) (*GitlabClient, error) {
	if endpoint == "" {
		return nil, common.ErrEndpointMissing
	}
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	q := &GitlabClient{
		Endpoint: endpointURL,
		Logger:   logger,
	}

	q.Endpoint.Path = ""
	q.Endpoint.RawPath = ""
	q.Endpoint.ForceQuery = false
	q.Endpoint.RawQuery = ""
	q.Endpoint.Fragment = ""

	return q, nil
}

func (c *GitlabClient) ProjectQuery() *ProjectQuery {
	return NewProjectQuery(c)
}

func (c *GitlabClient) EndpointClone() (*url.URL, error) {
	return url.Parse(c.Endpoint.String())
}
