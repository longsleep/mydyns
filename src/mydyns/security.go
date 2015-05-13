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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/csv"
	"log"
	"os"
	"strings"
)

type SecurityFile struct {
	security map[string][]byte
}

func NewSecurityFile(fn string) (*SecurityFile, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// security files are essentially csv files.
	reader := csv.NewReader(f)
	reader.Comma = ':'
	reader.Comment = '#'
	reader.TrimLeadingSpace = true

	entries, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	s := &SecurityFile{
		security: make(map[string][]byte),
	}
	for _, entry := range entries {
		s.security[entry[0]] = []byte(strings.Trim(entry[1], " "))
	}

	log.Printf("Loaded %d security entries\n", len(s.security))
	return s, nil

}

func (s *SecurityFile) Check(secret []byte, user string) bool {
	if len(secret) == 0 {
		entry, _ := s.security[user]
		if len(entry) == 0 {
			// Allow empty secret, if user has none.
			return true
		}
	}
	expectedSecret := s.Secret(user)
	return hmac.Equal(secret, expectedSecret)
}

func (s *SecurityFile) Secret(user string) []byte {
	entry, _ := s.security[user]
	if len(entry) == 0 {
		return entry
	}
	mac := hmac.New(sha256.New, entry)
	mac.Write([]byte(user))
	return mac.Sum(nil)
}
