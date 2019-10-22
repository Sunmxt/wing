package sae

import (
	"encoding/json"

	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"git.stuhome.com/Sunmxt/wing/model/account"
	"git.stuhome.com/Sunmxt/wing/model/common"
	"git.stuhome.com/Sunmxt/wing/model/scm"
	"github.com/jinzhu/gorm"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func Migrate(db *gorm.DB, cfg *config.WingConfiguration) (err error) {
	if err = db.AutoMigrate(&Orchestrator{}).Error; err != nil {
		return err
	}
	if err = db.AutoMigrate(&Application{}).Error; err != nil {
		return err
	}
	if err = db.AutoMigrate(&BuildDependency{}).Error; err != nil {
		return err
	}
	if err = db.AutoMigrate(&ApplicationCluster{}).Error; err != nil {
		return err
	}
	if err = db.AutoMigrate(&ClusterSpecification{}).Error; err != nil {
		return err
	}
	if err = db.AutoMigrate(&ApplicationDeployment{}).Error; err != nil {
		return err
	}
	return nil
}

type Orchestrator struct {
	common.Basic `json:"-"`

	Name   string           `gorm:"type:varchar(128);" json:"name"`
	Active int              `gorm:"type:tinyint;not null;" json:"active"`
	Type   int              `gorm:"type:tinyint;not null;" json:"type"`
	Extra  string           `gorm:"type:longtext" json:"extra"`
	Owner  *account.Account `gorm:"foreignkey:OwnerID;not null" json:"-"`

	OwnerID int `json:"owner_id"`
}

func (o *Orchestrator) DecodeExtra(v interface{}) error {
	return common.DecodeExtra(o.Extra, v)
}

func (o *Orchestrator) EncodeExtra(v interface{}) error {
	return common.EncodeExtra(&o.Extra, v)
}

func (o Orchestrator) TableName() string {
	return "orchestrator"
}

type KubernetesOrchestrator struct {
	Namespace  string `json:"namespace"`
	Kubeconfig string `json:"kubeconfig"`
}

const (
	Kubernetes          = 1
	KubernetesIncluster = 2
)

func (o *Orchestrator) KubeconfigGetter() clientcmd.KubeconfigGetter {
	return func() (*clientcmdapi.Config, error) {
		config := KubernetesOrchestrator{}
		if err := o.DecodeExtra(config); err != nil {
			return nil, err
		}
		kubeConfig := &clientcmdapi.Config{}
		if err := json.Unmarshal([]byte(config.Kubeconfig), kubeConfig); err != nil {
			return nil, err
		}
		return kubeConfig, nil
	}
}

type Application struct {
	common.Basic `json:"-"`

	Name        string           `gorm:"type:varchar(128);not null" json:"name"`
	ServiceName string           `gorm:"type:varchar(128);not null;unique;" json:"service_name"`
	Owner       *account.Account `gorm:"foreignkey:OwnerID;not null" json:"-"`
	Description string           `gorm:"type:longtext;"`
	Extra       string           `gorm:"type:longtext" json:"extra"`

	OwnerID int `json:"owner_id"`
}

func (m Application) TableName() string {
	return "application"
}

type BuildDependency struct {
	common.Basic `json:"-"`

	Build       *scm.CIRepositoryBuild `gorm:"foreignkey:BuildID" json:"-"`
	Application *Application           `gorm:"foreignkey:ApplicationID" json:"-"`
	Extra       string                 `gorm:"type:longtext" json:"extra"`

	ApplicationID int `json:"application_id"`
	BuildID       int `json:"build_id"`
}

func (m BuildDependency) TableName() string {
	return "build_dependency"
}

type ApplicationCluster struct {
	common.Basic

	Application   *Application          `gorm:"foreignkey:ApplicationID"`
	Orchestrator  *Orchestrator         `gorm:"foreignkey:OrchestratorID"`
	Specification *ClusterSpecification `gorm:"foreignkey:SpecID"`
	Active        int                   `gorm:"type:tinyint"`

	OrchestratorID  int
	SpecificationID int
	ApplicationID   int
}

func (m ApplicationCluster) TableName() string {
	return "application_cluster"
}

type ClusterSpecification struct {
	common.Basic

	Extra string `gorm:"longtext"`
}

func (s *ClusterSpecification) GetSpecification(spec *ClusterSpecificationDetail) error {
	return common.DecodeExtra(s.Extra, spec)
}

func (s *ClusterSpecification) UpdateSpecification(spec *ClusterSpecificationDetail) error {
	return common.EncodeExtra(&s.Extra, spec)
}

type ClusterSpecificationDetail struct {
	Command              string               `json:"command"`
	ReplicaCount         int                  `json:"replica"`
	TestingReplicaCount  int                  `json:"testing_replica"`
	EnvironmentVariables map[string]string    `json:"environment_variables"`
	Resource             *ResourceRequirement `json:"resource"`
	Product              []ProductRequirement `json:"product"`
	BaseImage            string               `json:"base_image"`
}

type ResourceRequirement struct {
	Core   float32 `json:"core"`
	Memory uint64  `json:"memory"`
}

type ProductRequirement struct {
	Namespace   string `json:"namespace"`
	Environment string `json:"environment"`
	Tag         string `json:"tag"`
}

func (m ClusterSpecification) TableName() string {
	return "cluster_specification"
}

type ApplicationDeployment struct {
	common.Basic

	Cluster          *ApplicationCluster   `gorm:"foreignkey:ClusterID"`
	OldSpecification *ClusterSpecification `gorm:"foreignkey:OldSpecificationID"`
	NewSpecification *ClusterSpecification `gorm:"foreignkey:NewSpecificationID"`
	State            int                   `gorm:"type:tinyint"`

	OldSpecificationID int
	NewSpecificationID int
	ClusterID          int
}

const (
	DeploymentFinished                 = 0
	DeploymentCanceled                 = 1
	DeploymentRollbacked               = 2
	DeploymentCreated                  = 3
	DeploymentImageBuildInProgress     = 4
	DeploymentImageBuildFinished       = 5
	DeploymentTestingReplicaInProgress = 6
	DeploymentTestingReplicaFinished   = 7
	DeploymentInProgress               = 8
	DeploymentRollbackInProgress       = 9
)

var ActiveDeploymentStates []int = []int{
	DeploymentCreated,
	DeploymentImageBuildFinished,
	DeploymentImageBuildInProgress,
	DeploymentTestingReplicaFinished,
	DeploymentTestingReplicaInProgress,
	DeploymentInProgress,
	DeploymentRollbackInProgress,
}

func (m ApplicationDeployment) TableName() string {
	return "application_deployment"
}
