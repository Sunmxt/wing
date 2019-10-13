package operator

import (
	"git.stuhome.com/Sunmxt/wing/controller/common"
	"git.stuhome.com/Sunmxt/wing/model/sae"
)

// ApplicationInstance describes instance information details.
type ApplicationInstance interface {
	Name() string
	PortMap() map[uint16]uint16
	State() uint
}

const (
	InstanceWaitForCreating = 1
	InstanceCreating        = 2
	InstanceRunning         = 3
	InstanceShuttingDown    = 4
	InstanceDowned          = 5
)

// Operator deploy application in cluster.
type Operator interface {
	// Use operation context
	Context(*common.OperationContext) Operator

	// Apply deployment
	Synchronize(*sae.ApplicationDeployment, int) (bool, error)

	// List application instance in cluster.
	ListInstance(*sae.Application) ([]ApplicationInstance, error)

	// Get last error
	LastError() error
}
