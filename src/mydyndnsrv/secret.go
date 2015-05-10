package main

import (
	"github.com/gorilla/securecookie"
	"io/ioutil"
)

type SecretFile struct {
	generator *securecookie.SecureCookie
}

func NewSecretFile(fn string) (*SecretFile, error) {
	secret, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	return &SecretFile{
		generator: securecookie.New(secret, nil),
	}, nil
}

func (s *SecretFile) Encode(name string, value interface{}) (string, error) {
	return s.generator.Encode(name, value)
}

func (s *SecretFile) Decode(name, value string, dst interface{}) error {
	return s.generator.Decode(name, value, dst)
}
