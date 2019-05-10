package uac

import (
	"github.com/jinzhu/gorm"
)

const (
	ACTIVE  = 0
	BLOCKED = 1
)

type Account struct {
	gorm.Model
	Name        string `gorm:"column:name,type:varchar(16)"`
	Credentials string `gorm:"column:credentials,type:varchar(64)"`
	State       int    `gorm:"column:state"`
	Extra       string `gorm:"column:extra,type:longtext"`
}

type Role struct {
	gorm.Model
	Name string
}

type RoleRecord struct {
	gorm.Model
	Role
	ResourceName string `gorm:"column:resource_name"`
	Verbs        int64  `gorm:"column:verbs"`
}

type RoleBinding struct {
	gorm.Model
	Account
	Role
}
