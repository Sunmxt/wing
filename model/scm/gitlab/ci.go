package gitlab

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
)

var UnavaliableJobNames []string = []string{
	"image", "services", "stages", "types", "before_script", "after_script", "variables", "cache",
}

type CICache struct {
	Paths []string `json:"paths"`
}

type CIService struct {
	Name       string `json:"name,omitempty"`
	Entrypoint string `json:"entrypoint,omitempty"`
	Command    string `json:"command,omitempty"`
	Alias      string `json:"alias,omitempty"`
}

type CIJob struct {
	Cache        *CICache               `json:"cache,omitempty" yaml:"cache,omitempty"`
	Image        string                 `json:"image,omitempty" yaml:"image,omitempty"`
	BeforeScript []string               `json:"before_script,omitempty" yaml:"before_script,omitempty"`
	AfterScript  []string               `json:"after_script,omitempty" yaml:"after_script,omitempty"`
	Stage        string                 `json:"stage,omitempty" yaml:"stage,omitempty"`
	Script       []string               `json:"script,omitempty" yaml:"script,omitempty"`
	Variables    map[string]interface{} `json:"varaibles,omitempty" yaml:"variables,omitempty"`
	AllowFailure bool                   `json:"allow_failure,omitempty" yaml:"allow_failure,omitempty"`
	Tags         []string               `json:"tags,omitempty" yaml:"tags,omitempty"`
	Only         []string               `json:"only,omitempty" yaml:"only,omitempty"`
}

type CIConfiguration struct {
	Image    string `yaml:"image,omitempty"`
	Services struct {
		Simple   []string
		Detailed []CIService
	} `yaml:"-"`
	Stages       []string               `yaml:"stages,omitempty"`
	BeforeScript []string               `yaml:"before_script,omitempty"`
	AfterScript  []string               `yaml:"after_script,omitempty"`
	Variables    map[string]interface{} `yaml:"variables,omitempty"`
	Cache        *CICache               `yaml:"cache,omitempty"`

	Jobs    map[string]*CIJob `yaml: "-"`
	Include struct {
		Remote   []string
		Local    []string
		Template []string
		File     []string
		OneLine  []string
	} `yaml:"-"`

	RawInclude []interface{} `yaml:"include,omitempty"`
	RawService []interface{} `yaml:"services,omitempty"`
}

func (c *CIConfiguration) encodeServices() {
	c.RawService = make([]interface{}, 0, len(c.Services.Simple)+len(c.Services.Detailed))
	for _, svc := range c.Services.Simple {
		c.RawService = append(c.RawService, svc)
	}
	for _, svc := range c.Services.Detailed {
		c.RawService = append(c.RawService, svc)
	}
}

func (c *CIConfiguration) Encode(w io.Writer) (err error) {
	compact := make(map[string]interface{})
	if c.Image != "" {
		compact["image"] = c.Image
	}
	if c.Stages != nil {
		compact["stages"] = c.Stages
	}
	if c.BeforeScript != nil {
		compact["before_script"] = c.BeforeScript
	}
	if c.AfterScript != nil {
		compact["after_script"] = c.AfterScript
	}
	if c.Variables != nil {
		compact["variables"] = c.Variables
	}
	if c.Cache != nil {
		compact["cache"] = c.Cache
	}
	// Services
	c.encodeServices()
	if len(c.RawService) > 0 {
		compact["services"] = c.RawService
	}
	// Includes
	c.RawInclude = nil
	appendInc := func(typeName string, refs []string) {
		if refs == nil {
			return
		}
		for _, ref := range refs {
			c.appendRawInclude(map[string]string{
				typeName: ref,
			})
		}
	}
	appendInc("remote", c.Include.Remote)
	appendInc("local", c.Include.Local)
	appendInc("template", c.Include.Template)
	appendInc("file", c.Include.File)
	for _, ref := range c.Include.OneLine {
		c.appendRawInclude(ref)
	}
	if c.RawInclude != nil {
		compact["include"] = c.RawInclude
	}
	// Jobs
	for name, job := range c.Jobs {
		compact[name] = job
	}
	return yaml.NewEncoder(w).Encode(compact)
}

func (c *CIConfiguration) AppendOneLineInclude(line string) {
	if c.Include.OneLine == nil {
		c.Include.OneLine = make([]string, 0, 1)
	}
	c.Include.OneLine = append(c.Include.OneLine, line)
}

