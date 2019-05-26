package model

const (
	Waiting     = 0
	Executed    = 1
	Finished    = 2
	Terminating = 3
	Terminated  = 4
)

type Deployment struct {
	Basic
	RelatedRevison int
	Spec           *AppSpec     `gorm:"foreignkey:SpecID;not null"`
	App            *Application `gorm:"foreignkey:AppID;not null"`
	Pods           string       `gorm:"not null"`
	State          int

	SpecID int
	AppID  int
}
