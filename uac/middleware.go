package uac

import (
	"encoding/base64"
	//"github.com/gin-gonic/gin"
)

type SessionWrapper struct {
	token []byte
}

func NewSessionWrapper(rawToken string) (*SessionWrapper, error) {
	token, err := base64.StdEncoding.DecodeString(rawToken)
	if err != nil {
		return nil, err
	}
	return &SessionWrapper{token: token}, nil
}

//func (s *SessionWrapper) Wrap(f gin.HandlerFunc) gin.HandlerFunc {
//}
