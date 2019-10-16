package kubernetes

import (
	"encoding/json"
	"errors"
	"strconv"

	"git.stuhome.com/Sunmxt/wing/common"
	ccommon "git.stuhome.com/Sunmxt/wing/controller/common"
	"git.stuhome.com/Sunmxt/wing/controller/sae/operator"
	"git.stuhome.com/Sunmxt/wing/model/sae"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	typedAppsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Operator ensure application instances in kubernetes cluster.
type Operator struct {
	context       *ccommon.OperationContext
	clusterConfig *rest.Config
	namespace     string
	Error         error
}

type SpecialConfig int

const (
	Incluster = SpecialConfig(1)
)

// NewKubernetesOperator Create new kubernetes operator
func NewKubernetesOperator(config interface{}) (*Operator, error) {
	o := &Operator{}
	configGetter := (clientcmd.KubeconfigGetter)(nil)

	switch v := config.(type) {
	case *rest.Config:
		o.clusterConfig = v
	case rest.Config:
		o.clusterConfig = &v
	case *sae.Orchestrator:
		configGetter = v.KubeconfigGetter()
	case sae.Orchestrator:
		configGetter = v.KubeconfigGetter()
	case clientcmd.KubeconfigGetter:
		configGetter = v
	case SpecialConfig:
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		o.clusterConfig = restConfig
	default:
		return nil, errors.New("Invalid config type.")
	}
	if o.clusterConfig == nil {
		restConfig, err := clientcmd.BuildConfigFromKubeconfigGetter("", configGetter)
		if err != nil {
			return nil, err
		}
		o.clusterConfig = restConfig
	}
	return o, nil
}

// Namespace generate new operator for specified namespace.
func (o *Operator) Namespace(ns string) *Operator {
	new := o.clone()
	new.namespace = ns
	return new
}

func (o *Operator) clone() *Operator {
	new := &Operator{
		context:       o.context,
		clusterConfig: o.clusterConfig,
		namespace:     o.namespace,
	}
	return new
}

// Context clone and set context to new operator.
func (o *Operator) Context(ctx *ccommon.OperationContext) operator.Operator {
	new := o.clone()
	new.context = ctx
	return new
}

// Synchronize applies application deployment
func (o *Operator) Synchronize(deploy *sae.ApplicationDeployment, targetState int) (bool, error) {
	var oldSpec, newSpec *sae.ClusterSpecificationDetail
	if deploy.OldSpecificationID > 0 {
		oldSpec = &sae.ClusterSpecificationDetail{}
		if err := deploy.OldSpecification.GetSpecification(oldSpec); err != nil {
			return false, err
		}
	}
	if deploy.NewSpecificationID > 0 {
		newSpec = &sae.ClusterSpecificationDetail{}
		if err := deploy.NewSpecification.UpdateSpecification(newSpec); err != nil {
			return false, err
		}
	}
	o.context.Log.Info("start synchronize kubernetes service. cluster = %v, app = %v,%v, orchestrator = %v",
		deploy.Cluster.ID, deploy.Cluster.ApplicationID, deploy.Cluster.Application.Name, deploy.Cluster.OrchestratorID)
	if err := o.synchronizeKubernetesService(deploy.Cluster.Application); err != nil {
		return false, err
	}
	switch targetState {
	case sae.DeploymentTestingReplicaFinished:
		return o.synchronizeTestingInstanceSet(deploy, oldSpec, newSpec)

	case sae.DeploymentRollbacked:
		return o.synchronizeRollback(deploy, oldSpec, newSpec)

	case sae.DeploymentFinished:
		return o.synchronizeFullInstanceSet(deploy, oldSpec, newSpec)

	default:
		return false, errors.New("Invalid target state.")
	}

	return false, nil
}

func (o *Operator) synchronizeKubernetesService(app *sae.Application) error {
	clientSet, err := kubernetes.NewForConfig(o.clusterConfig)
	if err != nil {
		return err
	}
	var service *corev1.Service
	ctl := clientSet.CoreV1().Services(o.namespace)
	created, updated := true, false
	patchers := []ServicePatcher{
		&ServiceBasicInformationPatcher{
			Application: app,
			Type:        corev1.ServiceTypeNodePort,
		},
		&ServiceTagPatcher{
			Tags: map[string]string{
				"wing.starstudio.org/application/service-name": app.ServiceName,
				"wing.starstudio.org/application/id":           strconv.FormatInt(int64(app.Basic.ID), 10),
			},
		},
		&ServiceSelectorPatcher{
			Selector: map[string]string{
				"wing.starstudio.org/application/service-name": app.ServiceName,
				"wing.starstudio.org/application/id":           strconv.FormatInt(int64(app.Basic.ID), 10),
			},
		},
	}
	for {
		o.context.Log.Info("synchronize kubernetes service resource for service \"%v\".", app.ServiceName)
		if service, err = ctl.Get(app.ServiceName, metav1.GetOptions{}); err != nil {
			if !k8serr.IsNotFound(err) {
				o.context.Log.Info("kubernetes service resource not found for service \"%v\"", app.ServiceName)
				created = false
				service = &corev1.Service{}
			}
			return err
		}
		for _, patcher := range patchers {
			patched, err := patcher.Patch(service)
			if err != nil {
				return nil
			}
			if patched {
				updated = true
			}
		}
		if updated {
			if created {
				if service, err = ctl.Update(service); err != nil {
					if !k8serr.IsNotFound(err) {
						created = false
						service = &corev1.Service{}
						continue
					}
				}
				o.context.Log.Info("Update kubernetes service resource for service \"%v\".", app.ServiceName)
			} else {
				if service, err = ctl.Create(service); err != nil {
					if !k8serr.IsAlreadyExists(err) {
						created = true
						continue
					}
				}
				o.context.Log.Info("create kubernetes service resource for service \"%v\".", app.ServiceName)
			}
		} else {
			o.context.Log.Info("No need to update kubernetes service resource for service \"%v\".", app.ServiceName)
		}
		break
	}
	return nil
}

func (o *Operator) synchronizeTestingInstanceSet(deploy *sae.ApplicationDeployment, oldSpec, newSpec *sae.ClusterSpecificationDetail) (updated bool, err error) {
	var clientSet *kubernetes.Clientset

	if newSpec == nil { // no new specification. finished.
		return false, nil
	}

	updated = false
	switch deploy.State {
	case sae.DeploymentImageBuildFinished, sae.DeploymentTestingReplicaInProgress:
		if clientSet, err = kubernetes.NewForConfig(o.clusterConfig); err != nil {
			return false, err
		}
		if deploy.State == sae.DeploymentImageBuildFinished {
			updated = true
			deploy.State = sae.DeploymentTestingReplicaInProgress
		}

		// Swarn up new testing instance set
		var deployment *v1.Deployment
		created := true
		deployName := TestingDeploymentNameFromServiceName(deploy.Cluster.Application.ServiceName, deploy.NewSpecificationID)
		ctl := clientSet.AppsV1().Deployments(o.namespace)
		if deployment, err = ctl.Get(deployName, metav1.GetOptions{}); err != nil {
			if k8serr.IsNotFound(err) {
				deployment = &v1.Deployment{}
				created = false
			}
			return updated, err
		}

		// Do patches
		patchers := []DeploymentPatcher{
			&DeploymentBasicInformationPatcher{
				Specification:  deploy.NewSpecification,
				Detail:         newSpec,
				DeploymentName: deployName,
				ServiceName:    deploy.Cluster.Application.ServiceName,
				Namespace:      o.namespace,
			},
			&DeploymentPodTagPatcher{
				Tags: map[string]string{
					"wing.starstudio.org/application/service-name": deploy.Cluster.Application.ServiceName,
					"wing.starstudio.org/application/id":           strconv.FormatInt(int64(deploy.Cluster.ApplicationID), 10),
					"wing.starstudio.org/cluster/id":               strconv.FormatInt(int64(deploy.ClusterID), 10),
					"wing.starstudio.org/pod-role":                 "pre-release",
				},
			},
		}
		patchUpdated := false
		for _, patcher := range patchers {
			patched, err := patcher.Patch(deployment)
			if err != nil {
				return updated, err
			}
			if patched {
				patchUpdated = true
			}
		}
		if patchUpdated {
			if created {
				deployment, err = ctl.Update(deployment)
			} else {
				deployment, err = ctl.Create(deployment)
			}
			if err != nil {
				return updated, err
			}
		} else {
			allUp := false
			if allUp, err = o.synchronizeAllTestingInstanceUp(deployment); err != nil {
				return updated, err
			}
			if allUp {
				if err = o.synchronizeEliminateOldTestingInstanceLabels(ctl, deploy, oldSpec); err != nil {
					return updated, err
				}
				deploy.State = sae.DeploymentTestingReplicaFinished
				updated = true
			}
		}
	case sae.DeploymentTestingReplicaFinished:
		return false, nil

	default:
		return false, nil
	}

	return updated, nil
}

func (o *Operator) synchronizeAllTestingInstanceUp(deployment *v1.Deployment) (bool, error) {
	return deployment.Status.Replicas == deployment.Status.ReadyReplicas, nil
}

func (o *Operator) synchronizeEliminateOldTestingInstanceLabels(ctl typedAppsv1.DeploymentInterface, deploy *sae.ApplicationDeployment, oldSpec *sae.ClusterSpecificationDetail) error {
	oldDeployName := TestingDeploymentNameFromServiceName(deploy.Cluster.Application.ServiceName, deploy.NewSpecificationID)
	patch := common.PatchValue{
		Op:    "remove",
		Value: "wing.starstudio.org/application/service-name",
		Path:  "/spec/template/metadata/lables",
	}
	patchString, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	if _, err = ctl.Patch(oldDeployName, types.JSONPatchType, []byte(patchString)); err != nil && !k8serr.IsNotFound(err) {
		return err
	}
	return nil
}

func (o *Operator) synchronizeScaleOldTestingInstanceSet(ctl typedAppsv1.DeploymentInterface, deployName string, replica uint) error {
	patch := common.PatchValue{
		Op:    "replace",
		Value: replica,
		Path:  "/spec/replicas",
	}
	patchString, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	if _, err := ctl.Patch(deployName, types.JSONPatchType, []byte(patchString)); err != nil && !k8serr.IsNotFound(err) {
		return err
	}
	return nil
}

func (o *Operator) synchronizeFullInstanceSet(deploy *sae.ApplicationDeployment, olsSpec, newSpec *sae.ClusterSpecificationDetail) (bool, error) {
	switch deploy.State {
	case sae.DeploymentTestingReplicaFinished, sae.DeploymentInProgress:
		clientSet, err := kubernetes.NewForConfig(o.clusterConfig)
		if err != nil {
			return false, nil
		}

		// Scale down testing instance set
		oldTestingDeploymentName := TestingDeploymentNameFromServiceName(deploy.Cluster.Application.ServiceName, deploy.OldSpecificationID)
		if err = o.synchronizeScaleOldTestingInstanceSet(clientSet.AppsV1().Deployments(o.namespace), oldTestingDeploymentName, 0); err != nil {
			return false, nil
		}
	}
	return false, errors.New("Not implemented.")
}

func (o *Operator) synchronizeRollback(deploy *sae.ApplicationDeployment, olsSpec, newSpec *sae.ClusterSpecificationDetail) (bool, error) {
	//clientset, err := kubernetes.NewForConfig(o.clusterConfig)
	//if err != nil {
	//	return false, err
	//}
	return false, errors.New("Not implemented.")
}

// ListInstance fetches instance information.
func (o *Operator) ListInstance(app *sae.Application) ([]operator.ApplicationInstance, error) {
	return nil, nil
}

// LastError return the latest error.
func (o *Operator) LastError() error {
	return o.Error
}
