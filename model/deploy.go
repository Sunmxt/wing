package model

import (
	"git.stuhome.com/Sunmxt/wing/model/common"
)

const (
	Waiting     = 0
	Executed    = 1
	Finished    = 2
	Terminating = 3
	Terminated  = 4
)

type Deployment struct {
	common.Basic
	RelatedRevison int
	Spec           *AppSpec     `gorm:"foreignkey:SpecID;not null"`
	App            *Application `gorm:"foreignkey:AppID;not null"`
	Pods           string       `gorm:"type:longtext;not null"`
	State          int

	SpecID int
	AppID  int
}
