package uac

import (
	"github.com/jinzhu/gorm"
)

type Account struct {
	gorm.Model
	Name  string
	Extra string
}

type Role struct {
	gorm.Model
	Name string
}

type RoleRecord struct {
	gorm.Model
	Role
	ResourceName string
	Verbs        int64
}

type RoleBinding struct {
	gorm.Model
	Account
	Role
}
