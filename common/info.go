package common

import (
	corev1 "k8s.io/api/core/v1"
)

type ApplicationPodInfo struct {
	Name         string                   `json:"name"`
	PodIP        string                   `json:"pod_ip"`
	PortMap      []ApplicationPortMapInfo `json:"port_maps"`
	State        int                      `json:"state"`
	RestartCount int                      `json:"restart_count"`
}

type ApplicationPortMapInfo struct {
	External uint `json:"external"`
	Internal uint `json:"internal"`
}

func (m *ApplicationPodInfo) FromManifest(pod *corev1.Pod) {
	var (
		container  *corev1.Container
		mainStatus *corev1.ContainerStatus
	)

	m.PodIP = pod.Status.PodIP
	for _, c := range pod.Spec.Containers {
		if c.Name == "application" {
			container = &c
		}
	}
	if container != nil {
		m.PortMap = make([]ApplicationPortMapInfo, len(container.Ports))
		for idx, port := range container.Ports {
			m.PortMap[idx].External = uint(port.HostPort)
			m.PortMap[idx].Internal = uint(port.ContainerPort)
		}
	}
	m.Name, _ = pod.ObjectMeta.Labels["app.wing.starstudio.org"]
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == "application" {
			mainStatus = &status
			break
		}
	}
	if mainStatus == nil {
		m.State = PodStateUnknown
	} else if mainStatus.Ready {
		m.State = PodReady
	} else if mainStatus.State.Waiting != nil {
		m.State = PodWaitingForSchedule
	} else if mainStatus.State.Running != nil {
		m.State = PodScheduled
	} else if mainStatus.State.Terminated != nil {
		m.State = PodTerminated
	} else {
		m.State = PodStateUnknown
	}
}
