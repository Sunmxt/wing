package model

const (
	ReadyToDeploy = 0
)

type AppSpec struct {
	Basic

	ImageRef string `gorm:"not null"`
	EnvVar   string `gorm:"not null"`
	Memory   uint64 `gorm:"not null"`
	CPUCore  int    `gorm:"not null"`
	Command  string `gorm:"not null"`
	Args     string `gorm:"not null"`
	Replica  int    `gorm:"not null"`
}

func (m AppSpec) TableName() string {
	return "application_spec"
}

type Application struct {
	Basic

	Name      string   `gorm:"not null;unique"`
	Owner     *Account `gorm:"foreignkey:OwnerID;not null"`
	OwnerID   int
	Extra     string
	KubeLabel string   `gorm:"not null"`
	Spec      *AppSpec `gorm:"foreignkey:SpecID;not null"`
	SpecID    int
}

func (m Application) TableName() string {
	return "application"
}
