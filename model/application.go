package model

const (
	ReadyToDeploy = 0
)

type AppSpec struct {
	Basic

	ImageRef string  `gorm:"type:varchar(128);not null"`
	EnvVar   string  `gorm:"type:longtext;not null"`
	Memory   uint64  `gorm:"not null"`
	CPUCore  float32 `gorm:"not null"`
	Command  string  `gorm:"type:longtext;not null"`
	Args     string  `gorm:"type:longtext;not null"`
	Replica  int     `gorm:"not null"`
}

func (m AppSpec) TableName() string {
	return "application_spec"
}

type Application struct {
	Basic

	Name      string   `gorm:"type:varchar(128);not null;unique"`
	Owner     *Account `gorm:"foreignkey:OwnerID;not null"`
	OwnerID   int
	Extra     string   `gorm:"type:longtext"`
	KubeLabel string   `gorm:"type:longtext;not null"`
	Spec      *AppSpec `gorm:"foreignkey:SpecID;not null"`
	SpecID    int
}

func (m Application) TableName() string {
	return "application"
}
