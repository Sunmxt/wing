package model

import (
	"git.stuhome.com/Sunmxt/wing/model/common"
)

type Settings struct {
	common.Basic
	Key   string `gorm:"type:varchar(64);not null;unique"`
	Value string `gorm:"type:longtext;not null"`
}

func (s Settings) TableName() string {
	return "settings"
}
