package controller

import (
	"encoding/json"
	"fmt"
	"git.stuhome.com/Sunmxt/wing/model"
	"git.stuhome.com/Sunmxt/wing/common"
	mcommon "git.stuhome.com/Sunmxt/wing/model/common"
	"github.com/jinzhu/gorm"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (ctx *OperationContext) UpdateDeploymentManifestToTerminating(dp *appsv1.Deployment) (updated bool) {
	if dp == nil {
		return false
	}
	mark, exists := dp.ObjectMeta.Annotations["state.wing.starstudio.org"]
	if !exists || mark != "terminating" {
		updated = true
		dp.ObjectMeta.Annotations["state.wing.starstudio.org"] = "terminating"
	}
	return false
}

func valueToString(v interface{}) (result string) {
	switch t := v.(type) {
	case int, float32, float64, bool:
		result = fmt.Sprintf("%v", t)
	case string:
		result = t
	}
	return
}

func (ctx *OperationContext) UpdateAppContainerFromSpec(container *corev1.Container, spec *model.AppSpec) (updated bool, err error) {
	updated = false
	if container.Image != spec.ImageRef {
		updated = true
		container.Image = spec.ImageRef
	}
	cpucore := fmt.Sprintf("%v", spec.CPUCore)
	if cpu, _ := container.Resources.Limits[corev1.ResourceCPU]; string(cpu.Format) != cpucore {
		updated = true
		cpu.Format = resource.Format(cpucore)
	}
	memoryQ := fmt.Sprintf("%vMi", spec.Memory)
	if memory, _ := container.Resources.Limits[corev1.ResourceMemory]; string(memory.Format) != memoryQ {
		updated = true
		memory.Format = resource.Format(memoryQ)
	}
	var mapping map[string]interface{}
	// EnvVar
	if err = json.Unmarshal([]byte(spec.EnvVar), &mapping); err != nil {
		return
	}
	if len(mapping) != len(container.Env) {
		container.Env = make([]corev1.EnvVar, 0, len(mapping))
		updated = true
	}
	envByName := map[string]int{}
	for idx, env := range container.Env {
		envByName[env.Name] = idx
	}
	for k, raw := range mapping {
		idx, exists := envByName[k]
		if !exists {
			updated = true
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  k,
				Value: valueToString(raw),
			})
			continue
		}
		env := container.Env[idx]
		if env.Name != k {
			env.Name = k
			updated = true
		}
		if env.Value != valueToString(raw) {
			env.Value = valueToString(raw)
			updated = true
		}
	}
	var listString []string
	if err = json.Unmarshal([]byte(spec.Command), &listString); err != nil {
		return false, err
	}
	if len(container.Command) != len(listString) {
		updated = true
		container.Command = listString
	} else {
		for idx := range container.Command {
			if container.Command[idx] != listString[idx] {
				container.Command = listString
				updated = true
				break
			}
		}
	}
	listString = nil
	if err = json.Unmarshal([]byte(spec.Args), &listString); err != nil {
		return false, err
	}
	if len(container.Args) != len(listString) {
		updated = true
		container.Args = listString
	} else {
		for idx := range container.Command {
			if container.Args[idx] != listString[idx] {
				container.Args = listString
				updated = true
				break
			}
		}
	}
	return updated, nil
}

func (ctx *OperationContext) CreateDeploymentManifestFromSpec(appName string, deployID int, spec *model.AppSpec) (dp *appsv1.Deployment, err error) {
	dp = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: common.GetNormalizedDeploymentName(appName, deployID),
		},
	}
	if _, err = ctx.UpdateDeploymentManifestFromSpec(appName, deployID, dp, spec); err != nil {
		ctx.Log.Error("Failed to update deployment manifest: " + err.Error())
	}
	return
}

