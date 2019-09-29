package scm

import (
	"github.com/jinzhu/gorm"
	"strconv"
)

func pickByID(db *gorm.DB, elem interface{}, id int) (err error) {
	err = db.Model(elem).Where("id = ?", id).First(elem).Error
	if gorm.IsRecordNotFoundError(err) {
		err = nil
	}
	return
}

func (p *SCMPlatform) ByID(db *gorm.DB, platformID int) error {
	return pickByID(db, p, platformID)
}

func (a *CIRepositoryApproval) ByID(db *gorm.DB, approvalID int) error {
	return pickByID(db, a, approvalID)
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
