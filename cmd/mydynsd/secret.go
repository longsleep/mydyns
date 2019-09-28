/*
Mydyns - run your own dynamic DNS zone
Copyright (C) 2015  Simon Eisenmann

This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; either version 2 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

package main

import (
	"github.com/gorilla/securecookie"
	"io/ioutil"
	"log"
)

type SecretFile struct {
	generator *securecookie.SecureCookie
}

func NewSecretFile(fn string) (*SecretFile, error) {
	secret, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	if len(secret) != 32 && len(secret) != 64 {
		log.Printf("Warning: secret size should be 32 or 64 bytes but is %d bytes\n", len(secret))
	}
	generator := securecookie.New(secret, nil)
	generator.MaxAge(0)
	return &SecretFile{
		generator: generator,
	}, nil
}

func (s *SecretFile) Encode(name string, value interface{}) (string, error) {
	return s.generator.Encode(name, value)
}

func (s *SecretFile) Decode(name, value string, dst interface{}) error {
	return s.generator.Decode(name, value, dst)
}
