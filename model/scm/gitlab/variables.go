package gitlab

import (
	"errors"
	"net/http"
	"strconv"

	"git.stuhome.com/Sunmxt/wing/common"
)

type VariableDetail struct {
}

type Variable struct {
	Key              string `json:"key" form:"key"`
	Value            string `json:"value" form:"value"`
	Type             string `json:"variable_type" form:"variable_type,omitempty"`
	Protected        bool   `json:"protected" form:"protected"`
	Masked           bool   `json:"masked" form:"masked"`
	EnvironmentScope string `json:"environment_scope" form:"environment_scope"`
}

type VariableContext struct {
	Project *Project
	Client  *GitlabClient
	Vars    []Variable
	Error   error
}

func NewVariableContext(client *GitlabClient) *VariableContext {
	return &VariableContext{
		Client: client,
	}
}

func (c *VariableContext) Clone() *VariableContext {
	return &VariableContext{
		Project: c.Project,
		Client:  c.Client,
	}
}

func (c *VariableContext) WithProject(project *Project) *VariableContext {
	nc := c.Clone()
	nc.Project = project
	return nc
}

func (c *VariableContext) preCheck() (err error) {
	if c.Project == nil {
		err = errors.New("Project not given.")
		c.Error = err
		return err
	}
	if c.Client == nil {
		err = common.ErrEndpointMissing
		c.Error = err
		return err
	}
	return nil
}

func (c *VariableContext) List() []Variable {
	projectIDString := strconv.FormatUint(uint64(c.Project.ID), 10)
	req, err := c.Client.NewRequest("GET", "api/v4/projects/"+projectIDString+"/variables", nil)
	if err != nil {
		c.Error = err
		return nil
	}
	if c.Vars != nil {
		c.Vars = c.Vars[0:0]
	}
	if _, err := c.Client.Do(req, &c.Vars); err != nil {
		c.Error = err
		return nil
	}
	return c.Vars
}

func (c *VariableContext) Create(variable *Variable) error {
	if err := c.preCheck(); err != nil {
		return err
	}
	if variable == nil {
		c.Error = errors.New("variable not given.")
		return c.Error
	}
	projectIDString := strconv.FormatUint(uint64(c.Project.ID), 10)
	req, err := c.Client.NewRequest("POST", "api/v4/projects/"+projectIDString+"/variables", variable)
	if err != nil {
		c.Error = err
		return err
	}
	if _, err := c.Client.Do(req, variable); err != nil {
		c.Error = err
		return err
	}
	return nil
}

func (c *VariableContext) Save(variable *Variable) error {
	if err := c.preCheck(); err != nil {
		return err
	}
	if variable == nil {
		c.Error = errors.New("variable not given.")
		return c.Error
	}
	projectIDString := strconv.FormatUint(uint64(c.Project.ID), 10)
	method := "PUT"
	req, err := c.Client.NewRequest(method, "api/v4/projects/"+projectIDString+"/variables/"+variable.Key, variable)
	if err != nil {
		c.Error = err
		return err
	}
	var resp *http.Response
	if resp, err = c.Client.Do(req, variable); err != nil {
		c.Error = err
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return c.Create(variable)
	}
	return nil
}
