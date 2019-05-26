package model

import (
	"errors"
	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type Basic struct {
	ID         int       `gorm:"primary_key;not null;auto_increment;unique"`
	CreateTime time.Time `gorm:"type:datetime;not null"`
	ModifyTime time.Time `gorm:"type:datetime;not null"`
}

func Migrate(db *gorm.DB, cfg *config.WingConfiguration) error {
	db.AutoMigrate(&Account{})
	db.AutoMigrate(&RoleModel{})
	db.AutoMigrate(&AppSpec{})
	db.AutoMigrate(&RoleRecord{}).AddForeignKey("role_id", "role(id)", "RESTRICT", "RESTRICT")
	db.AutoMigrate(&RoleBinding{}).AddForeignKey("account_id", "account(id)", "RESTRICT", "RESTRICT").AddForeignKey("role_id", "role(id)", "RESTRICT", "RESTRICT")
	db.AutoMigrate(&Application{}).AddForeignKey("owner_id", "account(id)", "RESTRICT", "RESTRICT").AddForeignKey("spec_id", "application_spec(id)", "RESTRICT", "RESTRICT")
	db.AutoMigrate(&Deployment{}).AddForeignKey("spec_id", "application_spec(id)", "RESTRICT", "RESTRICT").AddForeignKey("app_id", "application(id)", "RESTRICT", "RESTRICT")
	return initRBACRoot(db, cfg)
}

func initRBACRoot(db *gorm.DB, cfg *config.WingConfiguration) error {
	adminAccount := &Account{}

	hasher, err := NewSecretHasher(cfg.SessionToken)
	if err != nil {
		log.Error("[migration-init] Cannot create SecretHasher: " + err.Error())
		return err
	}

	// Admin account.
	if err = db.Where(&Account{Name: "admin"}).First(adminAccount).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			adminAccount.Credentials, err = hasher.HashString(cfg.InitialAdminCredentials)
			if err != nil {
				log.Error("[migration-init] Failed to hash credential: " + err.Error())
				return err
			}
			adminAccount.Name, adminAccount.State, adminAccount.Extra = "admin", 0, ""
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
