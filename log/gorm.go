package log

import (
	log "github.com/sirupsen/logrus"
)

type GormLogger *log.Entry

func NewGormLogger(entry *log.Entry) GormLogger {
	return GormLogger(entry)
}

//func (l GormLogger) Print(values interface{}) {
//}