func (ctx *OperationContext) DetermineDeploymentState(dp *appsv1.Deployment, state int) (currentState int) {
	mark, exists := "", false
	if dp != nil {
		mark, exists = dp.ObjectMeta.Annotations["state.wing.starstudio.org"]
	}

	currentState = state
	switch state {
	case model.Waiting, model.Terminated:
		if dp != nil { // Deployment existing. determine exactly state.
			if exists && mark == "terminating" {
				currentState = model.Terminating
			} else {
				currentState = model.Executed
			}
		}
	case model.Executed:
		if dp == nil { // Deployment not found. reset to waiting.
			currentState = model.Waiting
		}
	case model.Terminating:
		if dp == nil { // Deployment not found. set to terminated.
			currentState = model.Terminated
		}
	}
	return
}

func (ctx *OperationContext) SyncDeploymentState(deploy *model.Deployment, dp *appsv1.Deployment) (currentState int, err error) {
	var db *gorm.DB

	currentState = ctx.DetermineDeploymentState(dp, deploy.State)
	if deploy.State != currentState {
		ctx.Log.Infof("[SyncDeploymentState] State of deployment \"%v\" is determined as %v instead of %v.", deploy.App.Name, currentState, deploy.State)
		if db, err = ctx.Database(); err != nil {
			return
		}
		deploy.State = currentState
		if err = db.Save(deploy).Error; err != nil {
			return
		}
	}
	return
}

func (ctx *OperationContext) UpdateDeploymentManifestFromSpec(appName string, deployID int, manifest *appsv1.Deployment, spec *model.AppSpec) (updated bool, err error) {
	if manifest == nil || spec == nil {
		return false, nil
	}
	updated = false
	deployName := common.GetNormalizedDeploymentName(appName, deployID)
	if manifest.ObjectMeta.Name != deployName {
		updated = true
		manifest.ObjectMeta.Name = deployName
	}
	if manifest.ObjectMeta.Namespace != ctx.Runtime.Config.Kube.Namespace {
		updated = true
		manifest.ObjectMeta.Namespace = ctx.Runtime.Config.Kube.Namespace
	}
	if manifest.Spec.Replicas == nil {
		manifest.Spec.Replicas = new(int32)
	}
	if *manifest.Spec.Replicas != int32(spec.Replica) {
		updated = true
		*manifest.Spec.Replicas = int32(spec.Replica)
	}
	if manifest.Spec.Template.Spec.Containers == nil {
		updated = true
		manifest.Spec.Template.Spec.Containers = make([]corev1.Container, 0)
	}
	if manifest.Spec.Selector == nil {
		manifest.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"role.wing.starstudio.org": "app-container",
			},
		}
		updated = true
	}
	if manifest.Spec.Template.ObjectMeta.Labels == nil {
		manifest.Spec.Template.ObjectMeta.Labels = map[string]string{
			"role.wing.starstudio.org": "app-container",
		}
		updated = true
	}
	deployIDString := fmt.Sprintf("%v", deployID)
	if curDeployID, ok := manifest.Spec.Template.ObjectMeta.Labels["deploy.wing.starstudio.org"]; !ok || curDeployID != deployIDString {
		manifest.Spec.Template.ObjectMeta.Labels["deploy.wing.starstudio.org"] = deployIDString
		updated = true
	}
	if appLabel, ok := manifest.Spec.Template.ObjectMeta.Labels["app.wing.starstudio.org"]; !ok || appLabel != appName {
		manifest.Spec.Template.ObjectMeta.Labels["app.wing.starstudio.org"] = appName
		updated = true
	}

	var container *corev1.Container
	for _, c := range manifest.Spec.Template.Spec.Containers {
		if c.Name == "application" {
			container = &c
			break
		}
	}
	if container == nil {
		manifest.Spec.Template.Spec.Containers = append(manifest.Spec.Template.Spec.Containers, corev1.Container{
			Name: "application",
		})
		container = &manifest.Spec.Template.Spec.Containers[0]
		updated = true
	}
	return ctx.UpdateAppContainerFromSpec(container, spec)
}

