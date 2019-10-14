package common

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
)

func PickByColumn(db *gorm.DB, columnName string, elem interface{}, val interface{}) (err error) {
	err = db.Model(elem).Where(columnName+"= ?", val).First(elem).Error
	if gorm.IsRecordNotFoundError(err) {
		err = nil
	}
	return
}

func DecodeExtra(raw string, v interface{}) error {
	return json.Unmarshal([]byte(raw), v)
}

func EncodeExtra(raw *string, v interface{}) error {
	bin, err := json.Marshal(v)
	if err != nil {
		return err
	}
	*raw = string(bin)
	return nil
}
