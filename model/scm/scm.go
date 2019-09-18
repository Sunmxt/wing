package scm

import (
	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"git.stuhome.com/Sunmxt/wing/model/common"
	"github.com/jinzhu/gorm"
)

const (
	UnknownSCM = 0
	GitlabSCM  = 1

	Active   = 1
	Inactive = 0

	GitlabMergeRequest = 0

	Shell       = 1
	ShellDocker = 2
)

func Migrate(db *gorm.DB, cfg *config.WingConfiguration) (err error) {
	db.AutoMigrate(&SCMPlatform{})
	if db.Error != nil {
		return db.Error
	}

	return nil
}

type SCMPlatform struct {
	common.Basic

	Active      int    `gorm:"tinyint;not null;"`
	Type        int    `gorm:"tinyint;not null;"`
	Name        string `gorm:"varchar(128);not null;"`
	Description string `gorm:"longtext"`
	Extra       string `gorm:"longtext"`
}

func (scm *SCMPlatform) TableName() string {
	return "scm_platform"
}

type GitlabSCMExtra struct {
	Endpoint   string   `json:"endpoint"`
	PublicEndpoint string `json:"public_endpoint"`
	AccessToken string   `json:"access_token"`
}

//type CIRepository struct {
//	model.Basic
//
//	SCM       *SCMPlatform `gorm:"foreignkey:SCMPlatformID;not null"`
//	Owner     *Account     `gorm:"foreignkey:OwnerID;not null"`
//	Reference string       `gorm:"varchar(128);not null"`
//	Active    int          `gorm:"tinyint;not null;"`
//
//	SCMPlatformID int
//	OwnerID       int
//}
//
//type CIRepositoryApproval struct {
//	model.Basic
//
//	Type       int           `gorm:"tinyint;not null"`
//	Extra      string        `gorm:"longtext"`
//	Repository *CIRepository `gorm:"foreignkey:RepositionID;not null"`
//
//	RepositionID int
//}
//
//type CIRepositoryBuild struct {
//	model.Basic
//
//	ExecType     int           `gorm:"tinyint;not null;"`
//	Extra        string        `gorm:"longtext;"`
//	Active       int           `gorm:"tinyint;not null;"`
//	BuildCommand string        `gorm:"longtext;not null;"`
//	ProductPath  string        `gorm:"longtext;not null;"`
//	Repository   *CIRepository `gorm:"foreignkey:RepositionID;not null;"`
//
//	RepositionID int
//}
