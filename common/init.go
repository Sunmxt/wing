package common

import (
	log "github.com/sirupsen/logrus"
	"regexp"
)

func init() {
	registerI18NMessage()

	var err error
	if ReMail, err = regexp.Compile("^[\\w.\\-]+@(?:[a-z0-9]+(?:-[a-z0-9]+)*\\.)+[a-z]{2,3}$"); err != nil {
		log.Panic(err.Error())
	}
}
