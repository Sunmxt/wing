package common

import (
	"time"
)

type Basic struct {
	ID         int       `gorm:"primary_key;not null;auto_increment;unique"`
	CreateTime time.Time `gorm:"type:datetime;not null"`
	ModifyTime time.Time `gorm:"type:datetime;not null"`
}
