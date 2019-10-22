package sae

import (
	"git.stuhome.com/Sunmxt/wing/model/common"
	"github.com/jinzhu/gorm"
)

func (m *ApplicationDeployment) ByID(db *gorm.DB, id int) error {
	return common.PickByColumn(db, "id", m, id)
}

func (m *Orchestrator) ByID(db *gorm.DB, id int) error {
	return common.PickByColumn(db, "id", m, id)
}

func (m *ApplicationCluster) ByID(db *gorm.DB, id int) error {
	return common.PickByColumn(db, "id", m, id)
}

func (m *Application) ByID(db *gorm.DB, id int) error {
	return common.PickByColumn(db, "id", m, id)
}
