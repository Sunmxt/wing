package model

type Settings struct {
	Basic
	Key   string `gorm:"type:varchar(64);not null;unique"`
	Value string `gorm:"type:longtext;not null"`
}

func (s Settings) TableName() string {
	return "settings"
}
