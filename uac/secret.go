package uac

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
)

type SecretHasher struct {
	key []byte
}

func NewSecretHasher(rawKey string) (*SecretHasher, error) {
	key, err := base64.StdEncoding.DecodeString(rawKey)
	if err != nil {
		return nil, err
	}
	return &SecretHasher{key: key}, nil
}

func (s *SecretHasher) HashString(str string) (string, error) {
	return s.Hash([]byte(str))
}

func (s *SecretHasher) Hash(data []byte) (string, error) {
	mac := hmac.New(md5.New, s.key)
	if _, err := mac.Write(data); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}
