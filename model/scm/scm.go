package scm

import (
	"encoding/json"
	"errors"
	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"git.stuhome.com/Sunmxt/wing/log"
	"git.stuhome.com/Sunmxt/wing/model/account"
	"git.stuhome.com/Sunmxt/wing/model/common"
	"git.stuhome.com/Sunmxt/wing/model/scm/gitlab"
	"github.com/jinzhu/gorm"
	"strconv"
)

const (
	UnknownSCM = 0
	GitlabSCM  = 1

	Active   = 1
	Inactive = 0

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
	db.AutoMigrate(&CIRepository{}).AddIndex("idx_reference", "reference")
	if db.Error != nil {
		return db.Error
	}
	db.AutoMigrate(&CIRepositoryApproval{}).
		AddIndex("idx_reference", "reference").
		AddIndex("idx_owner", "owner_id").
		AddIndex("idx_scm_platform", "scm_platform_id")
	db.AutoMigrate(&CIRepositoryLog{}).
		AddIndex("idx_reference", "reference")

	if db.Error != nil {
		return db.Error
	}

	return nil
}

type SCMPlatform struct {
	common.Basic

	Active      int    `gorm:"tinyint;not null;"`
	Type        int    `gorm:"tinyint;not null;"`
	Token       string `gorm:"varchar(128);not null;"`
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

func (scm *SCMPlatform) GitlabClient(logger log.DetailedLogger) (client *gitlab.GitlabClient, err error) {
	extra := scm.GitlabExtra()
	if extra == nil {
		return nil, errors.New("Not Gitlab SCM.")
	}
	if client, err = gitlab.NewGitlabClient(extra.Endpoint, logger); err != nil {
		return nil, err
	}
	client.AccessToken = extra.AccessToken
	return client, nil
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

func GetGitlabProjectReference(project *gitlab.Project) string {
	return strconv.FormatUint(uint64(project.ID), 10)
}

func (r *CIRepository) TableName() string {
	return "ci_repository"
}

func (r *CIRepository) GitlabProjectID() uint {
	id, err := strconv.ParseUint(r.Reference, 10, 64)
	if err != nil {
		return 0
	}
	return uint(id)
}

type CIRepositoryLog struct {
	common.Basic

	Type      int
	Reference string `gorm:"varchar(255)"`
	Extra     string `gorm:"longtext"`
}

type CIRepositoryLogApprovalStageChangedExtra struct {
	OldStage   int `json:"old_stage"`
	NewStage   int `json:"new_stage"`
	ApprovalID int `json:"approval_id"`
}

const (
	CILogApprovalStageChanged               = 1
	CILogRuntimeProductPackageImageUploaded = 2
)

func (l *CIRepositoryLog) EncodeExtra(extra interface{}) error {
	bin, err := json.Marshal(extra)
	if err != nil {
		return err
	}
	l.Extra = string(bin)
	return nil
}

func (l *CIRepositoryLog) DecodeExtra(extra interface{}) error {
	if err := json.Unmarshal([]byte(l.Extra), extra); err != nil {
		return err
	}
	return nil
}

type CIRepositoryApproval struct {
	common.Basic

	Type        int              `gorm:"tinyint;not null"`
	SCM         *SCMPlatform     `gorm:"foreignkey:SCMPlatformID;not null"`
	Reference   string           `gorm:"varchar(128);not null"`
	Owner       *account.Account `gorm:"foreignkey:OwnerID;not null"`
	Extra       string           `gorm:"longtext"`
	Stage       int              `gorm:"tinyint;not null"`
	AccessToken string           `gorm:"varchar(128);"`

	SCMPlatformID int
	OwnerID       int
}

type GitlabApprovalExtra struct {
	RepositoryID   uint   `json:"repo_id"`
	WebURL         string `json:"web_url"`
	MergeRequestID uint   `json:"mr_id"`
}

const (
	ApprovalRejected        = 0
	ApprovalAccepted        = 1
	ApprovalCreated         = 2
	ApprovalWaitForAccepted = 4

	GitlabMergeRequestApproval = 1
)

func (r *CIRepositoryApproval) TableName() string {
	return "ci_repository_approval"
}

func (r *CIRepositoryApproval) GitlabExtra() *GitlabApprovalExtra {
	extra := &GitlabApprovalExtra{}
	if err := json.Unmarshal([]byte(r.Extra), extra); err != nil {
		return nil
	}
	return extra
}

func (r *CIRepositoryApproval) SetGitlabExtra(extra *GitlabApprovalExtra) error {
	bin, err := json.Marshal(extra)
	if err != nil {
		return err
	}
	r.Extra = string(bin)
	return nil
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
