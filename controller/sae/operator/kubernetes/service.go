package kubernetes

import (
	"git.stuhome.com/Sunmxt/wing/model/sae"
	corev1 "k8s.io/api/core/v1"
)

type ServicePatcher interface {
	Patch(*corev1.Service) (bool, error)
}

type ServiceBasicInformationPatcher struct {
	Application *sae.Application
	Type        corev1.ServiceType
}

func (p *ServiceBasicInformationPatcher) Patch(service *corev1.Service) (updated bool, err error) {
	updated = false
	if service.ObjectMeta.Name != p.Application.Name {
		updated = true
		service.ObjectMeta.Name = p.Application.Name
	}
	if service.Spec.Type != p.Type {
		updated = true
		service.Spec.Type = p.Type
	}
	return
}

type ServiceTagPatcher struct {
	Tags map[string]string
}

func (p *ServiceTagPatcher) Patch(service *corev1.Service) (updated bool, err error) {
	updated = false
	if len(p.Tags) < 1 {
		return false, nil
	}
	if service.ObjectMeta.Labels == nil {
		service.ObjectMeta.Labels = map[string]string{}
		updated = true
	}
	for key, value := range p.Tags {
		origin, ok := service.ObjectMeta.Labels[key]
		if origin != value || !ok {
			service.ObjectMeta.Labels[key] = value
			updated = true
		}
	}
	return updated, err
}

type ServiceSelectorPatcher struct {
	Selector map[string]string
}

func (p *ServiceSelectorPatcher) Patch(service *corev1.Service) (updated bool, err error) {
	return updateStringToStringMap(&service.Spec.Selector, p.Selector), nil
}
