package cicd

import (
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/model/scm/gitlab"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type GitlabCIConfigurationPatcher interface {
	Patch(*gitlab.CIConfiguration) (bool, error)
}

// Chain Patcher ensure that the patcher B runs only after the patcher A applies patches.
type GitlabChainPatcher struct {
	A GitlabCIConfigurationPatcher
	B GitlabCIConfigurationPatcher
}

func (p *GitlabChainPatcher) invoke(patcher GitlabCIConfigurationPatcher, cfg *gitlab.CIConfiguration) (bool, error) {
	if patcher == nil {
		return false, nil
	}
	return patcher.Patch(cfg)
}
func (p *GitlabChainPatcher) Patch(cfg *gitlab.CIConfiguration) (bool, error) {
	if p.A == nil {
		return p.invoke(p.B, cfg)
	}
	updated, err := p.A.Patch(cfg)
	if err != nil {
		return updated, err
	}
	if updated {
		_, err := p.invoke(p.B, cfg)
		return true, err
	}
	return false, err
}

// gitlab CI rules to load runtime from remote
type GitlabRemoteRuntimeLoad struct {
	RuntimeURL string
}

func (p *GitlabRemoteRuntimeLoad) MakeBash() string {
	// ensure dir
	cmd := "rm -rf /run/sear; mkdir /run/saer -p;"
	// download and extraxt runtime scripts.
	cmd += "( which curl && curl -L '" + p.RuntimeURL + "' || wget -O - '" + p.RuntimeURL + "' ) |"
	cmd += "tar -zxv -C /run/sear;"
	// load runtime
	cmd += "source /run/sear/bin/sar_activate"
	return cmd
}

func (p *GitlabRemoteRuntimeLoad) Patch(cfg *gitlab.CIConfiguration) (bool, error) {
	escaped, updated := common.EscapeForRegexp(p.RuntimeURL), false
	checker, err := regexp.Compile("rm.*mkdir.*curl.*" + escaped + "source.*sar_activate")
	if err != nil {
		return false, err
	}
	if cfg.BeforeScript == nil {
		cfg.BeforeScript = make([]string, 0, 1)
	}
	updated = true
	for _, line := range cfg.BeforeScript {
		if checker.FindIndex([]byte(line)) != nil {
			updated = false
			break
		}
	}
	if updated { // Not found.
		cfg.BeforeScript = append(cfg.BeforeScript, p.MakeBash())
	}
	return updated, nil
}

// gitlab CI rules to include dynmaic jobs
type GitlabDynamicJob struct {
	JobURL string
}

func (p *GitlabDynamicJob) Patch(cfg *gitlab.CIConfiguration) (bool, error) {
	updated := true
	checker, err := regexp.Compile(common.EscapeForRegexp(p.JobURL))
	if err != nil {
		return false, err
	}
	for _, ref := range cfg.Include.Remote {
		if checker.FindIndex([]byte(ref)) != nil {
			updated = false
			break
		}
	}
	if updated {
		cfg.AppendRemoteInclude(p.JobURL)
	}
	return updated, nil
}

// gitlab CI rules to ensure wing build stage
type GitlabWingStages struct{}

func (p *GitlabWingStages) Patch(cfg *gitlab.CIConfiguration) (updated bool, err error) {
	if cfg.Stages == nil {
		cfg.Stages = make([]string, 0, 3)
	}
	minIdx := 0
	stages := map[string]int{ // map stage to priority
		"test":      1,
		"build":     2,
		"integrate": 3,
	}
	set := common.NewStringSet()
	for stage, _ := range stages {
		set.Add(stage)
	}
	for idx, stage := range cfg.Stages {
		if set.In(stage) {
			minIdx = idx
			break
		}
	}
	set.Delete(cfg.Stages...)
	if set.Len() > 0 {
		stages := cfg.Stages[:minIdx]
		set.Visit(func(stage string) bool {
			stages = append(stages, stage)
			return true
		})
		stages = append(stages, cfg.Stages[minIdx:]...)
		cfg.Stages = stages
		updated = true
	}
	sort.Slice(cfg.Stages, func(i, j int) bool {
		left, ok := stages[cfg.Stages[i]]
		if !ok {
			return true
		}
		var right int
		right, ok = stages[cfg.Stages[j]]
		if !ok || left < right {
			return true
		}
		updated = true
		return false
	})
	return updated, nil
}

// gitlab CI rules to ensure services required by dymanic jobs.
type GitlabWingBuildServices struct{}

func (p *GitlabWingBuildServices) Patch(cfg *gitlab.CIConfiguration) (bool, error) {
	for _, svc := range cfg.Services.Simple {
		if strings.Index("dind", svc) >= 0 {
			return false, nil
		}
	}
	for _, svc := range cfg.Services.Detailed {
		if strings.Index("dind", svc.Name) >= 0 {
			return false, nil
		}
	}
	cfg.AppendSimpleService("docker:dind")
	return true, nil
}

// gitlab CI rules to ensure global variables exists.
type GitlabGlobalVariables struct {
	Variables map[string]interface{}
}

func (p *GitlabGlobalVariables) Patch(cfg *gitlab.CIConfiguration) (bool, error) {
	updated := false
	if p.Variables == nil {
		return false, nil
	}
	if cfg.Variables == nil {
		cfg.Variables = make(map[string]interface{})
	}
	for k, v := range p.Variables {
		origin, exists := cfg.Variables[k]
		if exists && reflect.DeepEqual(origin, v) {
			continue
		}
		cfg.Variables[k] = v
		updated = true
	}
	return updated, nil
}
