package model

import (
	"errors"
	"git.stuhome.com/Sunmxt/wing/cmd/config"
	"git.stuhome.com/Sunmxt/wing/model/account"
	"git.stuhome.com/Sunmxt/wing/model/scm"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"strings"
)

func Migrate(db *gorm.DB, cfg *config.WingConfiguration) (err error) {
	db.AutoMigrate(&account.Account{})
	db.AutoMigrate(&account.RoleModel{})
	db.AutoMigrate(&AppSpec{})
	db.AutoMigrate(&account.RoleRecord{}).AddForeignKey("role_id", "role(id)", "RESTRICT", "RESTRICT")
	db.AutoMigrate(&account.RoleBinding{}).AddForeignKey("account_id", "account(id)", "RESTRICT", "RESTRICT").AddForeignKey("role_id", "role(id)", "RESTRICT", "RESTRICT")
	db.AutoMigrate(&Application{}).AddForeignKey("owner_id", "account(id)", "RESTRICT", "RESTRICT").AddForeignKey("spec_id", "application_spec(id)", "RESTRICT", "RESTRICT")
	db.AutoMigrate(&Deployment{}).AddForeignKey("spec_id", "application_spec(id)", "RESTRICT", "RESTRICT").AddForeignKey("app_id", "application(id)", "RESTRICT", "RESTRICT")
	db.AutoMigrate(&Settings{})

	if err = scm.Migrate(db, cfg); err != nil {
		return err
	}

	return initRBACRoot(db, cfg)
}

func initRBACRoot(db *gorm.DB, cfg *config.WingConfiguration) (err error) {
	adminAccount := &account.Account{}

	hasher := account.NewMD5Hasher()
	// Admin account.
	if err = db.Where(&account.Account{Name: "admin"}).First(adminAccount).Error; err != nil {
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
	role := account.Role("admin")
	if errs := role.Grant("*", account.VerbAll).Update(db); errs != nil && len(errs) > 0 {
		msgs := make([]string, len(errs))
		for _, err := range errs {
			msgs = append(msgs, err.Error())
		}
		err = errors.New("[migration-init] Failed to create admin role: " + strings.Join(msgs, "\n    "))
		log.Error(err.Error())
		return err
	}
	if err = account.NewRBACContext("admin").Grant(role).Update(db); err != nil {
		log.Error("[migration-init] Failed to grant admin role to admin: " + err.Error())
	}

	return nil
}
