package common

import (
	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"github.com/RichardKnop/machinery/v1"
	"k8s.io/client-go/rest"
)

type WingRuntime struct {
	Config        *config.WingConfiguration
	ClusterConfig *rest.Config
	MachineID     string
	JobServer     *machinery.Server
}
