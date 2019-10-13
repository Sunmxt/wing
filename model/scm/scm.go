package scm

import (
	"encoding/json"
	"errors"
	"strconv"

	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"git.stuhome.com/Sunmxt/wing/log"
	"git.stuhome.com/Sunmxt/wing/model/account"
	"git.stuhome.com/Sunmxt/wing/model/common"
	"git.stuhome.com/Sunmxt/wing/model/scm/gitlab"
	"github.com/jinzhu/gorm"
)

const (
	UnknownSCM = 0
	GitlabSCM  = 1

	Active   = 1
	Inactive = 0
	Disabled = 2

	Shell       = 1
	ShellDocker = 2
)

var SCMPlatformTypeString map[uint]string = map[uint]string{
	UnknownSCM: "Unknown",
	GitlabSCM:  "Gitlab",
}

func Migrate(db *gorm.DB, cfg *config.WingConfiguration) (err error) {
	if err = db.AutoMigrate(&SCMPlatform{}).Error; err != nil {
		return err
	}
	db.AutoMigrate(&CIRepository{}).
		AddIndex("idx_reference", "reference").
		AddIndex("idx_scm_platform", "scm_platform_id")
	if db.Error != nil {
		return db.Error
	}
	db.AutoMigrate(&CIRepositoryApproval{}).
		AddIndex("idx_reference", "reference").
		AddIndex("idx_owner", "owner_id").
		AddIndex("idx_scm_platform", "scm_platform_id")
	db.AutoMigrate(&CIRepositoryLog{}).
		AddIndex("idx_reference", "reference")
	db.AutoMigrate(&CIRepositoryBuild{}).
		AddIndex("idx_repository_id", "repository_id")
	db.AutoMigrate(&CIRepositoryBuildProduct{}).AddUniqueIndex("idx_build_id_commit", "build_id", "commit_hash")

	if db.Error != nil {
		return db.Error
	}

	return nil
}

type SCMPlatform struct {
	common.Basic

	Active      int    `gorm:"type:tinyint;not null;"`
	Type        int    `gorm:"type:tinyint;not null;"`
	Token       string `gorm:"type:varchar(128);not null;"`
	Name        string `gorm:"type:varchar(128);not null;"`
	Description string `gorm:"type:longtext"`
	Extra       string `gorm:"type:longtext"`
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

	SCM         *SCMPlatform     `gorm:"foreignkey:SCMPlatformID;not null"`
	Owner       *account.Account `gorm:"foreignkey:OwnerID;not null"`
	Reference   string           `gorm:"type:varchar(128);not null"`
	Extra       string           `gorm:"type:longtext;not null"`
	AccessToken string           `gorm:"type:varchar(128);not null"`
	Active      int              `gorm:"type:tinyint;not null;"`

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
	Reference string `gorm:"type:varchar(255)"`
	Extra     string `gorm:"type:longtext"`
}

type CIRepositoryLogApprovalStageChangedExtra struct {
	OldStage   int `json:"old_stage"`
	NewStage   int `json:"new_stage"`
	ApprovalID int `json:"approval_id"`
}

type CIRepositoryLogBuildPackageExtra struct {
	Namespace   string `json:"namespace"`
	Environment string `json:"environment"`
	Tag         string `json:"tag"`
	Reason      string `json:"reason"`
	BuildID     int    `json:"build_id"`
	CommitHash  string `json:"commit_hash"`
}

const (
	CILogApprovalStageChanged = 1
	CILogPackageStart         = 2
	CILogPackageFailure       = 3
	CILogPackageSucceed       = 4
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

	Type        int              `gorm:"type:tinyint;not null"`
	SCM         *SCMPlatform     `gorm:"foreignkey:SCMPlatformID;not null"`
	Reference   string           `gorm:"type:varchar(128);not null"`
	Owner       *account.Account `gorm:"foreignkey:OwnerID;not null"`
	Extra       string           `gorm:"type:longtext"`
	Stage       int              `gorm:"type:tinyint;not null"`
	AccessToken string           `gorm:"type:varchar(128);"`

	SCMPlatformID int
	OwnerID       int
}

type GitlabApprovalExtra struct {
	RepositoryID         uint   `json:"repo_id"`
	WebURL               string `json:"web_url"`
	MergeRequestID       uint   `json:"mr_id"`
	InternalRepositoryID int    `json:"internal_repository_id"`
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

type CIRepositoryBuild struct {
	common.Basic

	Name         string        `gorm:"type:varchar(128);not null;"`
	Description  string        `gorm:"type:longtext;not null;"`
	ExecType     int           `gorm:"type:tinyint;not null;"`
	Extra        string        `gorm:"type:longtext;"`
	Active       int           `gorm:"type:tinyint;not null;"`
	BuildCommand string        `gorm:"type:longtext;not null;"`
	ProductPath  string        `gorm:"type:longtext;not null;"`
	Branch       string        `gorm:"type:varchar(255);not null;"`
	Repository   *CIRepository `gorm:"foreignkey:RepositoryID;not null;"`

	RepositoryID int
}

const (
	GitlabCIBuild = 1
)

type CIRepositoryBuildProduct struct {
	common.Basic

	CommitHash   string             `gorm:"type:varchar(128);not null;"`
	ProductToken string             `gorm:"type:varchar(128);not null;`
	Extra        string             `gorm:"type:longtext"`
	Active       int                `gorm:"type:tinyint;not null"`
	Stage        int                `gorm:"type:tinyint;not null"`
	Build        *CIRepositoryBuild `gorm:"foreignkey:BuildID;not null"`

	BuildID int
}

const (
	ProductBuilding     = 1
	ProductBuildSucceed = 2
	ProductBuildFailure = 3
)

type BuildProductExtra struct {
	Namespace   string `json:"namespace"`
	Environment string `json:"environment"`
	Tag         string `json:"tag"`
	LogID       int    `json:"log_id"`
}

func (l *CIRepositoryBuildProduct) EncodeExtra(extra interface{}) error {
	bin, err := json.Marshal(extra)
	if err != nil {
		return err
	}
	l.Extra = string(bin)
	return nil
}

func (l *CIRepositoryBuildProduct) DecodeExtra(extra interface{}) error {
	return json.Unmarshal([]byte(l.Extra), extra)
}