func (ctx *OperationContext) SyncDeployment(ID, targetState int) (deploy *model.Deployment, synced bool, err error) {
	var clientset *kubernetes.Clientset
	var dp *appsv1.Deployment
	var currentState int

	db, err := ctx.Database()
	if err != nil {
		return nil, false, err
	}
	deploy = &model.Deployment{
		Basic: mcommon.Basic{
			ID: ID,
		},
	}
	if clientset, err = ctx.KubeClient(); err != nil {
		return nil, false, err
	}
	if err = db.Where(deploy).Preload("App").Preload("Spec").First(deploy).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = nil
		} else {
			ctx.Log.Error("Deployment not loaded:" + err.Error())
		}
		return nil, false, err
	}

	kubeDeployName := common.GetNormalizedDeploymentName(deploy.App.Name, deploy.Basic.ID)
	ctx.Log.Infof("[SyncDeployment] Start deployment to kubernetes. (Application ID = %v, Name = %v, Normalized name = %v, Target state = %v)", ID, deploy.App.Name, kubeDeployName, targetState)

	// stage 1: determine current state
	ctl := clientset.AppsV1().Deployments(ctx.Runtime.Config.Kube.Namespace)
	if dp, err = ctl.Get(kubeDeployName, metav1.GetOptions{}); err != nil {
		if !k8serr.IsNotFound(err) {
			ctx.Log.Error("Failed to get deploy from kubernetes: " + err.Error())
			return deploy, false, err
		} else {
			dp = nil
		}
	}
	if currentState, err = ctx.SyncDeploymentState(deploy, dp); err != nil {
		return nil, false, err
	}

	// stage 2: sync to cluster
	updated := false
	switch currentState {
	case model.Waiting:
		switch targetState {
		case model.Waiting:
			return deploy, true, nil // Not altered.

		case model.Executed:
			if dp, err = ctx.CreateDeploymentManifestFromSpec(deploy.App.Name, deploy.Basic.ID, deploy.Spec); err != nil {
				return deploy, false, err
			}
			if dp, err = ctl.Create(dp); err != nil {
				return deploy, false, err
			}
			deploy.State = model.Executed
			updated = true

		case model.Terminated, model.Terminating:
			deploy.State, updated = model.Terminated, true

		default:
			return deploy, false, nil // Invalid state target.
		}

	case model.Executed:
		switch targetState {
		case model.Waiting:
			return deploy, false, nil // Invalid state target.

		case model.Executed:
			if dp != nil {
				// ensure manifest corrent
				if synced, err = ctx.UpdateDeploymentManifestFromSpec(deploy.App.Name, deploy.Basic.ID, dp, deploy.Spec); err != nil {
					return deploy, false, err
				}
				if synced {
					if dp, err = ctl.Update(dp); err != nil {
						return deploy, false, err
					}
				}
			} else {
				if dp, err = ctx.CreateDeploymentManifestFromSpec(deploy.App.Name, deploy.Basic.ID, deploy.Spec); err != nil {
					return deploy, false, err
				}
				if dp, err = ctl.Create(dp); err != nil {
					return deploy, false, err
				}
			}

		case model.Terminating:
			if ctx.UpdateDeploymentManifestToTerminating(dp) {
				if dp, err = ctl.Update(dp); err != nil {
					return deploy, false, err
				}
			}
			deploy.State, updated = model.Terminating, true

		case model.Finished:
			deploy.State, updated = model.Finished, true
		default:
			return deploy, false, nil
		}

	case model.Terminating:
		switch targetState {
		case model.Terminating:
			if dp == nil {
				deploy.State, updated = model.Terminated, true
			} else {
				if ctx.UpdateDeploymentManifestToTerminating(dp) {
					if dp, err = ctl.Update(dp); err != nil {
						return deploy, false, err
					}
				}
			}

		case model.Terminated:
			if dp != nil {
				if err = ctl.Delete(deploy.App.Name, &metav1.DeleteOptions{}); err != nil {
					return deploy, false, err
				}
			}
			deploy.State, updated = model.Terminated, true

		default:
			return deploy, false, err
		}

	case model.Terminated:
		if targetState != model.Terminated {
			return deploy, false, nil
		}
	case model.Finished:
		if targetState != model.Finished {
			return deploy, false, nil
		}
	}
	if updated {
		if err = db.Save(deploy).Error; err != nil {
			return deploy, false, err
		}
	}

	return deploy, true, nil
}

