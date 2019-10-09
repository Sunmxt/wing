package scm

import (
	"github.com/jinzhu/gorm"
	"strconv"
)

func pickByColumn(db *gorm.DB, columnName string, elem interface{}, val interface{}) (err error) {
	err = db.Model(elem).Where(columnName+"= ?", val).First(elem).Error
	if gorm.IsRecordNotFoundError(err) {
		err = nil
	}
	return
}

func (p *SCMPlatform) ByID(db *gorm.DB, platformID int) error {
	return pickByColumn(db, "id", p, platformID)
}

func (a *CIRepositoryApproval) ByID(db *gorm.DB, approvalID int) error {
	return pickByColumn(db, "id", a, approvalID)
}

func (b *CIRepositoryBuild) ByID(db *gorm.DB, buildID int) error {
	return pickByColumn(db, "id", b, buildID)
}

func (a *CIRepositoryApproval) ByReference(db *gorm.DB, reference string) error {
	return pickByColumn(db, "reference", a, reference)
}

func ReferenceLogApprovalStageChanged(platformID, repositoryID int) string {
	return "approval:" + strconv.FormatInt(int64(platformID), 10) + ":" + strconv.FormatInt(int64(repositoryID), 10)
}

func LogApprovalStageChanged(db *gorm.DB, platformID, repositoryID, approvalID, oldStage, newStage int) (*CIRepositoryLog, *CIRepositoryLogApprovalStageChangedExtra, error) {
	log := &CIRepositoryLog{
		Type:      CILogApprovalStageChanged,
		Reference: ReferenceLogApprovalStageChanged(platformID, repositoryID),
	}
	extra := &CIRepositoryLogApprovalStageChangedExtra{
		ApprovalID: approvalID,
		OldStage:   oldStage,
		NewStage:   newStage,
	}
	if err := log.EncodeExtra(extra); err != nil {
		return nil, nil, err
	}
	if err := db.Save(log).Error; err != nil {
		return log, extra, err
	}
	return log, extra, nil
}

func ReferenceLogBuildPackage(stateType, buildID int, commitHash string) string {
	return "build:" + strconv.FormatInt(int64(buildID), 10) + ":" + commitHash
}

func LogBuildPackage(db *gorm.DB, stateType int, buildID int, reason, namespace, environment, tag, commitHash string) (*CIRepositoryLog, *CIRepositoryLogBuildPackageExtra, error) {
	log := &CIRepositoryLog{
		Type: stateType,
		Reference: ReferenceLogBuildPackage(stateType, buildID, commitHash),
	}
	extra := &CIRepositoryLogBuildPackageExtra{
		Namespace: namespace,
		Environment: environment,
		Tag: tag,
		Reason: reason,
		BuildID: buildID,
		CommitHash: commitHash,
	}
	if err := log.EncodeExtra(extra); err != nil {
		return nil, nil, err
	}
	if err := db.Save(log).Error; err != nil {
		return log, extra, err
	}
	return log, extra, nil
}

func GetApprovalStageChangedLogs(db *gorm.DB, platformID, repositoryID, approvalID int) (logs []CIRepositoryLog, err error) {
	if err = db.Where(&CIRepositoryLog{
		Type:      CILogApprovalStageChanged,
		Reference: ReferenceLogApprovalStageChanged(platformID, repositoryID),
	}).Find(&logs).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			logs = make([]CIRepositoryLog, 0)
			return logs, nil
		}
		return nil, err
	}
	return logs, nil
}