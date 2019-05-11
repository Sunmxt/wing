package uac

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
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

type Role struct {
	ID   int `gorm:"primary_key"`
	Name string
}

func (m *Role) TableName() string {
	return "role"
}

type RoleRecord struct {
	ID           int    `gorm:"primary_key"`
	ResourceName string `gorm:"column:resource_name"`
	Verbs        int64  `gorm:"column:verbs"`
	Role         Role   `gorm:"foreignkey:ID;associate_foreignkey:RoleID"`
	RoleID       int
}

func (m *RoleRecord) TableName() string {
	return "role_record"
}

type RoleBinding struct {
	ID int `gorm:"primary_key"`

	Account Account `gorm:"foreignkey:ID;associate_foreignkey:AccountID"`
	Role    Role    `gorm:"foreignkey:ID;associate_foreignkey:RoleID"`

	RoleID    int
	AccountID int
}

func (m *RoleBinding) TableName() string {
	return "role_binding"
}

func initRBACRoot(db *gorm.DB, initCredentials string) error {
	adminAccount, adminRole := &Account{}, &Role{}

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
	err = db.Where(&Role{Name: "admin"}).First(adminRole).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			adminRole.Name = "admin"
			err = db.Save(adminRole).Error
		}
	}
	if err != nil {
		log.Error("[migration-init] Failed to initialize admin role.")
		return err
	}

	// Admin role records.
	adminRecord := &RoleRecord{}
	if err = db.Where(&RoleRecord{RoleID: adminRole.ID}).First(adminRecord).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.Error("[migration-init] Failed to initialize admin role.")
			return err
		}
	}
	adminRecord.ResourceName = "*"
	adminRecord.Verbs = VerbAll
	adminRecord.RoleID = adminRole.ID
	err = db.Save(adminRecord).Error
	if err != nil {
		log.Error("[migration-init] Failed to alter admin role.")
		return err
	}

	// Binding.
	binding := &RoleBinding{RoleID: adminRole.ID, AccountID: adminAccount.ID}
	err = db.Where(binding).First(binding).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = db.Save(binding).Error
		}
	}
	if err != nil {
		log.Error("[migration-init] Failed to initialize admin binding.")
		return err
	}

	return nil
}

func Migrate(db *gorm.DB, initCredentials string) error {
	db.AutoMigrate(&Account{})
	db.AutoMigrate(&Role{})
	db.AutoMigrate(&RoleRecord{}).AddForeignKey("role_id", "role(id)", "RESTRICT", "RESTRICT")
	db.AutoMigrate(&RoleBinding{}).AddForeignKey("account_id", "account(id)", "RESTRICT", "RESTRICT").AddForeignKey("role_id", "role(id)", "RESTRICT", "RESTRICT")

	return initRBACRoot(db, initCredentials)
}
