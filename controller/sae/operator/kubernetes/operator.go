package kubernetes

import (
	"errors"

	"git.stuhome.com/Sunmxt/wing/controller/common"
	"git.stuhome.com/Sunmxt/wing/controller/sae/operator"
	"git.stuhome.com/Sunmxt/wing/model/sae"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Operator ensure application instances in kubernetes cluster.
type Operator struct {
	context       *common.OperationContext
	clusterConfig *rest.Config
	Error         error
}

type SpecialConfig int

const (
	Incluster = SpecialConfig(1)
)

// NewKubernetesOperator Create new kubernetes operator
func NewKubernetesOperator(config interface{}) (operator.Operator, error) {
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

func (o *Operator) clone() *Operator {
	new := &Operator{
		context:       o.context,
		clusterConfig: o.clusterConfig,
	}
	return new
}

// Context clone and set context to new operator.
func (o *Operator) Context(ctx *common.OperationContext) operator.Operator {
	new := o.clone()
	new.context = ctx
	return new
}

// Synchronize applies application deployment
func (o *Operator) Synchronize(deploy *sae.ApplicationDeployment, targetState int) (bool, error) {
	clientset, err := kubernetes.NewForConfig(o.clusterConfig)
	if err != nil {
		return false, err
	}
	return false, nil
}

// ListInstance fetches instance information.
func (o *Operator) ListInstance(app *sae.Application) ([]operator.ApplicationInstance, error) {
	return nil, nil
}

// LastError return the latest error.
func (o *Operator) LastError() error {
	return o.Error
}
