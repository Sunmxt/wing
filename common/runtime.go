package common

import (
	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"k8s.io/client-go/rest"
)

type WingRuntime struct {
	Config        *config.WingConfiguration
	ClusterConfig *rest.Config
}
