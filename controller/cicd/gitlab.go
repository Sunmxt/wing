package cicd

import (
	"git.stuhome.com/Sunmxt/wing/common"
	"git.stuhome.com/Sunmxt/wing/model/scm/gitlab"
	"regexp"
	"sort"
)

type GitlabCIConfigurationPatcher interface {
	Patch(*gitlab.CIConfiguration) (bool, error)
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
	return false, nil
}
