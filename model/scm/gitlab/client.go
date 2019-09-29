package gitlab

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/log"
	"github.com/Sunmxt/form"
)

type GitlabClientLogger log.NormalLogger

type GitlabPagination struct {
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

type GitlabErrorMessage struct {
	Message      string `json:"message"`
	ErrorMessage string `json:"error"`
}

func (i *GitlabPagination) Next() bool {
	if i.Page != 0 && i.Page >= i.TotalPage {
		return false
	}
	i.Page = i.NextPage
	i.NextPage++
	i.PrevPage++
	return true
}

func (i *GitlabPagination) GetPerPage() uint {
	if i.PerPage > 0 {
		return i.PerPage
	}
	return 10
}

func (i *GitlabPagination) SetPage(page uint) {
	if page < 1 {
		page = 1
	}
	i.PrevPage = page - 1
	i.NextPage = page + 1
	i.Page = page
}

func (i *GitlabPagination) GetPagnationURLQueries() string {
	return "page=" + strconv.FormatUint(uint64(i.Page), 10) + "&" + "per_page=" + strconv.FormatUint(uint64(i.GetPerPage()), 10)
}

func (i *GitlabPagination) updateCursorFromResponse(client *GitlabClient, resp *http.Response) {
	tryParse := func(headerName string, defaultValue uint) uint {
		pages, exists := resp.Header[headerName]
		if exists && len(pages) > 0 && len(pages[0]) > 0 {
			page, err := strconv.ParseUint(pages[0], 10, 64)
			if err != nil {
				client.Err("[Gitlab Client] Parse \"" + headerName + "\" with value \"" + pages[0] + "\" failure: " + err.Error())
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

func (i *GitlabPagination) Reset() {
	i.Page = 1
	i.NextPage = 2
	i.PrevPage = 0
}

type GitlabClient struct {
	Error       error
	Endpoint    *url.URL
	Logger      GitlabClientLogger
	AccessToken string
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

func (c *GitlabClient) Err(args ...interface{}) {
	if c.Logger == nil {
		return
	}
	c.Logger.Error(args...)
}

func (c *GitlabClient) Info(args ...interface{}) {
	if c.Logger == nil {
		return
	}
	c.Logger.Info(args...)
}

func (c *GitlabClient) ProjectQuery() *ProjectQuery {
	return NewProjectQuery(c)
}

func (c *GitlabClient) MergeRequest() *MergeRequestContext {
	return NewMergeRequestContext(c)
}

func (c *GitlabClient) User() *UserContext {
	return NewUserContext(c)
}

func (c *GitlabClient) Variable() *VariableContext {
	return NewVariableContext(c)
}

func (c *GitlabClient) NewRequest(method string, requestURL string, options interface{}) (req *http.Request, err error) {
	if c.Endpoint == nil {
		err = common.ErrEndpointMissing
		c.Error = err
		return req, err
	}
	qURL, err := url.Parse(requestURL)
	if qURL.Scheme == "" {
		qURL.Scheme = c.Endpoint.Scheme
	}
	if qURL.Opaque == "" {
		qURL.Opaque = c.Endpoint.Opaque
	}
	if qURL.User == nil {
		qURL.User = c.Endpoint.User
	}
	if qURL.Host == "" {
		qURL.Host = c.Endpoint.Host
	}
	qURL.ForceQuery = false
	if qURL.Fragment == "" {
		qURL.Fragment = c.Endpoint.Fragment
	}
	var body io.Reader
	if options != nil {
		values, err := form.EncodeToValues(options)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(values.Encode())
	}
	if req, err = http.NewRequest(method, qURL.String(), body); err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if c.AccessToken != "" {
		req.Header.Set("Private-Token", c.AccessToken)
	}
	return req, nil
}

func (c *GitlabClient) Do(req *http.Request, results ...interface{}) (resp *http.Response, err error) {
	c.Error = nil
	client := http.Client{}

	resp, err = client.Do(req)
	c.Info("[Gitlab Client] sent http request " + req.URL.String() + ". got status " + strconv.FormatUint(uint64(resp.StatusCode), 10))
	if err != nil {
		c.Err("[Gitlab Client] http request get failure with code: " + err.Error())
		c.Error = err
		return nil, err
	}
	if err = c.parseError(resp); err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	bodyReader := io.TeeReader(resp.Body, &buf)
	for _, result := range results {
		if err = json.NewDecoder(bodyReader).Decode(result); err != nil {
			c.Error = err
			break
		}
		bodyReader = bytes.NewReader(buf.Bytes())
	}
	resp.Body = ioutil.NopCloser(bodyReader)
	return resp, nil
}

func (c *GitlabClient) parseError(resp *http.Response) error {
	if resp.StatusCode < 300 && resp.StatusCode >= 200 {
		return nil
	}
	msg := &GitlabErrorMessage{}
	err := json.NewDecoder(resp.Body).Decode(msg)
	if err != nil {
		return fmt.Errorf("[Gitlab Client] Request got status %v, but cannot parse error message: "+err.Error(), resp.StatusCode)
	}
	message := "status: " + strconv.FormatUint(uint64(resp.StatusCode), 10)
	if msg.Message != "" {
		message = "message: " + msg.Message + ". "
	}
	if msg.ErrorMessage != "" {
		message += "error: " + msg.ErrorMessage + ". "
	}
	return errors.New(message)
}

func (c *GitlabClient) EndpointClone() (*url.URL, error) {
	return url.Parse(c.Endpoint.String())
}
