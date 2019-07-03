package model

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
)

type SecretHasher interface {
	HashString(string) (string, error)
	Hash([]byte) (string, error)
}

type HMACHasher struct {
	key []byte
}

func NewHMACHasher(rawKey string) (SecretHasher, error) {
	key, err := base64.StdEncoding.DecodeString(rawKey)
	if err != nil {
		return nil, err
	}
	return &HMACHasher{key: key}, nil
}

func (s *HMACHasher) HashString(str string) (string, error) {
	return s.Hash([]byte(str))
}

func (s *HMACHasher) Hash(data []byte) (string, error) {
	mac := hmac.New(md5.New, s.key)
	if _, err := mac.Write(data); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

type MD5Hasher struct{}

func NewMD5Hasher() SecretHasher {
	return MD5Hasher(struct{}{})
}

func (s MD5Hasher) HashString(str string) (string, error) {
	sum := md5.Sum([]byte(str))
	return base64.StdEncoding.EncodeToString(sum[:]), nil
}

func (s MD5Hasher) Hash(data []byte) (string, error) {
	sum := md5.Sum(data)
	return base64.StdEncoding.EncodeToString(sum[:]), nil
}
