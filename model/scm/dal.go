package scm

import (
	"github.com/jinzhu/gorm"
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
