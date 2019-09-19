package scm

import (
	"encoding/json"
	"errors"
	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"git.stuhome.com/Sunmxt/wing/model/account"
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

var SCMPlatformTypeString map[uint]string = map[uint]string{
	UnknownSCM: "Unknown",
	GitlabSCM:  "Gitlab",
}

func Migrate(db *gorm.DB, cfg *config.WingConfiguration) (err error) {
	db.AutoMigrate(&SCMPlatform{})
	if db.Error != nil {
		return db.Error
	}
	db.AutoMigrate(&CIRepository{})
	if db.Error != nil {
		return db.Error
	}
	db.AutoMigrate(&CIRepositoryApproval{})
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

type GitlabSCMExtra struct {
	Endpoint       string `json:"endpoint"`
	PublicEndpoint string `json:"public_endpoint"`
	AccessToken    string `json:"access_token"`
}

func (scm *SCMPlatform) TableName() string {
	return "scm_platform"
}

func (scm *SCMPlatform) GitlabExtra() (extra *GitlabSCMExtra) {
	extra = &GitlabSCMExtra{}
	if err := json.Unmarshal([]byte(scm.Extra), extra); err != nil {
		return nil
	}
	return extra
}

func (scm *SCMPlatform) GitlabProjectQuery() (query *GitlabProjectQuery, err error) {
	extra := scm.GitlabExtra()
	if extra == nil {
		return nil, errors.New("Not Gitlab SCM.")
	}
	if query, err = NewGitlabProjectQuery(extra.Endpoint); err != nil {
		return nil, err
	}
	query.AccessToken = extra.AccessToken
	return
}

func (scm *SCMPlatform) PublicURL() string {
	switch scm.Type {
	case GitlabSCM:
		extra := scm.GitlabExtra()
		if extra == nil {
			return ""
		}
		return extra.PublicEndpoint
	}

	return ""
}

type CIRepository struct {
	common.Basic

	SCM       *SCMPlatform     `gorm:"foreignkey:SCMPlatformID;not null"`
	Owner     *account.Account `gorm:"foreignkey:OwnerID;not null"`
	Reference string           `gorm:"varchar(128);not null"`
	Active    int              `gorm:"tinyint;not null;"`

	SCMPlatformID int
	OwnerID       int
}

func (r *CIRepository) TableName() string {
	return "ci_repository"
}

type CIRepositoryApproval struct {
	common.Basic

	Type        int           `gorm:"tinyint;not null"`
	Extra       string        `gorm:"longtext"`
	Repository  *CIRepository `gorm:"foreignkey:RepositionID;not null"`
	AccessToken string        `gorm:"varchar(128);"`

	RepositionID int
}

func (r *CIRepositoryApproval) TableName() string {
	return "ci_repository_approval"
}

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
