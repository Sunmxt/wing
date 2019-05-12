package uac

import (
	"errors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
	"strings"
)

const (
	ACTIVE  = 0
	BLOCKED = 1

	VerbGet    = 1
	VerbDelete = 1 << 1
	VerbUpdate = 1 << 2
	VerbCreate = 1 << 3
	VerbAll    = VerbGet | VerbDelete | VerbUpdate | VerbCreate
)

type Account struct {
	ID          int    `gorm:"primary_key"`
	Name        string `gorm:"column:name;type:varchar(16)"`
	Credentials string `gorm:"column:credentials;type:varchar(64)"`
	State       int    `gorm:"column:state"`
	Extra       string `gorm:"column:extra;type:longtext"`
}

func (m *Account) TableName() string {
	return "account"
}

type RoleModel struct {
	ID   int `gorm:"primary_key"`
	Name string
}

func (m *RoleModel) TableName() string {
	return "role"
}

type RoleRecord struct {
	ID           int       `gorm:"primary_key"`
	ResourceName string    `gorm:"column:resource_name"`
	Verbs        int64     `gorm:"column:verbs"`
	Role         RoleModel `gorm:"foreignkey:ID;associate_foreignkey:RoleID"`
	RoleID       int
}

func (m *RoleRecord) TableName() string {
	return "role_record"
}

type RoleBinding struct {
	ID int `gorm:"primary_key"`

	Account Account   `gorm:"foreignkey:ID;associate_foreignkey:AccountID"`
	Role    RoleModel `gorm:"foreignkey:ID;associate_foreignkey:RoleID"`

	RoleID    int
	AccountID int
}

func (m *RoleBinding) TableName() string {
	return "role_binding"
}

func initRBACRoot(db *gorm.DB, initCredentials string) error {
	adminAccount := &Account{}

	// Admin account.
	err := db.Where(&Account{Name: "admin"}).First(adminAccount).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			adminAccount.Name, adminAccount.Credentials = "admin", initCredentials
			adminAccount.State, adminAccount.Extra = 0, ""
			err = db.Save(adminAccount).Error
		}
	}
	if err != nil {
		log.Error("[migration-init] Cannot initialize admin account.")
		return err
	}
	log.Infof("[migration-init] Admin account ID is %v", adminAccount.ID)

	// Admin role.
	role := Role("admin")
	if errs := role.Grant("*", VerbAll).Update(db); errs != nil && len(errs) > 0 {
		msgs := make([]string, len(errs))
		for _, err := range errs {
			msgs = append(msgs, err.Error())
		}
		err = errors.New("[migration-init] Failed to create admin role: " + strings.Join(msgs, "\n    "))
		log.Error(err.Error())
		return err
	}
	if err = NewRBACContext("admin").Grant(role).Update(db); err != nil {
		log.Error("[migration-init] Failed to grant admin role to admin: " + err.Error())
	}

	return nil
}

func Migrate(db *gorm.DB, initCredentials string) error {
	db.AutoMigrate(&Account{})
	db.AutoMigrate(&RoleModel{})
	db.AutoMigrate(&RoleRecord{}).AddForeignKey("role_id", "role(id)", "RESTRICT", "RESTRICT")
	db.AutoMigrate(&RoleBinding{}).AddForeignKey("account_id", "account(id)", "RESTRICT", "RESTRICT").AddForeignKey("role_id", "role(id)", "RESTRICT", "RESTRICT")

	return initRBACRoot(db, initCredentials)
}
