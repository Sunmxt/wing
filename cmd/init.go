package cmd

import (
	"git.stuhome.com/Sunmxt/wing/model"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (c *Wing) Init() {
	c.LogConfig()

	log.Info("[init] Initializing...")

	db, err := gorm.Open(c.Runtime.Config.DB.SQLEngine, c.Runtime.Config.DB.SQLDsn)
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

	log.Info("[migration] Apply migration.")
	if err = model.Migrate(db, c.Runtime.Config); err != nil {
		return
	}
}
