package kubernetes

import (
	"git.stuhome.com/Sunmxt/wing/model/sae"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"strconv"
)

type DeploymentPatcher interface {
	Patch(*v1.Deployment) (bool, error)
}

type DeploymentBasicInformationPatcher struct {
	Specification  *sae.ClusterSpecification
	Detail         *sae.ClusterSpecificationDetail
	DeploymentName string
	ServiceName    string
	Namespace      string
}

func (p *DeploymentBasicInformationPatcher) Patch(dp *v1.Deployment) (updated bool, err error) {
	if dp.ObjectMeta.Name != p.DeploymentName {
		dp.ObjectMeta.Name = p.DeploymentName
		updated = true
	}
	if dp.ObjectMeta.Namespace != p.Namespace {
		dp.ObjectMeta.Namespace = p.Namespace
		updated = true
	}
	if dp.Spec.Template.Spec.Containers == nil {
		dp.Spec.Template.Spec.Containers = make([]corev1.Container, 0)
		updated = true
	}
	if dp.Spec.Selector == nil {
		dp.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"wing.starsturio.org/role": "application-container",
			},
		}
		updated = true
	}
	if dp.Spec.Template.ObjectMeta.Labels == nil {
		dp.Spec.Template.ObjectMeta.Labels = map[string]string{
			"wing.starstudio.org/role": "application-container",
		}
		updated = true
	}
	if dp.Spec.Template.Spec.Containers == nil {
		dp.Spec.Template.Spec.Containers = make([]corev1.Container, 0)
		updated = true
	}
	var container *corev1.Container
	for _, c := dp.Spec.Template.Spec.Containers {
		if c.Name == p.ServiceName {
			container = &c
			break
		}
	}
	if container == nil {
		dp.Spec.Template.Spec.Containers = append(dp.Spec.Template.Spec.Containers, corev1.Container{
			Name: p.ServiceName,
		})
		container = &dp.Spec.Template.Spec.Containers[0]
		updated = true
	}
	if patched, patchErr := p.patchContainer(container); patchErr != nil {
		err = patchErr
	} else if patched {
		updated = true
	}
	return updated, err
}

func (p *DeploymentBasicInformationPatcher) patchContainer(c *corev1.Container) (updated bool, err error) {
	core := strconv.FormatFloat(float64(p.Detail.Resource.Core), "f", 3, 32)
	if c.Resources.Limits == nil {
		c.Resources.Limits = make(corev1.ResourceList)
	}
	cpu, _ := c.Resources.Limits[corev1.ResourceCPU]
	if string(cpu.Format) != core {
		updated = true
		cpu.Format = resource.Format(core)
		c.Resources.Limits[corev1.ResourceCPU] = cpu
	}
	space := strconv.FormatUint(uint64(p.Detail.Resource.Memory/1024/1024), 10) + "Mi"
	memory, _ := c.Resources.Limits[corev1.ResourceMemory]
	if string(memory.Format) != space {
		updated = true
		memory.Format = resource.Format(memory)
		c.Resources.Limits[corev1.ResourceMemory] = memory
	}
	if len(c.Env) != len(p.Detail.EnvironmentVariables) {
		c.Env = make([]corev1.EnvVar, 0, len(p.Detail.EnvironmentVariables))
		updated = true
	}
	indexEnvVar := map[string]int{}
	for idx, env := range c.Env {
		indexEnvVar[env.Name] = idx
	}
	for key, value := range p.Detail.EnvironmentVariables {
		idx, exists := indexEnvVar[key]
		if !exists {
			updated = true
			c.Env = append(c.Env, corev1.EnvVar{
				Name: key,
				Value: value,
			})
			continue
		}
		if c.Env[idx].Name != key {
			c.Env[idx].Name = key
			updated = true
		}
		if c.Env[idx].Value != value {
			c.Env[idx].Value = value
			updated = true
		}
	}
	return updated, nil
}

type DeploymentPodTagPatcher struct {
	Tags map[string]string
}

func (p *DeploymentPodTagPatcher) Patch(dp *v1.Deployment) (updated bool, err error) {
	updated = false
	if len(p.Tags) < 1 {
		return false, err
	}
	if dp.Spec.Template.ObjectMeta.Labels == nil {
		dp.Spec.Template.ObjectMeta.Labels = map[string]string{}
		updated = true
	}
	for key, value := range p.Tags {
		origin, ok := dp.Spec.Template.ObjectMeta.Labels[key]
		if origin != value || !ok {
			dp.Spec.Template.ObjectMeta.Labels[key] = value
			updated = true
		}
	}
	return updated, err
}
