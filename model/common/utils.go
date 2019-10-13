package common

import "github.com/jinzhu/gorm"

func PickByColumn(db *gorm.DB, columnName string, elem interface{}, val interface{}) (err error) {
	err = db.Model(elem).Where(columnName+"= ?", val).First(elem).Error
	if gorm.IsRecordNotFoundError(err) {
		err = nil
	}
	return
}
