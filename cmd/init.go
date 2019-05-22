package cmd

import (
	"git.stuhome.com/Sunmxt/wing/uac"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (c *Wing) Init() {
	c.LogConfig()

	log.Info("[init] Initializing...")

	db, err := gorm.Open(c.Config.DB.SQLEngine, c.Config.DB.SQLDsn)
	if err != nil {
		log.Error("[migration] Cannot open database: " + err.Error())
		return
	}

	defer func() {
		if err != nil {
			log.Error("[migration] Migration failure:" + err.Error())
		}
		db.Close()
	}()

	log.Info("[migration] Migrate UAC data models.")
	if err = uac.Migrate(db, c.Config); err != nil {
		return
	}
}