func (c *CIConfiguration) AppendRemoteInclude(line string) {
	if c.Include.Remote == nil {
		c.Include.Remote = make([]string, 0, 1)
	}
	c.Include.Remote = append(c.Include.Remote, line)
}

func (c *CIConfiguration) AppendTemplateInclude(line string) {
	if c.Include.Template == nil {
		c.Include.Template = make([]string, 0, 1)
	}
	c.Include.Template = append(c.Include.Template, line)
}

func (c *CIConfiguration) AppendLocalInclude(line string) {
	if c.Include.Local == nil {
		c.Include.Local = make([]string, 0, 1)
	}
	c.Include.Local = append(c.Include.Local, line)
}

func (c *CIConfiguration) AppendFileInclude(line string) {
	if c.Include.File == nil {
		c.Include.File = make([]string, 0, 1)
	}
	c.Include.File = append(c.Include.File, line)
}

func (c *CIConfiguration) appendRawInclude(inc interface{}) {
	if c.RawInclude == nil {
		c.RawInclude = make([]interface{}, 0)
	}
	c.RawInclude = append(c.RawInclude, inc)
}

func (c *CIConfiguration) decodeInclude(strict bool) (err error) {
	for _, rawInc := range c.RawInclude {
		switch v := rawInc.(type) {
		case string:
			c.AppendOneLineInclude(v)
		case map[interface{}]interface{}:
			for rawIncType, rawIncValue := range v {
				incType, ok := rawIncType.(string)
				if !ok && strict {
					return fmt.Errorf("unrecognized include key type. got %v.", rawIncType)
				} else {
					continue
				}
				var incValue string
				incValue, ok = rawIncValue.(string)
				if !ok && strict {
					return fmt.Errorf("unrecognized include value type. got %v", rawIncValue)
				} else {
					continue
				}
				switch incType {
				case "remote":
					c.AppendRemoteInclude(incValue)
				case "local":
					c.AppendLocalInclude(incValue)
				case "template":
					c.AppendTemplateInclude(incValue)
				case "file":
					c.AppendFileInclude(incValue)
				default:
					if strict {
						return fmt.Errorf("unrecognize include type: %v", incType)
					}
				}
			}
		default:
			if strict {
				return fmt.Errorf("cannot decode includes. got %v", v)
			}
		}
	}
	return nil
}

func (c *CIConfiguration) AppendSimpleService(svc string) {
	if c.Services.Simple == nil {
		c.Services.Simple = make([]string, 0, 1)
	}
	c.Services.Simple = append(c.Services.Simple, svc)
}

func (c *CIConfiguration) AppendService(svc *CIService) {
	if c.Services.Detailed == nil {
		c.Services.Detailed = make([]CIService, 0, 1)
	}
	c.Services.Detailed = append(c.Services.Detailed, *svc)
}

func (c *CIConfiguration) decodeServices(strict bool) (err error) {
	for _, rawSvc := range c.RawService {
		switch v := rawSvc.(type) {
		case string:
			c.AppendSimpleService(v)

		case map[interface{}]interface{}:
			svc := &CIService{}
			for rawProp, rawValue := range v {
				prop, ok := rawProp.(string)
				if !ok && strict {
					return fmt.Errorf("unrecognized service property key type. got %v", rawProp)
				} else {
					continue
				}
				var value string
				if !ok && strict {
					return fmt.Errorf("unrecognized service value type. got %v", rawValue)
				} else {
					continue
				}
				switch prop {
				case "name":
					svc.Name = value
				case "entrypoint":
					svc.Entrypoint = value
				case "command":
					svc.Command = value
				case "alias":
					svc.Alias = value
				default:
					return fmt.Errorf("unrecognized server property: %v", prop)
				}
			}
			c.AppendService(svc)
		}
	}
	return nil
}

func (c *CIConfiguration) Decode(r io.Reader, strict bool) (err error) {
	var buf bytes.Buffer
	r = io.TeeReader(r, &buf)

	// decode general structure.
	if err = yaml.NewDecoder(r).Decode(c); err != nil {
		return
	}
	if err = c.decodeInclude(strict); err != nil {
		return
	}
	// decode jobs
	r = bytes.NewReader(buf.Bytes())
	if err = yaml.NewDecoder(r).Decode(&c.Jobs); err != nil {
		if strict {
			return
		}
		err = nil
	}
	for _, name := range UnavaliableJobNames {
		delete(c.Jobs, name)
	}
	return nil
}