//func (ctx *OperationContext) GetDeploymentByID(deploymentID int) (*model.Deployment, *appsv1.Deployment, error) {
//	db, err := ctx.Database()
//	if err != nil {
//		return nil, nil, err
//	}
//	deploy := &model.Deployment{
//		Basic: model.Basic{
//			ID: deploymentID,
//		},
//	}
//	if err = db.Where(deploy).Preload("Spec").Preload("App").First(deploy).Error; err != nil {
//		if gorm.IsRecordNotFoundError(err) {
//			return nil, nil, nil
//		}
//	}
//	if dp, err = ctx.GetKubeDeploymentManifest(deploy.App.Name, deploy.Basic.ID); err != nil {
//		return nil, nil, err
//	}
//	currentState := ctx.DetermineDeploymentState(dp, deploy.State)
//	if currentState != deploy.State {
//		deploy.State = currentState
//		if err = db.Save(deploy).Error; err != nil {
//			return nil, nil, err
//		}
//	}
//	return deploy, dp, nil
//}

func (ctx *OperationContext) GetKubeDeploymentManifest(name string, deploymentID int) (dp *appsv1.Deployment, err error) {
	var clientset *kubernetes.Clientset
	if clientset, err = kubernetes.NewForConfig(ctx.Runtime.ClusterConfig); err != nil {
		return nil, err
	}

	kDeployName := common.GetNormalizedDeploymentName(name, deploymentID)
	if dp, err = clientset.AppsV1().Deployments(ctx.Runtime.Config.Kube.Namespace).Get(kDeployName, metav1.GetOptions{}); err != nil {
		return nil, err
	}
	return dp, nil
}

func (ctx *OperationContext) GetCurrentDeployment(appID int) (deploy *model.Deployment, dp *appsv1.Deployment, err error) {
	var db *gorm.DB

	db, err = ctx.Database()
	if err != nil {
		return nil, nil, err
	}

	deploy = &model.Deployment{
		AppID: appID,
	}
	if err = db.Where("State not in (?)", []int{model.Terminated, model.Finished}).
		Where(deploy).Order("id desc").Preload("Spec").Preload("App").
		First(deploy).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil, nil
		}
	}
	if deploy.State == model.Waiting {
		return deploy, nil, nil
	}

	if dp, err = ctx.GetKubeDeploymentManifest(deploy.App.Name, deploy.Basic.ID); err != nil {
		return nil, nil, err
	}
	if dp == nil {
		// Unmatched state detected. Sync to database.
		deploy.State = model.Waiting
		if err = db.Save(deploy).Error; err != nil {
			return nil, nil, err
		}
	}

	return deploy, dp, nil
}

func (ctx *OperationContext) ListApplicationPodInfo(appName string, deploymentID int) (podInfo []common.ApplicationPodInfo, err error) {
	if deploymentID < 1 {
		return nil, nil
	}
	var clientset *kubernetes.Clientset
	var podList *corev1.PodList
	if clientset, err = kubernetes.NewForConfig(ctx.Runtime.ClusterConfig); err != nil {
		return nil, err
	}
	ctl := clientset.CoreV1().Pods(ctx.Runtime.Config.Kube.Namespace)
	if podList, err = ctl.List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploy.wing.starstudio.org=%v, app.wing.starstudio.org=%v", deploymentID, appName),
	}); err != nil {
		return
	}
	// Pick information
	podInfo = make([]common.ApplicationPodInfo, len(podList.Items))
	for idx, pod := range podList.Items {
		podInfo[idx].FromManifest(&pod)
	}
	return
}
